////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package parse

import (
	"gitlab.com/elixxir/crypto/e2e"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/catalog"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/client/v4/e2e/parse/conversation"
	"gitlab.com/elixxir/client/v4/e2e/parse/partition"
	"gitlab.com/elixxir/client/v4/e2e/receive"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
)

const MaxMessageParts = 255

type Partitioner struct {
	baseMessageSize   int
	firstContentsSize int
	partContentsSize  int
	deltaFirstPart    int
	maxSize           int
	conversation      *conversation.Store
	partition         *partition.Store
}

func NewPartitioner(kv versioned.KV, messageSize int) *Partitioner {
	p := Partitioner{
		baseMessageSize:   messageSize,
		firstContentsSize: messageSize - firstHeaderLen,
		partContentsSize:  messageSize - headerLen,
		deltaFirstPart:    firstHeaderLen - headerLen,
		conversation:      conversation.NewStore(kv),
		partition:         partition.NewOrLoad(kv),
	}
	p.maxSize = p.firstContentsSize + (MaxMessageParts-1)*p.partContentsSize

	return &p
}

func (p *Partitioner) Partition(recipient *id.ID, mt catalog.MessageType,
	timestamp time.Time, payload []byte) ([][]byte, uint64, error) {

	if len(payload) > p.maxSize {
		return nil, 0, errors.Errorf("Payload is too long, max payload "+
			"length is %d, received %d", p.maxSize, len(payload))
	}

	// Get the ID of the sent message
	fullMessageID, messageID := p.conversation.Get(recipient).GetNextSendID()

	// Get the number of parts of the message; this equates to just a linear
	// equation
	numParts := uint8((len(payload) + p.deltaFirstPart + p.partContentsSize - 1) / p.partContentsSize)
	parts := make([][]byte, numParts)

	// Create the first message part
	var sub []byte
	sub, payload = splitPayload(payload, p.firstContentsSize)
	parts[0] = newFirstMessagePart(mt, messageID, numParts,
		timestamp, sub, p.baseMessageSize).bytes()

	// Create all subsequent message parts
	for i := uint8(1); i < numParts; i++ {
		sub, payload = splitPayload(payload, p.partContentsSize)
		parts[i] = newMessagePart(messageID, i, sub, p.baseMessageSize).bytes()
	}

	return parts, fullMessageID, nil
}

func (p *Partitioner) HandlePartition(sender *id.ID,
	contents []byte, relationshipFingerprint []byte,
	residue e2e.KeyResidue) (receive.Message, e2e.KeyResidue, bool) {

	if isFirst(contents) {
		// If it is the first message in a set, then handle it as so

		// Decode the message structure
		fm := firstMessagePartFromBytes(contents)

		// Handle the message ID
		messageID := p.conversation.Get(sender).
			ProcessReceivedMessageID(fm.getID())
		storageTimestamp := netTime.Now()
		return p.partition.AddFirst(sender, fm.getType(), messageID,
			fm.getPart(), fm.getNumParts(), fm.getTimestamp(), storageTimestamp,
			fm.getSizedContents(), relationshipFingerprint, residue)
	} else {
		// If it is a subsequent message part, handle it as so
		mp := messagePartFromBytes(contents)
		messageID :=
			p.conversation.Get(sender).ProcessReceivedMessageID(mp.getID())

		return p.partition.Add(sender, messageID, mp.getPart(),
			mp.getSizedContents(), relationshipFingerprint)
	}
}

// FirstPartitionSize returns the max partition payload size for the
// first payload
func (p *Partitioner) FirstPartitionSize() uint {
	return uint(p.firstContentsSize)
}

// SecondPartitionSize returns the max partition payload size for all
// payloads after the first payload
func (p *Partitioner) SecondPartitionSize() uint {
	return uint(p.partContentsSize)
}

// PayloadSize Returns the max payload size for a partitionable E2E
// message
func (p *Partitioner) PayloadSize() uint {
	return uint(p.maxSize)
}

func splitPayload(payload []byte, length int) ([]byte, []byte) {
	if len(payload) < length {
		return payload, payload
	}
	return payload[:length], payload[length:]
}

func isFirst(payload []byte) bool {
	return payload[idLen] == 0
}
