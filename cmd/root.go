////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Package cmd initializes the CLI and config parsers as well as the logger.
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/privategrity/client/api"
	"gitlab.com/privategrity/client/bindings"
	"gitlab.com/privategrity/client/globals"
	"gitlab.com/privategrity/client/user"
	"gitlab.com/privategrity/crypto/cyclic"
	"os"
	"time"
	"gitlab.com/privategrity/client/parse"
	"github.com/golang/protobuf/proto"
	"sync/atomic"
	"gitlab.com/privategrity/client/listener"
)

var verbose bool
var userId uint64
var destinationUserId uint64
var gwAddr string
var message string
var numNodes uint
var sessionFile string
var noRatchet bool
var dummyFrequency float64
var noBlockingTransmission bool
var rateLimiting uint32
var showVer bool

// Execute adds all child commands to the root command and sets flags
// appropriately.  This is called by main.main(). It only needs to
// happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		jww.ERROR.Println(err)
		os.Exit(1)
	}
}

func sessionInitialization() {
	if noBlockingTransmission {
		if !noRatchet {
			fmt.Printf("Cannot disable Blocking Transmission with" +
				" Ratcheting turned on\n")
		}
		api.DisableBlockingTransmission()
	}

	bindings.SetRateLimiting(int(rateLimiting))

	// Disable ratcheting if the flag is set
	if noRatchet {
		bindings.DisableRatchet()
	}

	var err error
	register := false

	//If no session file is passed initialize with RAM Storage
	if sessionFile == "" {
		err = bindings.InitClient(&globals.RamStorage{}, "", nil)
		if err != nil {
			fmt.Printf("Could Not Initialize Ram Storage: %s\n",
				err.Error())
			return
		}
		register = true
	} else {

		//If a session file is passed, check if it's valid
		_, err1 := os.Stat(sessionFile)

		if err1 != nil {
			//If the file does not exist, register a new user
			if os.IsNotExist(err1) {
				register = true
			} else {
				//Fail if any other error is received
				fmt.Printf("Error with file path: %s\n", err1.Error())
			}
		}

		//Initialize client with OS Storage
		err = bindings.InitClient(&globals.DefaultStorage{}, sessionFile, nil)

		if err != nil {
			fmt.Printf("Could Not Initialize OS Storage: %s\n", err.Error())
			return
		}
	}

	// Handle parsing gateway addresses from the config file
	gateways := viper.GetStringSlice("gateways")
	if gwAddr == "" {
		// If gwAddr was not passed via command line, check config file
		if len(gateways) < 1 {
			// No gateways in config file or passed via command line
			fmt.Printf("Error: No gateway specified! Add to" +
				" configuration file or pass via command line using -g!")
			return
		} else {
			// List of gateways found in config file, select one to use
			// TODO: For now, just use the first one?
			gwAddr = gateways[0]
		}
	}

	//Register a new user if requested
	if register {
		_, err := bindings.Register(
			cyclic.NewIntFromBytes(user.UserHash(user.ID(userId))).
				TextVerbose(32, 0), gwAddr, int(numNodes))
		if err != nil {
			fmt.Printf("Could Not Register User: %s\n", err.Error())
			return
		}
	}

	//log the user in
	_, err = bindings.Login(
		cyclic.NewIntFromUInt(userId).LeftpadBytes(8), gwAddr)

	if err != nil {
		fmt.Printf("Could Not Log In\n")
		return
	}
}

type FallbackListener struct {
	messagesReceived int64
}

func (l *FallbackListener) Hear(message *parse.Message) {
	sender, ok := user.Users.GetUser(message.Sender)
	var senderNick string
	if !ok {
		jww.ERROR.Printf("Couldn't get sender %v", message.Sender)
	} else {
		senderNick = sender.Nick
	}
	fmt.Printf("Message of type %v from %v, %v received with fallback: %s\n",
		message.BodyType, message.Sender, senderNick, string(message.Body))
}

