package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/broadcast"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/rounds"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/primitives/utils"
	"sync"
)

// singleCmd is the single-use subcommand that allows for sening and responding
// to single-use messages.
var broadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "Send broadcast messages",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmixParams, e2eParams := initParams()
		authCbs := makeAuthCallbacks(
			viper.GetBool(unsafeChannelCreationFlag), e2eParams)
		client := initE2e(cmixParams, e2eParams, authCbs)

		// Write user contact to file
		user := client.GetReceptionIdentity()
		jww.INFO.Printf("User: %s", user.ID)
		writeContact(user.GetContact())

		err := client.StartNetworkFollower(5 * time.Second)
		if err != nil {
			jww.FATAL.Panicf("Failed to start network follower: %+v", err)
		}

		// Wait until connected or crash on timeout
		connected := make(chan bool, 10)
		client.GetCmix().AddHealthCallback(
			func(isConnected bool) {
				connected <- isConnected
			})
		waitUntilConnected(connected)
		/* Set up underlying crypto broadcast.Channel */
		var channel *crypto.Channel
		var pk *rsa.PrivateKey
		keyPath := viper.GetString(broadcastKeyPathFlag)
		path, err := utils.ExpandPath(viper.GetString(broadcastChanPathFlag))
		if utils.Exists(path) {
			// Load symmetric from path
			cBytes, err := utils.ReadFile(path)
			if err != nil {
				jww.FATAL.Panicf("Failed to read channel from file at %s: %+v", path, err)
			}
			channel, err = crypto.UnmarshalChannel(cBytes)
			if err != nil {
				jww.FATAL.Panicf("Failed to unmarshal channel data %+v: %+v", cBytes, err)
			}
		} else {
			// Load in broadcast channel info
			name := viper.GetString(broadcastNameFlag)
			desc := viper.GetString(broadcastDescriptionFlag)
			if name == "" {
				jww.FATAL.Panicf("Name cannot be empty")
			} else if desc == "" {
				jww.FATAL.Panicf("description cannot be empty")
			}

			var cryptChannel *crypto.Channel
			if viper.GetBool(broadcastNewFlag) {
				// Create a new broadcast channel
				cryptChannel, pk, err = crypto.NewChannel(name, desc, client.GetRng().GetStream())
				if err != nil {
					jww.FATAL.Panicf("Failed to create new channel: %+v", err)
				}

				if keyPath != "" {
					err = utils.WriteFile(keyPath, rsa.CreatePrivateKeyPem(pk), os.ModePerm, os.ModeDir)
					if err != nil {
						jww.ERROR.Printf("Failed to write private key to path %s: %+v", path, err)
					}
				} else {
					fmt.Printf("Private key generated for channel: %+v", rsa.CreatePrivateKeyPem(pk))
				}
				fmt.Printf("New broadcast channel generated")
			} else {
				// Read rest of info from config & build object manually
				pubKeyBytes := []byte(viper.GetString(broadcastRsaPubFlag))
				pubKey, err := rsa.LoadPublicKeyFromPem(pubKeyBytes)
				if err != nil {
					jww.FATAL.Panicf("Failed to load public key at path: %+v", err)
				}
				salt := []byte(viper.GetString(broadcastSaltFlag))

				rid, err := crypto.NewChannelID(name, desc, salt, pubKeyBytes)
				if err != nil {
					jww.FATAL.Panicf("Failed to generate channel ID: %+v", err)
				}

				cryptChannel = &crypto.Channel{
					ReceptionID: rid,
					Name:        name,
					Description: desc,
					Salt:        salt,
					RsaPubKey:   pubKey,
				}
			}

			// Save channel to disk
			cBytes, err := cryptChannel.Marshal()
			if err != nil {
				jww.ERROR.Printf("Failed to marshal channel to bytes: %+v", err)
			}
			// Write to file if there
			if path != "" {
				err = utils.WriteFile(path, cBytes, os.ModePerm, os.ModeDir)
				if err != nil {
					jww.ERROR.Printf("Failed to write channel to file %s: %+v", path, err)
				}
			} else {
				fmt.Printf("Channel marshalled: %+v", cBytes)
			}
		}

		// Load key if needed
		if pk == nil && keyPath != "" {
			jww.DEBUG.Printf("Attempting to load private key at %s", keyPath)
			if ep, err := utils.ExpandPath(keyPath); err == nil {
				keyBytes, err := utils.ReadFile(ep)
				if err != nil {
					jww.ERROR.Printf("Failed to read private key from %s: %+v", ep, err)
				}
				pk, err = rsa.LoadPrivateKeyFromPem(keyBytes)
				if err != nil {
					jww.ERROR.Printf("Failed to load private key %+v: %+v", keyBytes, err)
				}
			} else {
				jww.ERROR.Printf("Failed to expand private key path: %+v", err)
			}
		}

		/* Broadcast client setup */

		// Select broadcast method
		symmetric := viper.GetString(broadcastSymmetricFlag)
		asymmetric := viper.GetString(broadcastAsymmetricFlag)

		// Connect to broadcast channel
		bcl, err := broadcast.NewBroadcastChannel(*channel, client.GetCmix(), client.GetRng())

		// Create & register symmetric receiver callback
		receiveChan := make(chan []byte, 100)
		scb := func(payload []byte,
			receptionID receptionID.EphemeralIdentity, round rounds.Round) {
			jww.INFO.Printf("Received symmetric message from %s over round %d", receptionID, round.ID)
			receiveChan <- payload
		}
		err = bcl.RegisterListener(scb, broadcast.Symmetric)
		if err != nil {
			jww.FATAL.Panicf("Failed to register asymmetric listener: %+v", err)
		}

		// Create & register asymmetric receiver callback
		asymmetricReceiveChan := make(chan []byte, 100)
		acb := func(payload []byte,
			receptionID receptionID.EphemeralIdentity, round rounds.Round) {
			jww.INFO.Printf("Received asymmetric message from %s over round %d", receptionID, round.ID)
			asymmetricReceiveChan <- payload
		}
		err = bcl.RegisterListener(acb, broadcast.Asymmetric)
		if err != nil {
			jww.FATAL.Panicf("Failed to register asymmetric listener: %+v", err)
		}

		jww.INFO.Printf("Broadcast listeners registered...")

		/* Broadcast messages to the channel */
		if symmetric != "" || asymmetric != "" {
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				jww.INFO.Printf("Attempting to send broadcasts...")

				sendDelay := time.Duration(viper.GetUint(sendDelayFlag))
				maxRetries := 10
				retries := 0
				for {
					// Wait for sendDelay before sending (to allow connection to establish)
					if maxRetries == retries {
						jww.FATAL.Panicf("Max retries reached")
					}
					time.Sleep(sendDelay*time.Millisecond*time.Duration(retries) + 1)

					/* Send symmetric broadcast */
					if symmetric != "" {
						// Create properly sized broadcast message
						broadcastMessage, err := broadcast.NewSizedBroadcast(bcl.MaxPayloadSize(), []byte(symmetric))
						if err != nil {
							jww.FATAL.Panicf("Failed to create sized broadcast: %+v", err)
						}
						rid, eid, err := bcl.Broadcast(broadcastMessage, cmix.GetDefaultCMIXParams())
						if err != nil {
							jww.ERROR.Printf("Failed to send symmetric broadcast message: %+v", err)
							retries++
							continue
						}
						fmt.Printf("Sent symmetric broadcast message: %s", symmetric)
						jww.INFO.Printf("Sent symmetric broadcast message to %s over round %d", eid, rid)
					}

					/* Send asymmetric broadcast */
					if asymmetric != "" {
						// Create properly sized broadcast message
						broadcastMessage, err := broadcast.NewSizedBroadcast(bcl.MaxAsymmetricPayloadSize(), []byte(asymmetric))
						if err != nil {
							jww.FATAL.Panicf("Failed to create sized broadcast: %+v", err)
						}
						if pk == nil {
							jww.FATAL.Panicf("CANNOT SEND ASYMMETRIC BROADCAST WITHOUT PRIVATE KEY")
						}
						rid, eid, err := bcl.BroadcastAsymmetric(pk, broadcastMessage, cmix.GetDefaultCMIXParams())
						if err != nil {
							jww.ERROR.Printf("Failed to send asymmetric broadcast message: %+v", err)
							retries++
							continue
						}
						fmt.Printf("Sent asymmetric broadcast message: %s", asymmetric)
						jww.INFO.Printf("Sent asymmetric broadcast message to %s over round %d", eid, rid)
					}

					wg.Done()
					break
				}
			}()

			wg.Wait()
		}
		/* Create properly sized broadcast message */

		/* Receive broadcast messages over the channel */
		jww.INFO.Printf("Waiting for message reception...")
		waitSecs := viper.GetUint(waitTimeoutFlag)
		expectedCnt := viper.GetUint(receiveCountFlag)
		waitTimeout := time.Duration(waitSecs) * time.Second
		receivedCount := uint(0)
		done := false
		for !done && expectedCnt != 0 {
			timeout := time.NewTimer(waitTimeout)
			select {
			case receivedPayload := <-asymmetricReceiveChan:
				receivedCount++
				receivedBroadcast, err := broadcast.DecodeSizedBroadcast(receivedPayload)
				if err != nil {
					jww.ERROR.Printf("Failed to decode sized broadcast: %+v", err)
					continue
				}
				fmt.Printf("Asymmetric broadcast message received: %s\n", string(receivedBroadcast))
				if receivedCount == expectedCnt {
					done = true
				}
			case receivedPayload := <-receiveChan:
				receivedCount++
				receivedBroadcast, err := broadcast.DecodeSizedBroadcast(receivedPayload)
				if err != nil {
					jww.ERROR.Printf("Failed to decode sized broadcast: %+v", err)
					continue
				}
				fmt.Printf("Symmetric broadcast message received: %s\n", string(receivedBroadcast))
				if receivedCount == expectedCnt {
					done = true
				}
			case <-timeout.C:
				fmt.Println("Timed out")
				jww.ERROR.Printf("Timed out on message reception after %s!", waitTimeout)
				done = true
			}
		}

		jww.INFO.Printf("Received %d/%d Messages!", receivedCount, expectedCnt)
		bcl.Stop()
		err = client.StopNetworkFollower()
		if err != nil {
			jww.WARN.Printf("Failed to cleanly close threads: %+v\n", err)
		}
	},
}

