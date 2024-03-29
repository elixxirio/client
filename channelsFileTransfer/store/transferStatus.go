////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package store

import (
	"strconv"
)

// TransferStatus indicates the state of the transfer.
type TransferStatus int

const (
	// Running indicates that the transfer is in the processes of sending
	Running TransferStatus = iota

	// Completed indicates that all file parts have been sent and arrived
	Completed

	// Failed indicates that the transfer has run out of sending retries
	Failed
)

// String prints the string representation of the TransferStatus. This function
// satisfies the fmt.Stringer interface.
func (ts TransferStatus) String() string {
	switch ts {
	case Running:
		return "running"
	case Completed:
		return "completed"
	case Failed:
		return "failed"
	default:
		return "INVALID STATUS: " + strconv.Itoa(int(ts))
	}
}
