////////////////////////////////////////////////////////////////////////////////
// Copyright © 2019 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Package io asynchronous sending functionality. This is managed by an outgoing
// messages channel and managed by the sender thread kicked off during
// initialization.
package io

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/client/parse"
	"gitlab.com/elixxir/comms/client"
	"sync"
	"time"
)

const PermissioningAddrID = "Permissioning"

type ConnAddr string

func (a ConnAddr) String() string {
	return string(a)
}

// ReceptionManager implements the Communications interface
type ReceptionManager struct {
	// Comms pointer to send/recv messages
	Comms *client.Comms

	nextId   func() []byte
	collator *Collator

	//Flags if the network is using tls or note
	Tls bool
	// blockTransmissions will use a mutex to prevent multiple threads from sending
	// messages at the same time.
	blockTransmissions bool // pass into receiver
	// transmitDelay is the minimum delay between transmissions.
	transmitDelay time.Duration // same
	// Map that holds a record of the messages that this client successfully
	// received during this session
	receivedMessages   map[string]struct{}
	recievedMesageLock sync.RWMutex

	sendLock sync.Mutex

	rekeyChan chan struct{}
}

func NewReceptionManager(rekeyChan chan struct{}, privKey, pubKey, salt []byte) *ReceptionManager {
	comms, err := client.NewClientComms("client", pubKey, privKey, salt)
	if comms == nil {
		globals.Log.ERROR.Printf("Failed to get client comms using constructor: %+v", err)
		comms = &client.Comms{}
	}

	cm := &ReceptionManager{
		nextId:             parse.IDCounter(),
		collator:           NewCollator(),
		blockTransmissions: true,
		transmitDelay:      1000 * time.Millisecond,
		receivedMessages:   make(map[string]struct{}),
		Comms:              comms,
		rekeyChan:          rekeyChan,
		Tls:                true,
	}

	return cm
}

// Connects to the permissioning server, if we know about it, to get the latest
// version from it
func (rm *ReceptionManager) GetRemoteVersion() (string, error) {
	permissioningHost, ok := rm.Comms.GetHost(PermissioningAddrID)
	if !ok {
		return "", errors.Errorf("Failed to find permissioning host with id %s", PermissioningAddrID)
	}
	registrationVersion, err := rm.Comms.
		SendGetCurrentClientVersionMessage(permissioningHost)
	if err != nil {
		return "", errors.Wrap(err, "Couldn't get current version from permissioning")
	}
	return registrationVersion.Version, nil
}

func (rm *ReceptionManager) DisableBlockingTransmission() { // flag passed into receiver
	rm.blockTransmissions = false
}

func (rm *ReceptionManager) SetRateLimit(delay time.Duration) { // pass into received
	rm.transmitDelay = delay
}
