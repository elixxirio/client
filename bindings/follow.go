///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/xx_network/primitives/netTime"
)

// StartNetworkFollower kicks off the tracking of the network. It starts long-
// running network client threads and returns an object for checking state and
// stopping those threads.
//
// Call this when returning from sleep and close when going back to sleep.
//
// These threads may become a significant drain on battery when offline, ensure
// they are stopped if there is no internet access.
//
// Threads Started:
//   - Network Follower (/network/follow.go)
//   	tracks the network events and hands them off to workers for handling.
//   - Historical Round Retrieval (/network/rounds/historical.go)
// 		retrieves data about rounds that are too old to be stored by the client.
//	 - Message Retrieval Worker Group (/network/rounds/retrieve.go)
//		requests all messages in a given round from the gateway of the last
//		nodes.
//	 - Message Handling Worker Group (/network/message/handle.go)
//		decrypts and partitions messages when signals via the Switchboard.
//	 - Health Tracker (/network/health),
//		via the network instance, tracks the state of the network.
//	 - Garbled Messages (/network/message/garbled.go)
//		can be signaled to check all recent messages that could be decoded. It
//		uses a message store on disk for persistence.
//	 - Critical Messages (/network/message/critical.go)
//		ensures all protocol layer mandatory messages are sent. It uses a
//		message store on disk for persistence.
//	 - KeyExchange Trigger (/keyExchange/trigger.go)
//		responds to sent rekeys and executes them.
//   - KeyExchange Confirm (/keyExchange/confirm.go)
//		responds to confirmations of successful rekey operations.
//   - Auth Callback (/auth/callback.go)
//      handles both auth confirm and requests.
func (c *Cmix) StartNetworkFollower(timeoutMS int) error {
	timeout := time.Duration(timeoutMS) * time.Millisecond
	return c.api.StartNetworkFollower(timeout)
}

// StopNetworkFollower stops the network follower if it is running. It returns
// an error if the follower is in the wrong state to stop or if it fails to stop
// it.
//
// if the network follower is running and this fails, the client object will
// most likely be in an unrecoverable state and need to be trashed.
func (c *Cmix) StopNetworkFollower() error {
	if err := c.api.StopNetworkFollower(); err != nil {
		return errors.New(fmt.Sprintf("Failed to stop the "+
			"network follower: %+v", err))
	}
	return nil
}

// WaitForNetwork will block until either the network is healthy or the passed
// timeout is reached. It will return true if the network is healthy.
func (c *Cmix) WaitForNetwork(timeoutMS int) bool {
	start := netTime.Now()
	timeout := time.Duration(timeoutMS) * time.Millisecond
	for netTime.Since(start) < timeout {
		if c.api.GetCmix().IsHealthy() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// NetworkFollowerStatus gets the state of the network follower. It returns a
// status with the following values:
//  Stopped  - 0
//  Running  - 2000
//  Stopping - 3000
func (c *Cmix) NetworkFollowerStatus() int {
	return int(c.api.NetworkFollowerStatus())
}

// HasRunningProcessies checks if any background threads are running and returns
// true if one or more are.
//
// This is meant to be used when NetworkFollowerStatus returns xxdk.Stopping.
// Due to the handling of comms on iOS, where the OS can block indefinitely, it
// may not enter the stopped state appropriately. This can be used instead.
func (c *Cmix) HasRunningProcessies() bool {
	return c.api.HasRunningProcessies()
}

// IsHealthy returns true if the network is read to be in a healthy state where
// messages can be sent.
func (c *Cmix) IsHealthy() bool {
	return c.api.GetCmix().IsHealthy()
}

// NetworkHealthCallback contains a callback that is used to receive
// notification if network health changes.
type NetworkHealthCallback interface {
	Callback(bool)
}

// AddHealthCallback adds a callback that gets called whenever the network
// health changes. Returns a registration ID that can be used to unregister.
func (c *Cmix) AddHealthCallback(nhc NetworkHealthCallback) int64 {
	return int64(c.api.GetCmix().AddHealthCallback(nhc.Callback))
}

// RemoveHealthCallback removes a health callback using its registration ID.
func (c *Cmix) RemoveHealthCallback(funcID int64) {
	c.api.GetCmix().RemoveHealthCallback(uint64(funcID))
}

type ClientError interface {
	Report(source, message, trace string)
}

// RegisterClientErrorCallback registers the callback to handle errors from the
// long-running threads controlled by StartNetworkFollower and
// StopNetworkFollower.
func (c *Cmix) RegisterClientErrorCallback(clientError ClientError) {
	errChan := c.api.GetErrorsChannel()
	go func() {
		for report := range errChan {
			go clientError.Report(report.Source, report.Message, report.Trace)
		}
	}()
}
