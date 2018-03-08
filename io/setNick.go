////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package io

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/privategrity/comms/mixclient"
	pb "gitlab.com/privategrity/comms/mixmessages"
	"gitlab.com/privategrity/client/globals"
)

func SetNick(addr string, user globals.User) {
	msg := pb.SetNickMessage
	_, err := mixclient.SetNick(addr, msg)

	if err != nil {
		jww.FATAL.Panicf("Failed to set nick: %v", err.Error())
	}
}
