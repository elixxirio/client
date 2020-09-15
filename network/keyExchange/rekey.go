////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package keyExchange

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/context"
	"gitlab.com/elixxir/client/context/message"
	"gitlab.com/elixxir/client/context/params"
	"gitlab.com/elixxir/client/context/utility"
	"gitlab.com/elixxir/client/storage/e2e"
	ds "gitlab.com/elixxir/comms/network/dataStructures"
	"gitlab.com/elixxir/crypto/diffieHellman"
	"gitlab.com/elixxir/primitives/states"
	"time"
)

func CheckKeyExchanges(ctx *context.Context, manager *e2e.Manager) {
	sessions := manager.TriggerNegotiations()
	for _, ses := range sessions {
		go trigger(ctx, manager, ses)
	}
}

// There are two types of key negotiations that can be triggered, creating a new
// session and negotiation, or resenting a negotiation for an already created
// session. They run the same negotiation, the former does it on a newly created
// session while the latter on an extand
func trigger(ctx *context.Context, manager *e2e.Manager, session *e2e.Session) {
	var negotiatingSession *e2e.Session
	switch session.NegotiationStatus() {
	// If the passed session is triggering a negotiation on a new session to
	// replace itself, then create the session
	case e2e.NewSessionTriggered:
		//create the session, pass a nil private key to generate a new one
		negotiatingSession = manager.NewSendSession(nil,
			e2e.GetDefaultSessionParams(), session.GetID())
		//move the state of the triggering session forward
		session.SetNegotiationStatus(e2e.NewSessionCreated)
	// If the session has not successfully negotiated, redo its negotiation
	case e2e.Unconfirmed:
		negotiatingSession = session
	default:
		jww.FATAL.Panicf("Session %s provided invalid e2e "+
			"negotiating status: %s", session, session.NegotiationStatus())
	}

	// send the rekey notification to the partner
	err := negotiate(ctx, negotiatingSession)
	// if sending the negotiation fails, revert the state of the session to
	// unconfirmed so it will be triggered in the future
	if err != nil {
		jww.ERROR.Printf("Failed to do Key Negotiation: %s", err)
		negotiatingSession.SetNegotiationStatus(e2e.Unconfirmed)
	}
}

func negotiate(ctx *context.Context, session *e2e.Session) error {
	e2eStore := ctx.Session.E2e()

	//generate public key
	pubKey := diffieHellman.GeneratePublicKey(session.GetMyPrivKey(),
		e2eStore.GetGroup())

	//build the payload
	payload, err := proto.Marshal(&RekeyTrigger{
		PublicKey: pubKey.Bytes(),
		SessionID: session.GetTrigger().Marshal(),
	})

	//If the payload cannot be marshaled, panic
	if err != nil {
		jww.FATAL.Printf("Failed to marshal payload for Key "+
			"Negotation Trigger with %s", session.GetPartner())
	}

	//send session
	m := message.Send{
		Recipient:   session.GetPartner(),
		Payload:     payload,
		MessageType: message.KeyExchangeTrigger,
	}

	//send the message under the key exchange
	e2eParams := params.GetDefaultE2E()
	e2eParams.Type = params.KeyExchange

	rounds, err := ctx.Manager.SendE2E(m, e2eParams)
	// If the send fails, returns the error so it can be handled. The caller
	// should ensure the calling session is in a state where the Rekey will
	// be triggered next time a key is used
	if err != nil {
		return errors.Errorf("Failed to send the key negotation message "+
			"for %s: %s", session, err)
	}

	//create the runner which will handle the result of sending the messages
	sendResults := make(chan ds.EventReturn, len(rounds))

	//Register the event for all rounds
	roundEvents := ctx.Manager.GetInstance().GetRoundEvents()
	for _, r := range rounds {
		roundEvents.AddRoundEventChan(r, sendResults, 1*time.Minute,
			states.COMPLETED, states.FAILED)
	}

	//Wait until the result tracking responds
	success, numTimeOut, numRoundFail := utility.TrackResults(sendResults, len(rounds))

	// If a single partition of the Key Negotiation request does not
	// transmit, the partner cannot read the result. Log the error and set
	// the session as unconfirmed so it will re-trigger the negotiation
	if !success {
		session.SetNegotiationStatus(e2e.Unconfirmed)
		return errors.Errorf("Key Negotiation for %s failed to "+
			"transmit %v/%v paritions: %v round failures, %v timeouts",
			session, numRoundFail+numTimeOut, len(rounds), numRoundFail,
			numTimeOut)
	}

	// otherwise, the transmission is a success and this should be denoted
	// in the session and the log
	jww.INFO.Printf("Key Negotiation transmission for %s sucesfull",
		session)
	session.SetNegotiationStatus(e2e.Sent)

	return nil
}
