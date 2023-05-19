////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/notifications"
	"sync"
)

// notifTrackerSingleton is used to track notifications objects so that they
// can be referenced by ID back over the bindings.
var notifTrackerSingleton = &notificationsTracker{
	tracked: make(map[int]notifWrapper),
	count:   0,
}

type Notifications interface {
	// AddToken registers the Token with the remote server if this manager is
	// in set to register, otherwise it will return ErrRemoteRegistrationDisabled
	// This will add the token to the list of tokens which are forwarded the messages
	// for connected IDs.
	// the App will tell the server what App to forward the notifications to.
	AddToken(newToken, app string) error
	// RemoveToken removes the given Token from the server
	// It will remove all registered identities if it is the last Token

	RemoveToken() error
	// SetMaxState sets the maximum functional state of any identity
	// downstream moduals will be told to clamp any state greater than maxState
	// down to maxState. Depending on UX requirements, they may still show the
	// state in an altered manner, for example greying out a description.
	// This is designed so when the state is raised, the old configs are
	// maintained.
	// This will unregister / re-register with the push server when leaving or
	// entering the Push maxState.
	// The default maxState is Push
	// will return an error if the maxState isnt a valid state
	SetMaxState(maxState notifications.NotificationState) error

	// GetMaxState returns the current MaxState
	GetMaxState() notifications.NotificationState

	// GetID returns the ID of the notifications object
	GetID() int
}

type notifWrapper struct {
	notifications.Manager
	id int
}

func (n notifWrapper) GetID() int {
	return n.id
}

func LoadNotifications(cmixId int) (Notifications, error) {
	mixBind, err := cmixTrackerSingleton.get(cmixId)
	if err != nil {
		return nil, err
	}
	mix := mixBind.api
	identity := mix.GetTransmissionIdentity()
	sig := mix.GetStorage().GetReceptionRegistrationValidationSignature()
	kv := mix.GetStorage().GetKV()
	comms := mix.GetComms()
	rng := mix.GetRng()

	notif := notifications.NewOrLoadManager(identity, sig, kv, comms, rng)

	return notifTrackerSingleton.make(notif), nil
}

////////////////////////////////////////////////////////////////////////////////
// Notifications Tracker                                                      //
////////////////////////////////////////////////////////////////////////////////

// notificationsTracker is a singleton used to keep track of extant notifications
// objects, preventing race conditions created by passing it over the bindings.
type notificationsTracker struct {
	// TODO: Key on Identity.ID to prevent duplication
	tracked map[int]notifWrapper
	count   int
	mux     sync.RWMutex
}

// make create an E2e from a xxdk.E2e, assigns it a unique ID, and adds it to
// the e2eTracker.
func (nt *notificationsTracker) make(nm notifications.Manager) Notifications {
	nt.mux.Lock()
	defer nt.mux.Unlock()

	notifID := nt.count
	nt.count++

	nw := notifWrapper{
		Manager: nm,
		id:      notifID,
	}

	nt.tracked[notifID] = nw

	return nw
}

// get a notifWrapper from the notificationsTracker given its ID.
func (nt *notificationsTracker) get(id int) (notifWrapper, error) {
	nt.mux.RLock()
	defer nt.mux.RUnlock()

	c, exist := nt.tracked[id]
	if !exist {
		return notifWrapper{}, errors.Errorf("Cannot get Notifications"+
			" for ID %d, does not exist", id)
	}

	return c, nil
}

// delete a notifWrapper from the notificationsTracker.
func (nt *notificationsTracker) delete(id int) {
	nt.mux.Lock()
	defer nt.mux.Unlock()

	delete(nt.tracked, id)
}