type TextListener struct {
	messagesReceived int64
}

func (l *TextListener) Hear(message *parse.Message) {
	result := parse.TextMessage{}
	proto.Unmarshal(message.Body, &result)

	sender, ok := user.Users.GetUser(message.Sender)
	var senderNick string
	if !ok {
		jww.ERROR.Printf("Couldn't get sender %v", message.Sender)
	} else {
		senderNick = sender.Nick
	}
	fmt.Printf("Message from %v, %v Received: %s\n", message.Sender,
		senderNick, result.Message)

	atomic.AddInt64(&l.messagesReceived, 1)
}

type ChannelbotListener struct {
	messagesReceived int64
}

func (l *ChannelbotListener) Hear(message *parse.Message) {
	result := parse.ChannelMessage{}
	proto.Unmarshal(message.Body, &result)

	sender, ok := user.Users.GetUser(message.Sender)
	var senderNick string
	if !ok {
		jww.ERROR.Printf("Couldn't get sender %v", message.Sender)
	} else {
		senderNick = sender.Nick
	}

	speakerID := user.NewIDFromBytes(result.SpeakerID)

	fmt.Printf("Message from channel %v, %v: ",
		message.Sender, senderNick)
	typedBody, _ := parse.Parse(result.Message)
	listener.Listeners.Speak(&parse.Message{
		TypedBody: *typedBody,
		Sender:    speakerID,
		Receiver:  0,
	})
	atomic.AddInt64(&l.messagesReceived, 1)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Runs a client for cMix anonymous communication platform",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Main client run function

		if showVer {
			printVersion()
			return
		} else {
			cmd.MarkPersistentFlagRequired("userid")
			cmd.MarkPersistentFlagRequired("numnodes")
		}

		var dummyPeriod time.Duration
		var timer *time.Timer

		// Do calculation for dummy messages if the flag is set
		if dummyFrequency != 0 {
			dummyPeriod = time.Nanosecond *
				(time.Duration(float64(1000000000) * (float64(1.0) / dummyFrequency)))
		}

		sessionInitialization()

		// Loop until we don't get a message, draining the queue of messages on the
		// gateway buffer
		gotMsg := false
		for i := 0; i < 5; i++ {
			time.Sleep(1000 * time.Millisecond) // wait 1s in between
			msg, _ := bindings.TryReceive()
			if msg.GetPayload() == "" && gotMsg {
				break
			}
			if msg.GetPayload() != "" {
				i = 0 // Make sure to loop until all messages exhausted
				gotMsg = true
			}
		}

		// Only send a message if we have a message to send (except dummy messages)
		if message != "" {
			// Get the recipient's nick
			recipientNick := ""
			u, ok := user.Users.GetUser(user.ID(destinationUserId))
			if ok {
				recipientNick = u.Nick
			}
			wireOut := bindings.FormatTextMessage(message)

			fmt.Printf("Sending Message to %d, %v: %s\n", destinationUserId,
				recipientNick, message)

			//Send the message
			bindings.Send(api.APIMessage{
				SenderID:    user.ID(userId),
				Payload:     string(wireOut),
				RecipientID: user.ID(destinationUserId),
			})
		}

		if dummyFrequency != 0 {
			timer = time.NewTimer(dummyPeriod)
		}

		// Set up the listeners for both of the types the client needs for
		// the integration test
		// Normal text messages
		text := TextListener{}
		api.Listen(user.ID(0), 1, &text, false)
		// Channelbot messages
		channel := ChannelbotListener{}
		api.Listen(user.ID(0), 2, &channel, false)
		// All other messages
		fallback := FallbackListener{}
		api.Listen(user.ID(0), 0, &fallback, true)

		// Loop until we get a message, then print and exit
		for text.messagesReceived == 0 && channel.messagesReceived == 0 {

			_, _ = bindings.TryReceive()

			end := false

			//If dummy messages are enabled, send the next one
			if dummyPeriod != 0 {
				end = false
				<-timer.C

				contact := ""
				u, ok := user.Users.GetUser(user.ID(destinationUserId))
				if ok {
					contact = u.Nick
				}
				fmt.Printf("Sending Message to %d, %v: %s\n", destinationUserId,
					contact, message)

				message := api.APIMessage{
					SenderID: user.ID(userId),
					Payload:  message, RecipientID: user.ID(
						destinationUserId)}
				bindings.Send(message)

				timer = time.NewTimer(dummyPeriod)
				//otherwise just wait to check for new messages
			} else {
				time.Sleep(200 * time.Millisecond)
			}

			if end {
				break
			}

		}

		//Logout
		err := bindings.Logout()

		if err != nil {
			fmt.Printf("Could not logout: %s\n", err.Error())
			return
		}

	},
}

