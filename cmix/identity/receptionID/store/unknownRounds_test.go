////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package store

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/xx_network/primitives/id"
	"reflect"
	"sync/atomic"
	"testing"
)

// Happy path
func TestNewUnknownRounds(t *testing.T) {
	kv := versioned.NewKV(ekv.MakeMemstore())
	expectedKv, err := kv.Prefix(unknownRoundPrefix)
	require.NoError(t, err)
	expectedStore := &UnknownRounds{
		rounds: make(map[id.Round]*uint64),
		kv:     expectedKv,
		params: DefaultUnknownRoundsParams(),
	}

	store := NewUnknownRounds(kv, DefaultUnknownRoundsParams())

	// Compare manually created object with NewUnknownRounds
	if !reflect.DeepEqual(expectedStore, store) {
		t.Errorf("NewUnknownRounds returned incorrect Store."+
			"\nexpected: %+v\nreceived: %+v", expectedStore, store)
	}

	if err := store.save(); err != nil {
		t.Fatalf("save could not write to disk: %v", err)
	}

	expectedData, err := json.Marshal(store.rounds)
	if err != nil {
		t.Fatalf("json.Marshal produced an error: %v", err)
	}

	key, err := store.kv.Get(unknownRoundsStorageKey, unknownRoundsStorageVersion)
	if err != nil {
		t.Fatalf("get encoutnered an error when getting Store from KV: %v", err)
	}

	// Check that the stored data is the data outputted by marshal
	if !bytes.Equal(expectedData, key.Data) {
		t.Errorf("NewUnknownRounds returned incorrect Store."+
			"\nexpected: %+v\nreceived: %+v", expectedData, key.Data)
	}

}

// Full test.
func TestUnknownRounds_Iterate(t *testing.T) {
	kv := versioned.NewKV(ekv.MakeMemstore())
	store := NewUnknownRounds(kv, DefaultUnknownRoundsParams())

	// Return true only for rounds that are even
	mockChecker := func(rid id.Round) bool {
		return uint64(rid)%2 == 0
	}

	// Construct 3 lists of round IDs
	roundListLen := 25
	unknownRounds := make([]id.Round, roundListLen)
	roundListEven := make([]id.Round, roundListLen)
	for i := 0; i < roundListLen; i++ {
		// Will contain a list of round Ids in range [0, 25)
		unknownRounds[i] = id.Round(i)
		// Will contain even round ID's in range [50, 100)
		roundListEven[i] = id.Round((i + roundListLen) * 2)

	}

	// Add unknown rounds to map
	for _, rnd := range unknownRounds {
		roundVal := uint64(0)
		store.rounds[rnd] = &roundVal
	}

	// Iterate over initial map
	received := store.Iterate(mockChecker, nil, func(round id.Round) { return })

	// Check the received list for 2 conditions:
	// a) that returned rounds are no longer in the map
	// b) that returned round Ids are even (as per our checker)
	for _, rnd := range received {
		// Our returned list should contain only even rounds.
		if uint64(rnd)%2 != 0 {
			t.Errorf("Unexpected result from iterate. "+
				"Round %d should not be in received list", rnd)
		}

		// Elements in the returned list should be deleted from the map.
		if _, ok := store.rounds[rnd]; ok {
			t.Errorf("Returned rounds from iterate should be removed from "+
				"map. Found round %d in map", rnd)
		}

	}

	// Add even round list to map
	received = store.Iterate(mockChecker, roundListEven, func(_ id.Round) {})

	if len(received) != 0 {
		t.Errorf("Second iteration should return an empty list (no even "+
			"rounds are left).\nreturned: %v", received)
	}

	// Iterate over map until all rounds have checks incremented over maxCheck
	for i := 0; i < defaultMaxCheck+1; i++ {
		_ = store.Iterate(mockChecker, []id.Round{}, func(_ id.Round) {})

	}

	// Check map has been cleared out
	if len(store.rounds) != 0 {
		t.Errorf("Map should be empty after %d iterations", defaultMaxCheck)
	}
}

// Unit test
func TestLoadUnknownRounds(t *testing.T) {
	kv := versioned.NewKV(ekv.MakeMemstore())
	store := NewUnknownRounds(kv, DefaultUnknownRoundsParams())

	// Construct 3 lists of round IDs
	roundListLen := 25
	expectedRounds := make([]id.Round, roundListLen)
	for i := 0; i < roundListLen; i++ {
		// Will contain a list of round IDs in range [0, 25)
		expectedRounds[i] = id.Round(i)

	}

	// Add unknown rounds to map
	expectedCheckVal := 0
	for _, rnd := range expectedRounds {
		roundVal := uint64(expectedCheckVal)
		store.rounds[rnd] = &roundVal
	}

	if err := store.save(); err != nil {
		t.Fatalf("save produced an error: %v", err)
	}

	// Load the store from kv
	receivedStore := LoadUnknownRounds(kv, DefaultUnknownRoundsParams())

	// Check the state of the map of the loaded store
	for _, rnd := range expectedRounds {
		check, ok := receivedStore.rounds[rnd]
		if !ok {
			t.Fatalf("Expected round %d in loaded store", rnd)
		}

		if atomic.LoadUint64(check) != 0 {
			t.Fatalf("Loaded value in map is unexpected."+
				"\nexpected: %v\nreceived: %v",
				expectedCheckVal, atomic.LoadUint64(check))
		}
	}

	// Check save used in iterate call

	// Check that LoadStore works after iterate call (which implicitly saves)
	mockChecker := func(round id.Round) bool { return false }
	received := store.Iterate(mockChecker, nil, func(_ id.Round) {})

	// Iterate is being called as a dummy, should not return anything
	if len(received) != 0 {
		t.Fatalf("Returned list from iterate should not return any rounds."+
			"\nreceived: %v", received)
	}

	// Increment check value (iterate will increment all rounds' checked value)
	expectedCheckVal++

	// Load the store from kv
	receivedStore = LoadUnknownRounds(kv, DefaultUnknownRoundsParams())

	// Check the state of the map of the loaded store
	for _, rnd := range expectedRounds {
		check, ok := receivedStore.rounds[rnd]
		if !ok {
			t.Fatalf("Expected round %d in loaded store", rnd)
		}

		if atomic.LoadUint64(check) != uint64(expectedCheckVal) {
			t.Fatalf("Loaded value in map is unexpected."+
				"\nexpected: %v\nreceived: %v",
				expectedCheckVal, atomic.LoadUint64(check))
		}
	}

}
