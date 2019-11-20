////////////////////////////////////////////////////////////////////////////////
// Copyright © 2019 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package io

import (
	"fmt"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/cmixproto"
	"gitlab.com/elixxir/client/crypto"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/client/parse"
	"gitlab.com/elixxir/client/user"
	"gitlab.com/elixxir/comms/connect"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/crypto/e2e"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/elixxir/primitives/id"
	"gitlab.com/elixxir/primitives/switchboard"
	"strings"
	"time"
)

const reportDuration = 30 * time.Second

var errE2ENotFound = errors.New("E2EKey for matching fingerprint not found, can't process message")

// MessageReceiver is a polling thread for receiving messages
func (rm *ReceptionManager) MessageReceiver(session user.Session, delay time.Duration,
	receptionHost *connect.Host, callback func(error)) {
	// FIXME: It's not clear we should be doing decryption here.
	if session == nil {
		globals.Log.FATAL.Panicf("No user session available")
	}
	pollingMessage := pb.ClientRequest{
		UserID: session.GetCurrentUser().User.Bytes(),
	}
	quit := session.GetQuitChan()

	NumChecks := 0
	NumMessages := 0

	reportTicker := time.NewTicker(reportDuration)

	var encryptedMessages []*format.Message

	globals.Log.DEBUG.Printf("Gateway Polling for Message Reception Begun")

	for {
		// TODO: replace timer with ticker
		timerDelay := time.NewTimer(delay)
		NumChecks++
		select {
		case <-quit:
			globals.Log.DEBUG.Printf("Stopped message receiver\n")
			return
		case <-timerDelay.C:

			//check if a report on the polling status is due, report to logs if
			//it is
			select {
			case <-reportTicker.C:
				globals.Log.DEBUG.Printf("Over the passed %v "+
					"gateway has been checked %v time and %v messages recieved",
					reportDuration, NumChecks, NumMessages)
			default:
			}

			NumChecks++

			var err error

			encryptedMessages, err = rm.receiveMessagesFromGateway(session, &pollingMessage, receptionHost)

			if err != nil {

				if strings.Contains(err.Error(), "Client has exceeded communications rate limit") {
					globals.Log.WARN.Printf("Rate limit excceded on gateway, pausing polling for 5 seconds")
					time.Sleep(5 * time.Second)
				}
				callback(err)
			}
			NumMessages += len(encryptedMessages)
		case <-rm.rekeyChan:
			encryptedMessages = session.PopGarbledMessages()
		}

		if len(encryptedMessages) != 0 {

			decryptedMessages, senders, garbledMessages := rm.decryptMessages(session, encryptedMessages)

			if len(garbledMessages) != 0 {
				session.AppendGarbledMessage(garbledMessages...)
			}

			if decryptedMessages != nil {
				for i := range decryptedMessages {
					// TODO Handle messages that do not need partitioning
					assembledMessage := rm.collator.AddMessage(decryptedMessages[i],
						senders[i], time.Minute)
					if assembledMessage != nil {
						// we got a fully assembled message. let's broadcast it
						broadcastMessageReception(assembledMessage, session.GetSwitchboard())
					}
				}
			}
		}
	}
}

func handleE2EReceiving(session user.Session,
	message *format.Message) (*id.User, bool, error) {
	keyFingerprint := message.GetKeyFP()

	// Lookup reception key
	recpKey := session.GetKeyStore().
		GetRecvKey(keyFingerprint)

	rekey := false
	if recpKey == nil {
		// TODO Handle sending error message to SW
		return nil, false, fmt.Errorf("E2EKey for matching fingerprint not found, can't process message")
	} else if recpKey.GetOuterType() == parse.Rekey {
		// If key type is rekey, the message is a rekey from partner
		rekey = true
	}

	sender := recpKey.GetManager().GetPartner()

	globals.Log.DEBUG.Printf("E2E decrypting message")
	var err error
	if rekey {
		err = crypto.E2EDecryptUnsafe(session.GetE2EGroup(), recpKey.GetKey(), message)
	} else {
		err = crypto.E2EDecrypt(session.GetE2EGroup(), recpKey.GetKey(), message)
	}

	if err != nil {
		// TODO handle Garbled message to SW
	}

	// Get partner from Key Manager of receiving key
	// since there is no space in message for senderID
	// Get decrypted partner public key from message
	// Send rekey message to switchboard
	if rekey {
		partner := recpKey.GetManager().GetPartner()
		partnerPubKey := message.Contents.Get()
		rekeyMsg := &parse.Message{
			Sender: partner,
			TypedBody: parse.TypedBody{
				MessageType: int32(cmixproto.Type_NO_TYPE),
				Body:        partnerPubKey,
			},
			InferredType: parse.Rekey,
			Receiver:     session.GetCurrentUser().User,
		}
		go session.GetSwitchboard().Speak(rekeyMsg)
	}
	return sender, rekey, err
}

