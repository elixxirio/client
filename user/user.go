////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package user

import (
	"crypto/sha256"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/primitives/userid"
)

// Globally instantiated Registry
var Users = newRegistry()

const NUM_DEMO_USERS = 40

var DemoUserNicks = []string{"David", "Jim", "Ben", "Rick", "Spencer", "Jake",
	"Mario", "Will", "Allan", "Jono", "", "", "UDB", "", "", "", "Payments"}
var DemoChannelNames = []string{"#General", "#Engineering", "#Lunch",
	"#Random"}

// Interface for User Registry operations
type Registry interface {
	NewUser(id *id.UserID, nickname string) *User
	DeleteUser(id *id.UserID)
	GetUser(id *id.UserID) (user *User, ok bool)
	UpsertUser(user *User)
	CountUsers() int
	LookupUser(hid string) (uid *id.UserID, ok bool)
	LookupKeys(uid *id.UserID) (*NodeKeys, bool)
}

type UserMap struct {
	// Map acting as the User Registry containing User -> ID mapping
	userCollection map[id.UserID]*User
	// Increments sequentially for User.ID values
	idCounter uint64
	// Temporary map acting as a lookup table for demo user registration codes
	// Key type is string because keys must implement == and []byte doesn't
	userLookup map[string]*id.UserID
	//Temporary placed to store the keys for each user
	keysLookup map[id.UserID]*NodeKeys
}

// newRegistry creates a new Registry interface
func newRegistry() Registry {
	if len(DemoChannelNames) > 10 || len(DemoUserNicks) > 30 {
		globals.Log.ERROR.Print("Not enough demo users have been hardcoded.")
	}
	uc := make(map[id.UserID]*User)
	ul := make(map[string]*id.UserID)
	nk := make(map[id.UserID]*NodeKeys)

	// Deterministically create NUM_DEMO_USERS users
	// TODO Replace this with real user registration/discovery
	for i := uint64(1); i <= NUM_DEMO_USERS; i++ {
		currentID := new(id.UserID).SetUints(&[4]uint64{0, 0, 0, i})
		t := new(User)
		k := new(NodeKeys)

		// Generate user parameters
		t.UserID = currentID
		currentID.RegistrationCode()
		// TODO We need a better way to generate base/recursive keys
		h := sha256.New()
		h.Write([]byte(string(20000 + i)))
		k.TransmissionKeys.Base = cyclic.NewIntFromBytes(h.Sum(nil))
		h = sha256.New()
		h.Write([]byte(string(30000 + i)))
		k.TransmissionKeys.Recursive = cyclic.NewIntFromBytes(h.Sum(nil))
		h = sha256.New()
		h.Write([]byte(string(40000 + i)))
		k.ReceptionKeys.Base = cyclic.NewIntFromBytes(h.Sum(nil))
		h = sha256.New()
		h.Write([]byte(string(50000 + i)))
		k.ReceptionKeys.Recursive = cyclic.NewIntFromBytes(h.Sum(nil))

		// Add user to collection and lookup table
		uc[*t.UserID] = t
		// Detect collisions in the registration code
		if _, ok := ul[t.UserID.RegistrationCode()]; ok {
			globals.Log.ERROR.Printf(
				"Collision in demo user list creation at %v. "+
					"Please fix ASAP (include more bits to the reg code.", i)
		}
		ul[t.UserID.RegistrationCode()] = t.UserID
		nk[*t.UserID] = k
	}

	// Channels have been hardcoded to users starting with 31
	for i := 0; i < len(DemoUserNicks); i++ {
		currentID := new(id.UserID).SetUints(&[4]uint64{0, 0, 0, uint64(i) + 1})
		uc[*currentID].Nick = DemoUserNicks[i]
	}

	for i := 0; i < len(DemoChannelNames); i++ {
		currentID := new(id.UserID).SetUints(&[4]uint64{0, 0, 0, uint64(i) + 31})
		uc[*currentID].Nick = DemoChannelNames[i]
	}

	// With an underlying UserMap data structure
	return Registry(&UserMap{userCollection: uc,
		idCounter: uint64(NUM_DEMO_USERS),
		userLookup: ul,
		keysLookup: nk})
}

// Struct representing a User in the system
type User struct {
	UserID *id.UserID
	Nick   string
}

// DeepCopy performs a deep copy of a user and returns a pointer to the new copy
func (u *User) DeepCopy() *User {
	if u == nil {
		return nil
	}
	nu := new(User)
	nu.UserID = u.UserID
	nu.Nick = u.Nick
	return nu
}

// NewUser creates a new User object with default fields and given address.
func (m *UserMap) NewUser(id *id.UserID, nickname string) *User {
	return &User{UserID: id, Nick: nickname}
}

// GetUser returns a user with the given ID from userCollection
// and a boolean for whether the user exists
func (m *UserMap) GetUser(id *id.UserID) (user *User, ok bool) {
	user, ok = m.userCollection[*id]
	user = user.DeepCopy()
	return
}

// DeleteUser deletes a user with the given ID from userCollection.
func (m *UserMap) DeleteUser(id *id.UserID) {
	// If key does not exist, do nothing
	delete(m.userCollection, *id)
}

// UpsertUser inserts given user into userCollection or update the user if it
// already exists (Upsert operation).
func (m *UserMap) UpsertUser(user *User) {
	m.userCollection[*user.UserID] = user
}

// CountUsers returns a count of the users in userCollection
func (m *UserMap) CountUsers() int {
	return len(m.userCollection)
}

// LookupUser returns the user id corresponding to the demo registration code
func (m *UserMap) LookupUser(hid string) (*id.UserID, bool) {
	uid, ok := m.userLookup[hid]
	return uid, ok
}

// LookupKeys returns the keys for the given user from the temporary key map
func (m *UserMap) LookupKeys(uid *id.UserID) (*NodeKeys, bool) {
	nk, t := m.keysLookup[*uid]
	return nk, t
}
