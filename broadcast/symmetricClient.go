////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package broadcast

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/message"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
)

// Error messages.
const (
	// symmetricClient.Broadcast
	errNetworkHealth = "cannot send broadcast when the network is not healthy"
	errPayloadSize   = "size of payload %d must be %d"
)

// Tags.
const (
	symmCMixSendTag              = "SymmBcast"
	symmetricBroadcastServiceTag = "SymmetricBroadcast"
)

// MaxPayloadSize returns the maximum size for a broadcasted payload.
func (bc *broadcastClient) MaxSymmetricPayloadSize() int {
	return bc.net.GetMaxMessageLength()
}

// Broadcast broadcasts the payload to the channel.
func (bc *broadcastClient) Broadcast(payload []byte, cMixParams cmix.CMIXParams) (
	id.Round, ephemeral.Id, error) {
	if !bc.net.IsHealthy() {
		return 0, ephemeral.Id{}, errors.New(errNetworkHealth)
	}

	if len(payload) != bc.MaxSymmetricPayloadSize() {
		return 0, ephemeral.Id{},
			errors.Errorf(errPayloadSize, len(payload), bc.MaxSymmetricPayloadSize())
	}

	// Encrypt payload
	rng := bc.rng.GetStream()
	encryptedPayload, mac, fp := bc.channel.EncryptSymmetric(payload, rng)
	rng.Close()

	// Create service
	service := message.Service{
		Identifier: bc.channel.ReceptionID.Bytes(),
		Tag:        symmetricBroadcastServiceTag,
	}

	if cMixParams.DebugTag == cmix.DefaultDebugTag {
		cMixParams.DebugTag = symmCMixSendTag
	}

	return bc.net.Send(
		bc.channel.ReceptionID, fp, service, encryptedPayload, mac, cMixParams)
}