// init is the initialization function for Cobra which defines commands
// and flags.
func init() {
	// NOTE: The point of init() is to be declarative.
	// There is one init in each sub command. Do not put variable declarations
	// here, and ensure all the Flags are of the *P variety, unless there's a
	// very good reason not to have them as local params to sub command."
	cobra.OnInitialize(initConfig, initLog)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose mode for debugging")
	rootCmd.PersistentFlags().BoolVarP(&noRatchet, "noratchet", "", false,
		"Avoid ratcheting the keys for forward secrecy")

	rootCmd.PersistentFlags().BoolVarP(&noBlockingTransmission, "noBlockingTransmission",
		"", false, "Sets if transmitting messages blocks or not.  "+
			"Defaults to true if unset.")

	rootCmd.PersistentFlags().Uint32VarP(&rateLimiting, "rateLimiting", "",
		1000, "Sets the amount of time, in ms, "+
			"that the client waits between sending messages.  "+
			"set to zero to disable.  "+
			"Automatically disabled if 'blockingTransmission' is false")

	rootCmd.PersistentFlags().Uint64VarP(&userId, "userid", "i", 0,
		"ID to sign in as")
	rootCmd.PersistentFlags().StringVarP(&gwAddr, "gwaddr", "g", "",
		"Gateway address to send messages to")
	// TODO: support this negotiating separate keys with different servers
	rootCmd.PersistentFlags().UintVarP(&numNodes, "numnodes", "n", 1,
		"The number of servers in the network that the client is"+
			" connecting to")

	rootCmd.PersistentFlags().StringVarP(&sessionFile, "sessionfile", "f",
		"", "Passes a file path for loading a session.  "+
			"If the file doesn't exist the code will register the user and"+
			" store it there.  If not passed the session will be stored"+
			" to ram and lost when the cli finishes")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&message, "message", "m", "", "Message to send")
	rootCmd.PersistentFlags().Uint64VarP(&destinationUserId, "destid", "d", 0,
		"ID to send message to")
	rootCmd.Flags().BoolVarP(&showVer, "version", "V", false,
		"Show the server version information.")

	rootCmd.Flags().Float64VarP(&dummyFrequency, "dummyfrequency", "", 0,
		"Frequency of dummy messages in Hz.  If no message is passed, "+
			"will transmit a random message.  Dummies are only sent if this flag is passed")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {}

// initLog initializes logging thresholds and the log path.
func initLog() {
	// If verbose flag set then log more info for debugging
	if verbose || viper.GetBool("verbose") {
		jww.SetLogThreshold(jww.LevelInfo)
		jww.SetStdoutThreshold(jww.LevelInfo)
	} else {
		jww.SetLogThreshold(jww.LevelWarn)
		jww.SetStdoutThreshold(jww.LevelWarn)
	}
	if viper.Get("logPath") != nil {
		// Create log file, overwrites if existing
		logPath := viper.GetString("logPath")
		logFile, err := os.Create(logPath)
		if err != nil {
			jww.WARN.Println("Invalid or missing log path, default path used.")
		} else {
			jww.SetLogOutput(logFile)
		}
	}
}
