////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"encoding/json"

	"gitlab.com/elixxir/client/v4/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/single"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/xx_network/primitives/id"
)

////////////////////////////////////////////////////////////////////////////////
// Public Wrapper Methods                                                     //
////////////////////////////////////////////////////////////////////////////////

// TransmitSingleUse transmits payload to recipient via single-use.
//
// Parameters:
//   - e2eID - ID of the e2e object in the tracker
//   - recipient - marshalled contact.Contact object
//   - tag - identifies the single-use message
//   - payload - message contents
//   - paramsJSON - JSON marshalled single.RequestParams
//   - responseCB - the callback that will be called when a response is received
//
// Returns:
//   - []byte - the JSON marshalled bytes of the SingleUseSendReport object,
//     which can be passed into WaitForRoundResult to see if the send succeeded.
func TransmitSingleUse(e2eID int, recipient []byte, tag string, payload,
	paramsJSON []byte, responseCB SingleUseResponse) ([]byte, error) {
	e2eCl, err := e2eTrackerSingleton.get(e2eID)
	if err != nil {
		return nil, err
	}

	recipientContact, err := contact.Unmarshal(recipient)
	if err != nil {
		return nil, err
	}

	rcb := &singleUseResponse{response: responseCB}

	params, err := parseSingleUseParams(paramsJSON)
	if err != nil {
		return nil, err
	}

	rids, eid, err := single.TransmitRequest(recipientContact, tag, payload,
		rcb, params, e2eCl.api.GetCmix(), e2eCl.api.GetRng().GetStream(),
		e2eCl.api.GetStorage().GetE2EGroup())

	if err != nil {
		return nil, err
	}
	sr := SingleUseSendReport{
		EphID:       eid.EphId.Int64(),
		ReceptionID: eid.Source,
		RoundsList:  makeRoundsList(rids...),
		RoundURL:    getRoundURL(rids[0]),
	}
	return json.Marshal(sr)
}

// Listen starts a single-use listener on a given tag using the passed in E2e
// object and SingleUseCallback func.
//
// Parameters:
//   - e2eID - ID of the e2e object in the tracker
//   - tag - identifies the single-use message
//   - cb - the callback that will be called when a response is received
//
// Returns:
//   - Stopper - an interface containing a function used to stop the listener
func Listen(e2eID int, tag string, cb SingleUseCallback) (Stopper, error) {
	e2eCl, err := e2eTrackerSingleton.get(e2eID)
	if err != nil {
		return nil, err
	}

	suListener := singleUseListener{scb: cb}
	dhPk, err := e2eCl.api.GetReceptionIdentity().GetDHKeyPrivate()
	if err != nil {
		return nil, err
	}
	l := single.Listen(tag, e2eCl.api.GetReceptionIdentity().ID, dhPk,
		e2eCl.api.GetCmix(), e2eCl.api.GetStorage().GetE2EGroup(), suListener)
	return &stopper{l: l}, nil
}

// JSON Types

// SingleUseSendReport is the bindings-layer struct used to represent
// information returned by single.TransmitRequest.
//
// SingleUseSendReport JSON example:
//
//	{
//	 "Rounds":[1,5,9],
//	 "RoundURL": "https://dashboard.xx.network/rounds/25?xxmessenger=true",
//	 "EphID":1655533,
//	 "ReceptionID":"emV6aW1hAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD"}
//	}
type SingleUseSendReport struct {
	RoundsList
	RoundURL    string
	ReceptionID *id.ID
	EphID       int64
}

// SingleUseResponseReport is the bindings-layer struct used to represent
// information passed to the single.Response callback interface in response to
// single.TransmitRequest.
//
// SingleUseResponseReport JSON example:
//
//	{
//	 "Rounds":[1,5,9],
//	 "RoundURL": "https://dashboard.xx.network/rounds/25?xxmessenger=true",
//	 "Payload":"rSuPD35ELWwm5KTR9ViKIz/r1YGRgXIl5792SF8o8piZzN6sT4Liq4rUU/nfOPvQEjbfWNh/NYxdJ72VctDnWw==",
//	 "EphID":1655533,
//	 "ReceptionID":"emV6aW1hAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD"},
//	 "Err":"",
//	}
type SingleUseResponseReport struct {
	RoundsList
	RoundURL    string
	Payload     []byte
	ReceptionID *id.ID
	EphID       int64
	Err         error
}