func init() {
	// Single-use subcommand options
	broadcastCmd.Flags().StringP(broadcastNameFlag, "", "",
		"Symmetric channel name")
	bindFlagHelper(broadcastNameFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastRsaPubFlag, "", "",
		"Broadcast channel rsa pub key")
	bindFlagHelper(broadcastRsaPubFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastSaltFlag, "", "",
		"Broadcast channel salt")
	bindFlagHelper(broadcastSaltFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastDescriptionFlag, "", "",
		"Broadcast channel description")
	bindFlagHelper(broadcastDescriptionFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastChanPathFlag, "", "",
		"Broadcast channel output path")
	bindFlagHelper(broadcastChanPathFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastKeyPathFlag, "", "",
		"Broadcast channel private key output path")
	bindFlagHelper(broadcastKeyPathFlag, broadcastCmd)

	broadcastCmd.Flags().BoolP(broadcastNewFlag, "", false,
		"Create new broadcast channel")
	bindFlagHelper(broadcastNewFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastSymmetricFlag, "", "",
		"Send symmetric broadcast message")
	_ = viper.BindPFlag("symmetric", broadcastCmd.Flags().Lookup("symmetric"))
	bindFlagHelper(broadcastSymmetricFlag, broadcastCmd)

	broadcastCmd.Flags().StringP(broadcastAsymmetricFlag, "", "",
		"Send asymmetric broadcast message (must be used with keyPath)")
	_ = viper.BindPFlag("asymmetric", broadcastCmd.Flags().Lookup("asymmetric"))
	bindFlagHelper(broadcastAsymmetricFlag, broadcastCmd)

	rootCmd.AddCommand(broadcastCmd)
}
