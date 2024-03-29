////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package cmix

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/elixxir/client/v4/cmix/attempts"
	"gitlab.com/elixxir/client/v4/cmix/rounds"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/cmix/gateway"
	"gitlab.com/elixxir/client/v4/cmix/nodes"
	"gitlab.com/elixxir/client/v4/event"
	"gitlab.com/elixxir/client/v4/stoppable"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/comms/network"
	"gitlab.com/elixxir/crypto/cmix"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/primitives/excludedRounds"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"gitlab.com/xx_network/primitives/netTime"
)

// TargetedCmixMessage defines a recipient target pair in a sendMany cMix
// message.
type TargetedCmixMessage struct {
	Recipient   *id.ID
	Payload     []byte
	Fingerprint format.Fingerprint
	Service     Service
	Mac         []byte
}

// SendMany sends many "raw" cMix message payloads to the provided
// recipients all in the same round.
// Returns the round ID of the round the payloads was sent or an error if it
// fails.
// This does not have end-to-end encryption on it and is used exclusively as
// a send for higher order cryptographic protocols. Do not use unless
// implementing a protocol on top.
// Due to sending multiple payloads, this leaks more metadata than a
// standard cMix send and should be in general avoided.
//
//	recipient - cMix ID of the recipient.
//	fingerprint - Key Fingerprint. 256-bit field to store a 255-bit
//	   fingerprint, highest order bit must be 0 (panic otherwise). If your
//	   system does not use key fingerprints, this must be random bits.
//	service - Reception Service. The backup way for a client to identify
//	   messages on receipt via trial hashing and to identify notifications.
//	   If unused, use message.GetRandomService to fill the field with
//	   random data.
//	payload - Contents of the message. Cannot exceed the payload size for a
//	   cMix message (panic otherwise).
//	mac - 256-bit field to store a 255-bit mac, highest order bit must be 0
//	   (panic otherwise). If used, fill with random bits.
//
// Will return an error if the network is unhealthy or if it fails to send
// (along with the reason). Blocks until successful send or err.
// WARNING: Do not roll your own crypto
func (c *client) SendMany(messages []TargetedCmixMessage,
	params CMIXParams) (rounds.Round, []ephemeral.Id, error) {
	recipients := recipientsFromTargetedMessage(messages)
	assembler := func(rid id.Round) ([]TargetedCmixMessage, error) {
		return messages, nil
	}

	return c.sendManyWithAssembler(recipients, assembler, params)
}

// SendManyWithAssembler sends variable cMix payloads to the provided recipients.
// The payloads sent are based on the ManyMessageAssembler function passed in,
// which accepts a round ID and returns the necessary payload data.
// Returns the round IDs of the rounds the payloads were sent or an error if it
// fails.
// This does not have end-to-end encryption on it and is used exclusively as
// a send operation for higher order cryptographic protocols. Do not use unless
// implementing a protocol on top.
//
//	recipients - cMix IDs of the recipients.
//	assembler - ManyMessageAssembler function, accepting round ID and returning
//	            a list of TargetedCmixMessage.
//
// Will return an error if the network is unhealthy or if it fails to send
// (along with the reason). Blocks until successful sends or errors.
// WARNING: Do not roll your own crypto.
func (c *client) SendManyWithAssembler(recipients []*id.ID,
	assembler ManyMessageAssembler, params CMIXParams) (
	rounds.Round, []ephemeral.Id, error) {
	return c.sendManyWithAssembler(recipients, assembler, params)
}

// sendManyWithAssembler wraps the passed in ManyMessageAssembler in a
// manyMessageAssembler for sendManyCmixHelper.
func (c *client) sendManyWithAssembler(recipients []*id.ID,
	assembler ManyMessageAssembler, params CMIXParams) (rounds.Round,
	[]ephemeral.Id, error) {

	if !c.Monitor.IsHealthy() {
		return rounds.Round{}, []ephemeral.Id{},
			errors.New("Cannot send cMix message when the" +
				" network is not healthy")
	}

	assemblerFunc := func(rid id.Round) ([]assembledCmixMessage, error) {
		messages, err := assembler(rid)
		if err != nil {
			return nil, err
		}

		acms := make([]assembledCmixMessage, len(messages))
		for i := range messages {
			msg := format.NewMessage(c.session.GetCmixGroup().GetP().ByteLen())
			msg.SetKeyFP(messages[i].Fingerprint)
			msg.SetContents(messages[i].Payload)
			msg.SetMac(messages[i].Mac)
			sih, err := messages[i].Service.Hash(messages[i].Recipient, msg.GetContents())
			if err != nil {
				return nil, err
			}
			msg.SetSIH(sih)

			acms[i] = assembledCmixMessage{
				Recipient: messages[i].Recipient,
				Message:   msg,
			}
		}
		return acms, nil
	}

	return sendManyCmixHelper(c.Sender, assemblerFunc, recipients, params,
		c.instance, c.session.GetCmixGroup(), c.Registrar, c.rng, c.events,
		c.session.GetTransmissionID(), c.comms, c.attemptTracker)
}