func (rm *ReceptionManager) receiveMessagesFromGateway(session user.Session,
	pollingMessage *pb.ClientRequest, receiveGateway *connect.Host) ([]*format.Message, error) {
	// Get the last message ID received
	pollingMessage.LastMessageID = session.GetLastMessageID()
	// FIXME: dont do this over an over

	// Gets a list of mssages that are newer than the last one recieved
	messageIDs, err := rm.Comms.SendCheckMessages(receiveGateway, pollingMessage)

	if err != nil {
		return nil, err
	}

	if len(messageIDs.IDs) < 0 {
		globals.Log.DEBUG.Printf("Checking novelty of %v messageIDs", len(messageIDs.IDs))
	}

	messages := make([]*format.Message, len(messageIDs.IDs))
	mIDs := make([]string, len(messageIDs.IDs))

	// fixme: this could miss messages if the client has not seen them but
	// the gateway say them before a message the client has seen

	// Loops through every new message and retrieves it
	bufLoc := 0
	for _, messageID := range messageIDs.IDs {
		// Get the first unseen message from the list of IDs
		rm.recievedMesageLock.RLock()
		_, received := rm.receivedMessages[messageID]
		rm.recievedMesageLock.RUnlock()
		if !received {
			globals.Log.INFO.Printf("Got a message waiting on the gateway: %v",
				messageID)
			// We haven't seen this message before.
			// So, we should retrieve it from the gateway.
			newMessage, err := rm.Comms.SendGetMessage(receiveGateway,
				&pb.ClientRequest{
					UserID: session.GetCurrentUser().User.
						Bytes(),
					LastMessageID: messageID,
				})
			if err != nil {
				globals.Log.WARN.Printf(
					"Couldn't receive message with ID %v while"+
						" polling gateway", messageID)
			} else {
				if newMessage.PayloadA == nil ||
					newMessage.PayloadB == nil {
					globals.Log.INFO.Println("Message fields not populated")
					continue
				}

				msg := format.NewMessage()
				msg.SetPayloadA(newMessage.PayloadA)
				msg.SetDecryptedPayloadB(newMessage.PayloadB)

				globals.Log.WARN.Printf(
					"Loc: %d, %v", bufLoc, messageID)
				messages[bufLoc] = msg
				mIDs[bufLoc] = messageID
				bufLoc++
			}
		}
	}
	// record that the messages were recieved so they are not re-retrieved
	if bufLoc > 0 {
		for i := 0; i < bufLoc; i++ {
			globals.Log.INFO.Printf(
				"Adding message ID %v to received message IDs", mIDs[i])
			rm.recievedMesageLock.Lock()
			rm.receivedMessages[mIDs[i]] = struct{}{}
			rm.recievedMesageLock.Unlock()
		}
		session.SetLastMessageID(mIDs[bufLoc-1])
		err = session.StoreSession()
		if err != nil {
			globals.Log.ERROR.Printf("Could not store session "+
				"after messages received from gateway: %+v", err)
		}
	}

	return messages[:bufLoc], nil
}

func (rm *ReceptionManager) decryptMessages(session user.Session,
	encryptedMessages []*format.Message) ([]*format.Message, []*id.User,
	[]*format.Message) {

	messages := make([]*format.Message, len(encryptedMessages))
	senders := make([]*id.User, len(encryptedMessages))
	messagesSendersLoc := 0

	garbledMessages := make([]*format.Message, len(encryptedMessages))
	garbledMessagesLoc := 0

	for _, msg := range encryptedMessages {
		var err error = nil
		var rekey bool
		var unpadded []byte
		var sender *id.User
		garbled := false

		// If message is E2E, handle decryption
		if e2e.IsUnencrypted(msg) {
			// If message is non E2E, need to un-pad payload
			unpadded, err = e2e.Unpad(msg.Contents.Get())
			if err == nil {
				msg.Contents.SetRightAligned(unpadded)
			}

			keyFP := msg.AssociatedData.GetKeyFP()
			sender = id.NewUserFromBytes(keyFP[:])
		} else {
			sender, rekey, err = handleE2EReceiving(session, msg)

			if err == errE2ENotFound {
				garbled = true
				err = nil
			}
		}

		if err != nil {
			globals.Log.WARN.Printf(
				"Message did not decrypt properly, "+
					"not adding to messages array: %v", err.Error())
		} else if rekey {
			globals.Log.INFO.Printf("Correctly processed rekey message," +
				" not adding to messages array")
		} else if garbled {
			garbledMessages[garbledMessagesLoc] = msg
			garbledMessagesLoc++
		} else {
			messages[messagesSendersLoc] = msg
			senders[messagesSendersLoc] = sender
			messagesSendersLoc++
		}
	}

	return messages[:messagesSendersLoc], senders[:messagesSendersLoc], garbledMessages[:garbledMessagesLoc]
}

func broadcastMessageReception(message *parse.Message,
	listeners *switchboard.Switchboard) {

	listeners.Speak(message)
}

// skipErrChecker checks checks if the error is fatal or should be ignored
func skipErrChecker(err error) bool {
	if strings.Contains(err.Error(), "Could not find any message IDs for this user") {
		return true
	}

	return false

}
