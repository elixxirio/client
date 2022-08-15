///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"gitlab.com/elixxir/client/wasm"
	"os"
	"os/signal"
	"syscall"
	"syscall/js"
)

func main() {
	fmt.Println("Go Web Assembly")

	// wasm/cmix.go
	js.Global().Set("NewCmix", js.FuncOf(wasm.NewCmix))
	js.Global().Set("LoadCmix", js.FuncOf(wasm.LoadCmix))

	// wasm/e2e.go
	js.Global().Set("Login", js.FuncOf(wasm.Login))
	js.Global().Set("LoginEphemeral", js.FuncOf(wasm.LoginEphemeral))

	// wasm/identity.go
	js.Global().Set("StoreReceptionIdentity",
		js.FuncOf(wasm.StoreReceptionIdentity))
	js.Global().Set("LoadReceptionIdentity",
		js.FuncOf(wasm.LoadReceptionIdentity))
	js.Global().Set("GetIDFromContact",
		js.FuncOf(wasm.GetIDFromContact))
	js.Global().Set("GetPubkeyFromContact",
		js.FuncOf(wasm.GetPubkeyFromContact))
	js.Global().Set("SetFactsOnContact",
		js.FuncOf(wasm.SetFactsOnContact))
	js.Global().Set("GetFactsFromContact",
		js.FuncOf(wasm.GetFactsFromContact))

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	os.Exit(0)
}