// assembledCmixMessage is a message structure containing the ready-to-send
// Message (format.Message) and the Recipient that the message is intended.
type assembledCmixMessage struct {
	Recipient *id.ID
	Message   format.Message
}

// sendManyCmixHelper is a helper function for client.SendManyCMIX.
//
// NOTE: Payloads sent are not end-to-end encrypted, metadata is NOT protected
// with this call; see SendE2E for end-to-end encryption and full privacy
// protection. Internal SendMany, which bypasses the network check, will
// attempt to send to the network without checking state. It has a built-in
// retry system which can be configured through the params object.
//
// If the message is successfully sent, the ID of the round sent it is returned,
// which can be registered with the network instance to get a callback on its
// status.
func sendManyCmixHelper(sender gateway.Sender, assembler manyMessageAssembler,
	recipients []*id.ID, param CMIXParams, instance *network.Instance,
	grp *cyclic.Group, registrar nodes.Registrar, rng *fastRNG.StreamGenerator,
	events event.Reporter, senderId *id.ID, comms SendCmixCommsInterface,
	attemptTracker attempts.SendAttemptTracker) (
	rounds.Round, []ephemeral.Id, error) {

	if param.RoundTries == 0 {
		return rounds.Round{}, []ephemeral.Id{},
			errors.Errorf("invalid parameter set, "+
				"RoundTries cannot be 0: %+v", param)
	}

	timeStart := netTime.Now()
	var attempted excludedRounds.ExcludedRounds
	if param.ExcludedRounds != nil {
		attempted = param.ExcludedRounds
	} else {
		attempted = excludedRounds.NewSet()
	}

	maxTimeout := sender.GetHostParams().SendTimeout

	stream := rng.GetStream()
	defer stream.Close()

	recipientsStr := recipientsToStrings(recipients)

	jww.INFO.Printf("[SendMany-%s] Looking for round to send cMix "+
		"messages to [%s]", param.DebugTag, recipientsStr)

	numAttempts := 0
	if !param.Probe {
		optimalAttempts, ready := attemptTracker.GetOptimalNumAttempts()
		if ready {
			numAttempts = optimalAttempts
			jww.INFO.Printf("[SendMany-%s] Looking for round to send cMix "+
				"messages to %s, sending non probe with %d optimalAttempts",
				param.DebugTag, recipientsStr, numAttempts)
		} else {
			numAttempts = 4
			jww.INFO.Printf("[SendMany-%s] Looking for round to send cMix "+
				"messages to %s, sending non probe with %d non optimalAttempts, "+
				"insufficient data", param.DebugTag, recipientsStr, numAttempts)
		}
	} else {
		jww.INFO.Printf("[SendMany-%s] Looking for round to send cMix messages "+
			"to %s, sending probe with %d Attempts, insufficient data",
			param.DebugTag, recipientsStr, numAttempts)
		defer attemptTracker.SubmitProbeAttempt(numAttempts)
	}

	for numRoundTries := uint(0); numRoundTries < param.RoundTries; numRoundTries,
		numAttempts = numRoundTries+1, numAttempts+1 {
		elapsed := netTime.Since(timeStart)

		if elapsed > param.Timeout {
			jww.INFO.Printf("[SendMany-%s] No rounds to send to %s "+
				"were found before timeout %s", param.DebugTag,
				recipientsStr, param.Timeout)
			return rounds.Round{}, []ephemeral.Id{},
				errors.New("sending cMix message timed out")
		}

		if numRoundTries > 0 {
			jww.INFO.Printf("[SendMany-%s] Attempt %d to find round to "+
				"send message to %s", param.DebugTag,
				numRoundTries+1, recipientsStr)
		}

		remainingTime := param.Timeout - elapsed

		// Find the best round to send to, excluding attempted rounds
		bestRound, _, _ := instance.GetWaitingRounds().GetUpcomingRealtime(
			remainingTime, attempted, numAttempts, sendTimeBuffer)
		if bestRound == nil {
			continue
		}

		msgs, err := assembler(id.Round(bestRound.ID))
		if err != nil {
			jww.ERROR.Printf("Failed to compile messages: %+v", err)
			return rounds.Round{}, []ephemeral.Id{}, err
		}

		// Determine whether the selected round contains any nodes that are
		// blacklisted by the params.Network object
		containsBlacklisted := false
		if param.BlacklistedNodes != nil {
			for _, nodeId := range bestRound.Topology {
				nid := &id.ID{}
				copy(nid[:], nodeId)
				if _, isBlacklisted := param.BlacklistedNodes[*nid]; isBlacklisted {
					containsBlacklisted = true
					break
				}
			}
		}
		if containsBlacklisted {
			jww.WARN.Printf("[SendMany-%s] Round %d contains blacklisted "+
				"nodes, skipping...", param.DebugTag, bestRound.ID)
			continue
		}

		// flip leading bits randomly to thwart a tagging attack.
		// See SetGroupBits for more info
		for i := range msgs {
			cmix.SetGroupBits(msgs[i].Message, grp, stream)
		}

		// Retrieve host and key information from round
		msgDigests := messageListToDigestStrings(msgs)
		firstGateway, roundKeys, err := processRound(
			registrar, bestRound, recipientsStr, msgDigests)
		if err != nil {
			jww.INFO.Printf("[SendMany-%s] Error processing round: %v",
				param.DebugTag, err)
			jww.WARN.Printf("[SendMany-%s] SendMany failed to "+
				"process round %d (will retry): %+v", param.DebugTag,
				bestRound.ID, err)
			continue
		}

		// Build a slot for every message and recipient
		slots := make([]*pb.GatewaySlot, len(msgs))
		encMsgs := make([]format.Message, len(msgs))
		ephemeralIDs := make([]ephemeral.Id, len(msgs))
		for i, msg := range msgs {
			slots[i], encMsgs[i], ephemeralIDs[i], err = buildSlotMessage(
				msg.Message, msg.Recipient, firstGateway, stream, senderId,
				bestRound, roundKeys)
			if err != nil {
				jww.INFO.Printf("[SendMany-%s] Error building slot "+
					"received: %v", param.DebugTag, err)
				return rounds.Round{}, []ephemeral.Id{}, errors.Errorf("failed to build "+
					"slot message for %s: %+v", msg.Recipient, err)
			}
		}

		// Serialize lists into a printable format
		ephemeralIDsString := ephemeralIdListToString(ephemeralIDs)
		encMsgsDigest := messagesToDigestString(encMsgs)

		jww.INFO.Printf("[SendMany-%s]Sending to EphIDs [%s] (%s) on round %d, "+
			"(msgDigest: %s, ecrMsgDigest: %s) via gateway %s", param.DebugTag,
			ephemeralIDsString, recipientsStr, bestRound.ID, msgDigests,
			encMsgsDigest, firstGateway)

		// Wrap slots in the proper message type
		wrappedMessage := &pb.GatewaySlots{
			Messages: slots,
			RoundID:  bestRound.ID,
		}

		// Send the payload
		sendFunc := func(host *connect.Host, target *id.ID,
			timeout time.Duration) (interface{}, error) {
			// Use the smaller of the two timeout durations
			calculatedTimeout := calculateSendTimeout(bestRound, maxTimeout)
			if calculatedTimeout < timeout {
				timeout = calculatedTimeout
			}

			wrappedMessage.Target = target.Marshal()
			result, err := comms.SendPutManyMessages(
				host, wrappedMessage, timeout)
			if err != nil {
				err := handlePutMessageError(firstGateway, registrar,
					recipientsStr, bestRound, err)
				return result, errors.WithMessagef(err,
					"SendMany %s (via %s): %s",
					target, host, unrecoverableError)

			}
			return result, err
		}
		result, err := sender.SendToPreferred(
			[]*id.ID{firstGateway}, sendFunc, param.Stop, param.SendTimeout)

		// Exit if the thread has been stopped
		if stoppable.CheckErr(err) {
			return rounds.Round{}, []ephemeral.Id{}, err
		}

		// If the comm errors or the message fails to send, continue retrying
		if err != nil {
			if !strings.Contains(err.Error(), unrecoverableError) {
				jww.ERROR.Printf("[SendMany-%s] SendMany failed to "+
					"send to EphIDs [%s] (sources: %s) on round %d, trying "+
					"a new round %+v", param.DebugTag, ephemeralIDsString,
					recipientsStr, bestRound.ID, err)
				jww.INFO.Printf("[SendMany-%s] Error received, "+
					"continuing: %v", param.DebugTag, err)
				continue
			} else {
				jww.INFO.Printf("[SendMany-%s] Error received: %v",
					param.DebugTag, err)
			}
			return rounds.Round{}, []ephemeral.Id{}, err
		}

		// Return if it sends properly
		gwSlotResp := result.(*pb.GatewaySlotResponse)
		if gwSlotResp.Accepted {
			m := fmt.Sprintf("[SendMany-%s] Successfully sent to EphIDs "+
				"%s (sources: [%s]) in round %d (msgDigest: %s)",
				param.DebugTag, ephemeralIDsString, recipientsStr,
				bestRound.ID, msgDigests)
			jww.INFO.Print(m)
			events.Report(1, "MessageSendMany", "Metric", m)
			return rounds.MakeRound(bestRound), ephemeralIDs, nil
		} else {
			jww.FATAL.Panicf("[SendMany-%s] Gateway %s returned no "+
				"error, but failed to accept message when sending to EphIDs "+
				"[%s] (%s) on round %d", param.DebugTag, firstGateway,
				ephemeralIDsString, recipientsStr, bestRound.ID)
		}
	}

	return rounds.Round{}, []ephemeral.Id{},
		errors.New("failed to send the message, unknown error")
}
