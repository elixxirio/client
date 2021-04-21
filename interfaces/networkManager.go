///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package interfaces

import (
	"gitlab.com/elixxir/client/interfaces/message"
	"gitlab.com/elixxir/client/interfaces/params"
	"gitlab.com/elixxir/client/stoppable"
	"gitlab.com/elixxir/comms/network"
	"gitlab.com/elixxir/crypto/e2e"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
)

type NetworkManager interface {
	SendE2E(m message.Send, p params.E2E) ([]id.Round, e2e.MessageID, error)
	SendUnsafe(m message.Send, p params.Unsafe) ([]id.Round, error)
	SendCMIX(message format.Message, recipient *id.ID, p params.CMIX) (id.Round, ephemeral.Id, error)
    SendManyCMIX(messages []format.Message,	recipients []*id.ID, p params.CMIX) (id.Round, []ephemeral.Id, error)
	GetInstance() *network.Instance
	GetHealthTracker() HealthTracker
	Follow(report ClientErrorReport) (stoppable.Stoppable, error)
	CheckGarbledMessages()
	InProgressRegistrations() int
}

//for use in key exchange which needs to be callable inside of network
type SendE2E func(m message.Send, p params.E2E) ([]id.Round, e2e.MessageID, error)
