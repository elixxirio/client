///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package conversation

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/xx_network/primitives/netTime"
	"sync"
)

// Storage keys and versions.
const (
	ringBuffPrefix  = "ringBuffPrefix"
	ringBuffKey     = "ringBuffKey"
	ringBuffVersion = 0
	messageKey      = "ringBuffMessageKey"
	messageVersion  = 0
)

// Error messages.
const (
	saveMessageErr    = "failed to save message with message ID %s to storage: %+v"
	loadMessageErr    = "failed to load message with truncated ID %s from storage: %+v"
	loadBuffErr       = "failed to load ring buffer from storage: %+v"
	noMessageFoundErr = "failed to find message with message ID %s "
)

// Buff is a circular buffer which containing Message's.
type Buff struct {
	buff           []*Message
	lookup         map[truncatedMessageId]*Message
	oldest, newest int
	mux            sync.RWMutex
	kv             *versioned.KV
}

// NewBuff initializes a new ring buffer with size n.
func NewBuff(kv *versioned.KV, n int) (*Buff, error) {
	kv = kv.Prefix(ringBuffPrefix)

	// Construct object
	rb := &Buff{
		buff:   make([]*Message, n),
		lookup: make(map[truncatedMessageId]*Message, n),
		oldest: 0,
		newest: -1, // fixme: does this need to be neg, should this be a uint32 to match sequential message ID??
		kv:     kv,
	}

	// Save to storage and return
	return rb, rb.save()
}

// GetByMessageId looks up and returns the message with
// MessageId id from Buff.lookup. If the message does not exist,
// an error is returned.
func (rb *Buff) GetByMessageId(id MessageId) (*Message, error) {
	rb.mux.RLock()
	defer rb.mux.RUnlock()

	// Look up message
	msg, exists := rb.lookup[id.Truncate()]
	if !exists { // If message not found, return an error
		return nil, errors.Errorf(noMessageFoundErr, id)
	}

	// Return message if found
	return msg, nil
}

func (rb *Buff) GetNextMessage(id MessageId) (*Message, error) {
	rb.mux.RLock()
	defer rb.mux.RUnlock()

	// Look up message
	msg, exists := rb.lookup[id.Truncate()]
	if !exists { // If message not found, return an error
		return nil, errors.Errorf(noMessageFoundErr, id)
	}

	lookupId := msg.Id + 1

	// Check it's not before our first known id
	if lookupId < rb.oldest {
		return nil, errors.Errorf("requested ID %d is lower than oldest id %d", id, rb.oldest)
	}

	// Check it's not after our last known id
	if id > rb.newest {
		return nil, errors.Errorf("requested id %d is higher than most recent id %d", id, rb.newest)
	}

	return rb.buff[id%len(rb.buff)], nil

}

////////////////////////////////////////////////////////////////////////////////
// Storage Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// LoadBuff loads the ring buffer from storage. It loads all
// messages from storage and repopulates the buffer.
func LoadBuff(kv *versioned.KV) (*Buff, error) {
	kv = kv.Prefix(ringBuffPrefix)

	// Extract ring buffer from storage
	vo, err := kv.Get(ringBuffKey, ringBuffVersion)
	if err != nil {
		return nil, errors.Errorf(loadBuffErr, err)
	}

	// Unmarshal ring buffer from data
	newest, oldest, list := unmarshal(vo.Data)

	// Construct buffer
	rb := &Buff{
		buff:   make([]*Message, len(list)),
		lookup: make(map[truncatedMessageId]*Message, len(list)),
		oldest: oldest,
		newest: newest,
		mux:    sync.RWMutex{},
		kv:     kv,
	}

	// Load each message from storage
	for i, tmid := range list {
		msg, err := loadMessage(tmid, kv)
		if err != nil {
			return nil, err
		}

		// Place message into reconstructed buffer (RAM)
		rb.lookup[tmid] = msg
		rb.buff[i] = msg
	}

	return rb, nil
}

// save stores the ring buffer and its elements to storage.
func (rb *Buff) save() error {
	rb.mux.Lock()
	defer rb.mux.Unlock()

	// Save each message individually to storage
	for _, msg := range rb.buff {
		if err := rb.saveMessage(msg); err != nil {
			return errors.Errorf(saveMessageErr,
				msg.MessageId, err)
		}
	}

	return rb.saveBuff()
}

// saveBuff is a function which saves the marshalled Buff.
func (rb *Buff) saveBuff() error {
	obj := &versioned.Object{
		Version:   ringBuffVersion,
		Timestamp: netTime.Now(),
		Data:      rb.marshal(),
	}

	return rb.kv.Set(ringBuffKey, ringBuffVersion, obj)

}

// marshal creates a byte buffer containing serialized information
// on the Buff.
func (rb *Buff) marshal() []byte {
	// Create buffer of proper size
	// (newest (4 bytes) + oldest (4 bytes) +
	// (TruncatedMessageIdLen * length of buffer)
	buff := bytes.NewBuffer(nil)
	buff.Grow(4 + 4 + (TruncatedMessageIdLen * len(rb.lookup)))

	// Write newest index into buffer
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(rb.newest))
	buff.Write(b)

	// Write oldest index into buffer
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(rb.oldest))
	buff.Write(b)

	// Write the truncated message IDs into buffer
	for _, msg := range rb.buff {
		buff.Write(msg.MessageId.Truncate().Bytes())
	}

	return buff.Bytes()
}

// saveMessage saves a Message to storage, using the truncatedMessageId
// as the KV key.
func (rb *Buff) saveMessage(msg *Message) error {
	obj := &versioned.Object{
		Version:   messageVersion,
		Timestamp: netTime.Now(),
		Data:      rb.marshal(),
	}

	return rb.kv.Set(
		makeMessageKey(msg.MessageId.Truncate()), messageVersion, obj)

}

// unmarshal unmarshalls a byte slice into Buff information.
func unmarshal(b []byte) (newest, oldest int,
	list []truncatedMessageId) {
	buff := bytes.NewBuffer(b)

	// Read the newest index from the buffer
	newest = int(binary.LittleEndian.Uint32(buff.Next(4)))

	// Read the oldest index from the buffer
	oldest = int(binary.LittleEndian.Uint32(buff.Next(4)))

	// Initialize list to the number of truncated IDs
	list = make([]truncatedMessageId, 0, buff.Len()/TruncatedMessageIdLen)

	// Read each truncatedMessageId and save into list
	for next := buff.Next(TruncatedMessageIdLen); len(next) == TruncatedMessageIdLen; next = buff.Next(TruncatedMessageIdLen) {
		tmid := truncatedMessageId{}
		copy(tmid[:], next)
		list = append(list, tmid)
	}

	return
}

// makeMessageKey generates te key used to save a message to storage.
func makeMessageKey(tmid truncatedMessageId) string {
	return messageKey + tmid.String()
}
