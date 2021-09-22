////////////////////////////////////////////////////////////////////////////////
// Copyright © 2021 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// this is an in memory track of rounds that have been processed in this run of the
// xxdk. It only is enabled when loglevel is debug or higher. It will accumulate all
// rounds and then dump on exist. Is only enabled when run though the command line
// interface unless enabled explicitly in code.

package network

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/xx_network/primitives/id"
	"sort"
	"sync"
)

type RoundState uint8

const (
	Unchecked = iota
	Unknown
	NoMessageAvailable
	MessageAvailable
	Abandoned
)

func (rs RoundState) String() string {
	switch rs {
	case Unchecked:
		return "Unchecked"
	case Unknown:
		return "Unknown"
	case MessageAvailable:
		return "Message Available"
	case NoMessageAvailable:
		return "No Message Available"
	case Abandoned:
		return "Abandoned"
	default:
		return fmt.Sprintf("Unregistered Round State: %d", rs)
	}
}

type RoundTracker struct {
	state map[id.Round]RoundState
	mux   sync.Mutex
}

func NewRoundTracker() *RoundTracker {
	return &RoundTracker{
		state: make(map[id.Round]RoundState),
	}
}

func (rt *RoundTracker) denote(rid id.Round, state RoundState) {
	rt.mux.Lock()
	defer rt.mux.Unlock()
	if storedState, exists := rt.state[rid]; exists || storedState > state {
		return
	}
	rt.state[rid] = state
}

func (rt *RoundTracker) String() string {
	rt.mux.Lock()
	defer rt.mux.Unlock()
	jww.DEBUG.Printf("Debug Printing status of %d rounds", len(rt.state))
	keys := make([]int, 0, len(rt.state))
	for key := range rt.state {
		keys = append(keys, int(key))
	}

	sort.Ints(keys)

	stringification := ""
	for _, key := range keys {
		stringification += fmt.Sprintf("Round: %d, state:%s \n", key, rt.state[id.Round(key)])
	}

	return stringification
}
