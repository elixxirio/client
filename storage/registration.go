////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////
package storage

import (
	"encoding/json"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/client/storage/versioned"
	"time"
)

var currentRegistrationVersion = uint64(0)

// SetRegValidationSig builds the versioned object and sets it in the
// key-value store
func (s *Session) SetRegValidationSig(newVal []byte) error {
	// Construct the versioned object
	vo := &versioned.Object{
		Version:   currentRegistrationVersion,
		Timestamp: time.Now(),
		Data:      newVal,
	}

	// Construct the key and place in the key-value store
	key := "RegValidationSig"

	return s.kv.Set(key, vo)
}

// GetRegValidationSig pulls the versioned object by the key and parses
// it into the requested registration signature
func (s *Session) GetRegValidationSig() ([]byte, error) {
	key := "RegValidationSig"

	// Pull the object from the key-value store
	voData, err := s.kv.Get(key)
	if err != nil {
		return nil, err
	}

	if voData.Version != currentRegistrationVersion {
		globals.Log.WARN.Printf("Session.GetRegValidationSig: got "+
			"unexpected version %v, expected version %v",
			voData.Version, currentRegistrationVersion)
	}

	return voData.Data, nil
}

// SetRegState uses the SetInterface method to place the regstate into
// the key-value store
func (s *Session) SetRegState(newVal int64) error {
	key := "RegState"

	data, err := json.Marshal(newVal)
	if err != nil {
		return err
	}

	obj := versioned.Object{
		Version:   currentRegistrationVersion,
		Timestamp: time.Now(),
		Data:      data,
	}

	return s.kv.Set(key, &obj)
}

// GetRegValidationSig pulls the versioned object by the key and parses
// it into the requested registration signature
func (s *Session) GetRegState() (int64, error) {
	// Construct the key from the
	key := "RegState"

	// Pull the object from the key-value store
	voData, err := s.kv.Get(key)
	if err != nil {
		return 0, err
	}

	if voData.Version != currentRegistrationVersion {
		globals.Log.WARN.Printf("Session.GetRegState: got unexpected "+
			"version %v, expected version %v",
			voData.Version, currentRegistrationVersion)
	}

	var data int64
	err = json.Unmarshal(voData.Data, &data)
	if err != nil {
		return 0, err
	}

	return data, nil

}