// SingleUseCallbackReport is the bindings-layer struct used to represent
// single -use messages received by a callback passed into single.Listen.
//
// SingleUseCallbackReport JSON example:
//
//	{
//	  "Rounds":[1,5,9],
//	  "RoundURL": "https://dashboard.xx.network/rounds/25?xxmessenger=true",
//	  "Payload":"rSuPD35ELWwm5KTR9ViKIz/r1YGRgXIl5792SF8o8piZzN6sT4Liq4rUU/nfOPvQEjbfWNh/NYxdJ72VctDnWw==",
//	  "Partner":"emV6aW1hAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD",
//	  "EphID":1655533,
//	  "ReceptionID":"emV6aW1hAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD"}
//	}
type SingleUseCallbackReport struct {
	RoundsList
	RoundURL    string
	Payload     []byte
	Partner     *id.ID
	EphID       int64
	ReceptionID *id.ID
}

////////////////////////////////////////////////////////////////////////////////
// Function Types                                                             //
////////////////////////////////////////////////////////////////////////////////

// Stopper is a public interface returned by Listen, allowing users to stop the
// registered listener.
type Stopper interface {
	Stop()
}

// SingleUseCallback func is passed into Listen and called when messages are
// received.
//
// Parameters:
//   - callbackReport - the JSON marshalled bytes of the SingleUseCallbackReport
//     object, which can be passed into Cmix.WaitForRoundResult to see if the
//     send operation succeeded.
type SingleUseCallback interface {
	Callback(callbackReport []byte, err error)
}

// SingleUseResponse is the public facing callback function passed by bindings
// clients into TransmitSingleUse.
//
// Parameters:
//   - callbackReport - the JSON marshalled bytes of the SingleUseResponseReport
//     object, which can be passed into Cmix.WaitForRoundResult to see if the
//     send operation succeeded.
type SingleUseResponse interface {
	Callback(responseReport []byte, err error)
}

////////////////////////////////////////////////////////////////////////////////
// Callback Wrappers                                                          //
////////////////////////////////////////////////////////////////////////////////

/* Listener Struct */

// singleUseListener is the internal struct used to wrap a SingleUseCallback
// function, which matches the single.Receiver interface.
type singleUseListener struct {
	scb SingleUseCallback
}

// Callback is called whenever a single-use message is heard by the listener
// and translates the info to a SingleUseCallbackReport that is marshalled and
// passed to bindings.
func (sl singleUseListener) Callback(
	req *single.Request, eid receptionID.EphemeralIdentity, rl []rounds.Round) {
	var rids []id.Round
	for _, r := range rl {
		rids = append(rids, r.ID)
	}

	// Todo: what other info from req needs to get to bindings
	scr := SingleUseCallbackReport{
		Payload:     req.GetPayload(),
		RoundsList:  makeRoundsList(rids...),
		RoundURL:    getRoundURL(rids[0]),
		Partner:     req.GetPartner(),
		EphID:       eid.EphId.Int64(),
		ReceptionID: eid.Source,
	}

	sl.scb.Callback(json.Marshal(scr))
}

/* Listener stopper */

// stopper is the internal struct backing the Stopper interface, allowing us
// to pass the listener Stop method to the bindings layer.
type stopper struct {
	l single.Listener
}

func (s *stopper) Stop() {
	s.l.Stop()
}

/* Response Struct */

// singleUseResponse is the private struct backing SingleUseResponse, which
// subscribes to the single.Response interface.
type singleUseResponse struct {
	response SingleUseResponse
}

// Callback builds a SingleUseSendReport and passes the JSON marshalled version
// into the callback.
func (sr singleUseResponse) Callback(payload []byte,
	receptionID receptionID.EphemeralIdentity, rounds []rounds.Round, err error) {
	var rids []id.Round
	for _, r := range rounds {
		rids = append(rids, r.ID)
	}
	// If the callback get's called on a timed out request, the rounds list will be
	// empty. In this case, roundURL should be empty as well.
	roundURL := ""
	if len(rids) > 0 {
		roundURL = getRoundURL(rids[0])
	}
	sendReport := SingleUseResponseReport{
		RoundsList:  makeRoundsList(rids...),
		RoundURL:    roundURL,
		ReceptionID: receptionID.Source,
		EphID:       receptionID.EphId.Int64(),
		Payload:     payload,
		Err:         err,
	}
	sr.response.Callback(json.Marshal(&sendReport))
}
