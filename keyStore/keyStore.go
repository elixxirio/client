package keyStore

import (
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/elixxir/primitives/id"
	"sync"
)

// Local types in order to implement functions that
// return real types instead of interfaces
type outKeyMap sync.Map
type inKeyMap sync.Map

// Transmission Keys map
// Maps id.User to *KeyStack
var TransmissionKeys = new(outKeyMap)

// Transmission ReKeys map
// Maps id.User to *KeyStack
var TransmissionReKeys = new(outKeyMap)

// Receiption Keys map
// Maps format.Fingerprint to *E2EKey
var ReceptionKeys = new(inKeyMap)

// Stores a KeyStack entry for given user
func (m *outKeyMap) Store(user *id.User, keys *KeyStack) {
	(*sync.Map)(m).Store(*user, keys)
}

// Pops first key from KeyStack for given user
// Atomically updates Key Manager Sending state
// Returns *E2EKey and KeyAction
func (m *outKeyMap) Pop(user *id.User) (*E2EKey, KeyAction) {
	val, ok := (*sync.Map)(m).Load(*user)

	if !ok {
		return nil, None
	}

	keyStack := val.(*KeyStack)

	// Pop key
	e2eKey := keyStack.Pop()
	// Update Key Manager State
	action := e2eKey.GetManager().updateState(e2eKey.GetOuterType() == format.Rekey)
	return e2eKey, action
}

// Deletes a KeyStack entry for given user
func (m *outKeyMap) Delete(user *id.User) {
	(*sync.Map)(m).Delete(*user)
}

// Stores an *E2EKey for given fingerprint
func (m *inKeyMap) Store(fingerprint format.Fingerprint, key *E2EKey) {
	(*sync.Map)(m).Store(fingerprint, key)
}

// Pops key for given fingerprint, i.e,
// returns and deletes it from the map
// Atomically updates Key Manager Receiving state
// Returns nil if not found
func (m *inKeyMap) Pop(fingerprint format.Fingerprint) *E2EKey {
	val, ok := (*sync.Map)(m).Load(fingerprint)

	var key *E2EKey
	if !ok {
		return nil
	} else {
		key = val.(*E2EKey)
	}
	// Delete key from map
	m.Delete(fingerprint)
	// Update Key Manager Receiving State
	key.GetManager().updateRecvState(
		key.GetOuterType() == format.Rekey,
		key.keyNum)
	return key
}

// Deletes a key for given fingerprint
func (m *inKeyMap) Delete(fingerprint format.Fingerprint) {
	(*sync.Map)(m).Delete(fingerprint)
}

// Deletes keys from a given list of fingerprints
func (m *inKeyMap) DeleteList(fingerprints []format.Fingerprint) {
	for _, fp := range fingerprints {
		m.Delete(fp)
	}
}
