////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package user

import (
	"crypto/sha256"
	"encoding/binary"
	"gitlab.com/privategrity/client/globals"
	"gitlab.com/privategrity/crypto/cyclic"
	"gitlab.com/privategrity/crypto/hash"
	"strconv"
)

// TODO use this type for User IDs consistently throughout
// FIXME use string or []byte for this - string works as a key for hash maps
// and []byte is compatible with more other languages.
// string probably makes more sense
type ID uint64

var IDLen = 8

// TODO remove this when ID becomes a string
func (u ID) Bytes() []byte {
	result := make([]byte, IDLen)
	binary.BigEndian.PutUint64(result, uint64(u))
	return result
}

// TODO clean this up
func (u ID) RegistrationCode() string {
	return cyclic.NewIntFromUInt(uint64(NewIDFromBytes(UserHash(u)))).TextVerbose(32, 0)
}

func NewIDFromBytes(id []byte) ID {
	// to keep compatibility with old user registration codes, we need to use
	// the last part of the byte array that we pass in
	// FIXME break compatibility here during the migration to 128 bit ids
	result := ID(binary.BigEndian.Uint64(id[len(id)-IDLen:]))
	return result
}

// Converts from human-readable string to user ID
// NOTE This will break when we migrate to the new 128-bit user IDs
func NewIDFromString(id string, base int) (ID, error) {
	newID, err := strconv.ParseUint(id, 10, 64)
	return ID(newID), err
}

// Globally instantiated Registry
var Users = newRegistry()
var NUM_DEMO_USERS = int(40)
var DEMO_USER_NICKS = []string{"David", "Jim", "Ben", "Rick", "Spencer", "Jake",
	"Mario", "Will", "Allan", "Jono", "", "", "UDB", "", "", "", "Payments"}
var DEMO_CHANNEL_NAMES = []string{"#General", "#Engineering", "#Lunch",
	"#Random"}

// Interface for User Registry operations
type Registry interface {
	NewUser(id ID, nickname string) *User
	DeleteUser(id ID)
	GetUser(id ID) (user *User, ok bool)
	UpsertUser(user *User)
	CountUsers() int
	LookupUser(hid string) (uid ID, ok bool)
	LookupKeys(uid ID) (*NodeKeys, bool)
	GetContactList() ([]ID, []string)
}

type UserMap struct {
	// Map acting as the User Registry containing User -> ID mapping
	userCollection map[ID]*User
	// Increments sequentially for User.ID values
	idCounter uint64
	// Temporary map acting as a lookup table for demo user registration codes
	// Key type is string because keys must implement == and []byte doesn't
	userLookup map[string]ID
	//Temporary placed to store the keys for each user
	keysLookup map[ID]*NodeKeys
}

// newRegistry creates a new Registry interface
func newRegistry() Registry {
	if len(DEMO_CHANNEL_NAMES) > 10 || len(DEMO_USER_NICKS) > 30 {
		globals.Log.ERROR.Print("Not enough demo users have been hardcoded.")
	}
	uc := make(map[ID]*User)
	ul := make(map[string]ID)
	nk := make(map[ID]*NodeKeys)

	// Deterministically create NUM_DEMO_USERS users
	for i := 1; i <= NUM_DEMO_USERS; i++ {
		t := new(User)
		k := new(NodeKeys)

		// Generate user parameters
		t.UserID = ID(i)
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
		uc[t.UserID] = t
		ul[string(UserHash(t.UserID))] = t.UserID
		nk[t.UserID] = k
	}

	// Channels have been hardcoded to users 101-200
	for i := 0; i < len(DEMO_USER_NICKS); i++ {
		uc[ID(i+1)].Nick = DEMO_USER_NICKS[i]
	}
	for i := 0; i < len(DEMO_CHANNEL_NAMES); i++ {
		uc[ID(i+31)].Nick = DEMO_CHANNEL_NAMES[i]
	}

	// With an underlying UserMap data structure
	return Registry(&UserMap{userCollection: uc,
		idCounter:  uint64(NUM_DEMO_USERS),
		userLookup: ul,
		keysLookup: nk})
}

// Struct representing a User in the system
type User struct {
	UserID ID
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

// UserHash generates a hash of the UID to be used as a registration code for
// demos
// TODO Should we use the full-length hash? Should we even be doing registration
// like this?
func UserHash(uid ID) []byte {
	h, _ := hash.NewCMixHash()
	h.Write(uid.Bytes())
	huid := h.Sum(nil)
	huid = huid[len(huid)-IDLen:]
	return huid
}

// NewUser creates a new User object with default fields and given address.
func (m *UserMap) NewUser(id ID, nickname string) *User {
	return &User{UserID: id, Nick: nickname}
}

// GetUser returns a user with the given ID from userCollection
// and a boolean for whether the user exists
func (m *UserMap) GetUser(id ID) (user *User, ok bool) {
	user, ok = m.userCollection[id]
	user = user.DeepCopy()
	return
}

// DeleteUser deletes a user with the given ID from userCollection.
func (m *UserMap) DeleteUser(id ID) {
	// If key does not exist, do nothing
	delete(m.userCollection, id)
}

// UpsertUser inserts given user into userCollection or update the user if it
// already exists (Upsert operation).
func (m *UserMap) UpsertUser(user *User) {
	m.userCollection[user.UserID] = user
}

// CountUsers returns a count of the users in userCollection
func (m *UserMap) CountUsers() int {
	return len(m.userCollection)
}

// LookupUser returns the user id corresponding to the demo registration code
func (m *UserMap) LookupUser(hid string) (uid ID, ok bool) {
	uid, ok = m.userLookup[hid]
	return
}

// LookupKeys returns the keys for the given user from the temporary key map
func (m *UserMap) LookupKeys(uid ID) (*NodeKeys, bool) {
	nk, t := m.keysLookup[uid]
	return nk, t
}

func (m *UserMap) GetContactList() (ids []ID, nicks []string) {
	ids = make([]ID, len(m.userCollection))
	nicks = make([]string, len(m.userCollection))

	index := uint64(0)
	for _, user := range m.userCollection {
		ids[index] = user.UserID
		nicks[index] = user.Nick
		index++
	}
	return ids, nicks
}
