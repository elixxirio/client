////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package fileTransfer

import (
	"reflect"
	"testing"
)

// Tests that NewParams returns the expected Params object.
func TestNewParams(t *testing.T) {
	expected := Params{
		MaxThroughput: 42,
	}

	received := NewParams(expected.MaxThroughput)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Received Params does not match expected."+
			"\nexpected: %+v\nreceived: %+v", expected, received)
	}
}

// Tests that DefaultParams returns a Params object with the expected defaults.
func TestDefaultParams(t *testing.T) {
	expected := Params{MaxThroughput: defaultMaxThroughput}
	received := DefaultParams()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Received Params does not match expected."+
			"\nexpected: %+v\nreceived: %+v", expected, received)
	}
}
