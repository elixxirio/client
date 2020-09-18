////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package network

import (
	"fmt"
	"gitlab.com/elixxir/client/context"
	"gitlab.com/elixxir/client/context/message"
	"gitlab.com/elixxir/client/context/params"
	"gitlab.com/elixxir/client/context/stoppable"
	"gitlab.com/elixxir/client/network/parse"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/crypto/e2e"
	"gitlab.com/elixxir/primitives/format"
	//jww "github.com/spf13/jwalterweatherman"
)

// ReceiveMessages is called by a MessageReceiver routine whenever new CMIX
// messages are available at a gateway.
func ReceiveMessages(ctx *context.Context, roundInfo *pb.RoundInfo,
	network *Manager) {
	msgs := getMessagesFromGateway(ctx, roundInfo)
	for _, m := range msgs {
		receiveMessage(ctx, m)
	}
}

// StartMessageReceivers starts a worker pool of message receivers, which listen
// on a channel for rounds in which to check for messages and run them through
// processing.
func StartMessageReceivers(ctx *context.Context,
	network *Manager) stoppable.Stoppable {
	stoppers := stoppable.NewMulti("MessageReceivers")
	opts := params.GetDefaultNetwork()
	receiverCh := network.GetRoundUpdateCh()
	for i := 0; i < opts.NumWorkers; i++ {
		stopper := stoppable.NewSingle(
			fmt.Sprintf("MessageReceiver%d", i))
		go MessageReceiver(ctx, receiverCh, stopper.Quit())
		stoppers.Add(stopper)
	}
	return stoppers
}

// MessageReceiver waits until quit signal or there is a round available
// for which to check for messages available on the round updates channel.
func MessageReceiver(ctx *context.Context, network *Manager,
	updatesCh chan *pb.RoundInfo, quitCh <-chan struct{}) {
	done := false
	for !done {
		select {
		case <-quitCh:
			done = true
		case round := <-updatesCh:
			ReceiveMessages(ctx, round, network)
		}
	}
}

func getMessagesFromGateway(ctx *context.Context,
	roundInfo *pb.RoundInfo, network *Manager) []*pb.Slot {
	comms := network.Comms
	roundTop := roundInfo.GetTopology()
	lastNode := id.NewIdFromBytes(roundTop[len(roundTop-1)])
	lastGw := lastNode.SetType(id.Gateway)
	gwHost := comms.GetHost(lastGw)

	user := ctx.Session.User().GetCryptographicIdentity()
	userID := user.GetUserID().Bytes()

	// First get message id list
	msgReq := pb.GetMessages{
		ClientID: userID,
		RoundID:  roundInfo.ID,
	}
	msgResp, err := comms.RequestMessages(gwHost, msgReq)
	if err != nil {
		jww.ERROR.Printf(err.Error())
		return nil
	}

	// If no error, then we have checked the round and finished processing
	ctx.Session.GetCheckedRounds.Check(roundInfo.ID)
	network.Processing.Done(roundInfo.ID)

	if !msgResp.GetHasRound() {
		jww.ERROR.Printf("host %s does not have roundID: %d",
			gwHost, roundInfo.ID)
		return nil
	}

	msgs := msgResp.GetMessages()

	if msgs == nil || len(msgs) == 0 {
		jww.ERROR.Printf("host %s has no messages for client %s "+
			" in round %d", gwHost, user, roundInfo.ID)
		return nil
	}

	return msgs

}

func receiveMessage(ctx *context.Context, rawMsg *pb.Slot) {
	// We've done all the networking, now process the message
	msg := format.NewMessage()
	msg.SetPayloadA(rawMsg.GetPayloadA())
	msg.SetPayloadB(rawMsg.GetPayloadB())
	fingerprint := msg.GetKeyFP()

	sess := ctx.Session
	e2eKS := sess.GetE2e()
	partitioner := sess.GetPartition()

	var sender *id.ID
	var unencrypted *format.Message
	var encTy message.EncryptionType

	key, isE2E := e2eKS.PopKey(fingerprint)
	if key != nil && isE2E {
		// Decrypt encrypted message
		unencrypted, err := key.Decrypt(msg)
		// set sender only if decryption worked
		// otherwise don't so it gets sent to garbled message
		if err != nil {
			jww.ERROR.Printf(err.Error())
		} else {
			sender = key.Session.GetPartner()
		}
		encTy = message.E2E
	} else {
		// SendUnsafe Message?
		isUnencrypted, sender := e2e.IsUnencrypted(msg)
		if isUnencrypted {
			unencrypted = msg
		}
		encTy = message.None
	}

	// Save off garbled messages
	if unencrypted == nil || sender == nil {
		jww.ERROR.Printf("garbled message: %s", msg)
		sess.GetGarbledMessages().Add(msg)
		return
	}

	// Process the decrypted/unencrypted message partition, to see if
	// we get a full message
	xxMsg, ok := partitioner.HandlePartition(sender, encTy, unencrypted)
	// Share completed message on switchboard
	if ok && xxMsg != nil {
		ct.Switchboard.Speak(xxMsg)
	}
}
