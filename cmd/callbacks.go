///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

// callbacks.go implements all of the required api callbacks for the cli
package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/xxdk"

	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/rounds"
	"gitlab.com/elixxir/client/e2e/receive"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/xx_network/primitives/id"
)

// authCallbacks implements the auth.Callbacks interface.
type authCallbacks struct {
	autoConfirm bool
	confCh      chan *id.ID
	params      xxdk.E2EParams
}

func makeAuthCallbacks(autoConfirm bool, params xxdk.E2EParams) *authCallbacks {
	return &authCallbacks{
		autoConfirm: autoConfirm,
		confCh:      make(chan *id.ID, 10),
		params:      params,
	}
}

func (a *authCallbacks) Request(requestor contact.Contact,
	receptionID receptionID.EphemeralIdentity,
	round rounds.Round, messenger *xxdk.E2e) {
	msg := fmt.Sprintf("Authentication channel request from: %s\n",
		requestor.ID)
	jww.INFO.Printf(msg)
	fmt.Printf(msg)
	if a.autoConfirm {
		jww.INFO.Printf("Channel Request: %s",
			requestor.ID)
		if viper.GetBool(verifySendFlag) { // Verify message sends were successful
			acceptChannelVerified(messenger, requestor.ID, a.params)
		} else {
			acceptChannel(messenger, requestor.ID)
		}

		a.confCh <- requestor.ID
	}

}

func (a *authCallbacks) Confirm(requestor contact.Contact,
	_ receptionID.EphemeralIdentity,
	_ rounds.Round, _ *xxdk.E2e) {
	jww.INFO.Printf("Channel Confirmed: %s", requestor.ID)
	a.confCh <- requestor.ID
}

func (a *authCallbacks) Reset(requestor contact.Contact,
	_ receptionID.EphemeralIdentity,
	_ rounds.Round, _ *xxdk.E2e) {
	msg := fmt.Sprintf("Authentication channel reset from: %s\n",
		requestor.ID)
	jww.INFO.Printf(msg)
	fmt.Printf(msg)
}

func registerMessageListener(client *xxdk.E2e) chan receive.Message {
	recvCh := make(chan receive.Message, 10000)
	listenerID := client.GetE2E().RegisterChannel("DefaultCLIReceiver",
		receive.AnyUser(), catalog.NoType, recvCh)
	jww.INFO.Printf("Message ListenerID: %v", listenerID)
	return recvCh
}
