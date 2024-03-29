////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package storage

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/client/v4/storage/utility"
	"gitlab.com/xx_network/primitives/ndf"
)

const ndfKey = "ndf"

func (s *session) SetNDF(def *ndf.NetworkDefinition) {
	err := SaveNDF(s.kv, def)
	if err != nil {
		jww.FATAL.Printf("Failed to dave the NDF: %+v", err)
	}
	s.ndf = def
}

func (s *session) GetNDF() *ndf.NetworkDefinition {
	if s.ndf != nil {
		return s.ndf
	}
	def, err := utility.LoadNDF(s.kv, ndfKey)
	if err != nil {
		jww.FATAL.Printf("Could not load the NDF: %+v", err)
	}

	s.ndf = def
	return def
}

func SaveNDF(kv versioned.KV, def *ndf.NetworkDefinition) error {
	return utility.SaveNDF(kv, ndfKey, def)
}
