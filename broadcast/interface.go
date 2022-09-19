////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package broadcast

import (
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/message"
	"gitlab.com/elixxir/client/cmix/rounds"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/rsa"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"time"
)

// ListenerFunc is registered when creating a new broadcasting channel and
// receives all new broadcast messages for the channel.
type ListenerFunc func(payload []byte,
	receptionID receptionID.EphemeralIdentity, round rounds.Round)

// Channel is the public-facing interface to interact with broadcast channels
type Channel interface {
	// MaxPayloadSize returns the maximum size for a symmetric broadcast payload
	MaxPayloadSize() int

	// MaxRSAToPublicPayloadSize returns the maximum size for an asymmetric
	// broadcast payload
	MaxRSAToPublicPayloadSize() int

	// Get returns the underlying crypto.Channel
	Get() *crypto.Channel

	// Broadcast broadcasts the payload to the channel. The payload size must be
	// equal to MaxPayloadSize or smaller.
	Broadcast(payload []byte, cMixParams cmix.CMIXParams) (
		rounds.Round, ephemeral.Id, error)

	// BroadcastWithAssembler broadcasts a payload over a symmetric channel.
	// With a payload assembled after the round is selected, allowing the round
	// info to be included in the payload. Network must be healthy to send.
	// Requires a payload of size bc.MaxSymmetricPayloadSize() or smaller
	BroadcastWithAssembler(assembler Assembler, cMixParams cmix.CMIXParams) (
		rounds.Round, ephemeral.Id, error)

	// BroadcastRSAtoPublic broadcasts the payload to the channel. Requires a
	// healthy network state to send Payload length less than or equal to
	// bc.MaxRSAToPublicPayloadSize, and the channel PrivateKey must be passed in
	BroadcastRSAtoPublic(pk rsa.PrivateKey, payload []byte,
		cMixParams cmix.CMIXParams) (rounds.Round, ephemeral.Id, error)

	// BroadcastRSAToPublicWithAssembler broadcasts the payload to the channel with
	// a function which builds the payload based upon the ID of the selected round.
	// Requires a healthy network state to send Payload must be shorter or equal in
	// length to bc.MaxRSAToPublicPayloadSize when returned, and the channel
	// PrivateKey must be passed in
	BroadcastRSAToPublicWithAssembler(
		pk rsa.PrivateKey, assembler Assembler,
		cMixParams cmix.CMIXParams) (rounds.Round, ephemeral.Id, error)

	// RegisterListener registers a listener for broadcast messages
	RegisterListener(listenerCb ListenerFunc, method Method) error

	// Stop unregisters the listener callback and stops the channel's identity
	// from being tracked.
	Stop()
}

// Assembler is a function which allows a bre
type Assembler func(rid id.Round) (payload []byte, err error)

// Client contains the methods from cmix.Client that are required by
// symmetricClient.
type Client interface {
	SendWithAssembler(recipient *id.ID, assembler cmix.MessageAssembler,
		cmixParams cmix.CMIXParams) (rounds.Round, ephemeral.Id, error)
	IsHealthy() bool
	AddIdentity(id *id.ID, validUntil time.Time, persistent bool)
	AddService(clientID *id.ID, newService message.Service,
		response message.Processor)
	DeleteClientService(clientID *id.ID)
	RemoveIdentity(id *id.ID)
	GetMaxMessageLength() int
}
