////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package connect

import (
	"gitlab.com/elixxir/client/e2e/receive"
	"gitlab.com/elixxir/client/restlike"
	"google.golang.org/protobuf/proto"
)

// response is the response handler for a Request
type response struct {
	responseCallback restlike.RequestCallback
}

// Hear handles for connect.Connection message responses for a Request
func (r response) Hear(item receive.Message) {
	newMessage := &restlike.Message{}

	// Unmarshal the payload
	err := proto.Unmarshal(item.Payload, newMessage)
	if err != nil {
		newMessage.Error = err.Error()
	}

	// Send the response payload to the responseCallback
	r.responseCallback(newMessage)
}

// Name is used for debugging
func (r response) Name() string {
	return "Restlike"
}
