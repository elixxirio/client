////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// params.go provides functions for getting and setting parameters in bindings.

package bindings

import (
	"encoding/json"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/fileTransfer"
	e2eFileTransfer "gitlab.com/elixxir/client/fileTransfer/e2e"
	"gitlab.com/elixxir/client/single"
	"gitlab.com/elixxir/client/xxdk"
)

// GetDefaultCMixParams returns a JSON serialized object with all of the cMix
// parameters and their default values. Call this function and modify the JSON
// to change cMix settings.
func GetDefaultCMixParams() []byte {
	defaultParams := xxdk.GetDefaultCMixParams()
	data, err := defaultParams.Marshal()
	if err != nil {
		jww.FATAL.Panicf("Failed to JSON marshal cMix params: %+v", err)
	}
	return data
}

// GetDefaultE2EParams returns a JSON serialized object with all of the E2E
// parameters and their default values. Call this function and modify the JSON
// to change E2E settings.
func GetDefaultE2EParams() []byte {
	defaultParams := xxdk.GetDefaultE2EParams()
	data, err := defaultParams.Marshal()
	if err != nil {
		jww.FATAL.Panicf("Failed to JSON marshal E2E params: %+v", err)
	}
	return data
}

// GetDefaultFileTransferParams returns a JSON serialized object with all the
// file transfer parameters and their default values. Call this function and
// modify the JSON to change file transfer settings.
func GetDefaultFileTransferParams() []byte {
	defaultParams := fileTransfer.DefaultParams()
	data, err := json.Marshal(defaultParams)
	if err != nil {
		jww.FATAL.Panicf("Failed to JSON marshal file transfer params: %+v", err)
	}
	return data
}

// GetDefaultSingleUseParams returns a JSON serialized object with all the
// single-use parameters and their default values. Call this function and modify
// the JSON to change single use settings.
func GetDefaultSingleUseParams() []byte {
	defaultParams := single.GetDefaultRequestParams()
	data, err := defaultParams.MarshalJSON()
	if err != nil {
		jww.FATAL.Panicf("Failed to JSON marshal single-use params: %+v", err)
	}
	return data
}

// GetDefaultE2eFileTransferParams returns a JSON serialized object with all the
// E2E file transfer parameters and their default values. Call this function and
// modify the JSON to change single use settings.
func GetDefaultE2eFileTransferParams() []byte {
	defaultParams := e2eFileTransfer.DefaultParams()
	data, err := json.Marshal(defaultParams)
	if err != nil {
		jww.FATAL.Panicf("Failed to JSON marshal e2e file transfer params: %+v", err)
	}
	return data
}

func parseE2eFileTransferParams(data []byte) (e2eFileTransfer.Params, error) {
	p := &e2eFileTransfer.Params{}
	return *p, json.Unmarshal(data, p)
}

func parseSingleUseParams(data []byte) (single.RequestParams, error) {
	p := &single.RequestParams{}
	return *p, p.UnmarshalJSON(data)
}

func parseFileTransferParams(data []byte) (fileTransfer.Params, error) {
	p := &fileTransfer.Params{}
	return *p, json.Unmarshal(data, p)
}

func parseCMixParams(data []byte) (xxdk.CMIXParams, error) {
	p := &xxdk.CMIXParams{}
	err := p.Unmarshal(data)
	return *p, err
}

func parseE2EParams(data []byte) (xxdk.E2EParams, error) {
	p := &xxdk.E2EParams{}
	err := p.Unmarshal(data)
	return *p, err
}
