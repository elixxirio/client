////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package broadcast

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/blake2b"

	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/crypto/signature/rsa"

	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/message"
	"gitlab.com/elixxir/client/cmix/rounds"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	cMixCrypto "gitlab.com/elixxir/crypto/cmix"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/primitives/format"
)

// Tests that broadcastClient adheres to the Channel interface.
var _ Channel = (*broadcastClient)(nil)

// Tests that cmix.Client adheres to the Client interface.
var _ Client = (cmix.Client)(nil)

// Tests that mockProcessor adheres to the message.Processor interface.
var _ message.Processor = (*mockProcessor)(nil)

// mockProcessor adheres to the message.Processor interface.
type mockProcessor struct {
	messages []format.Message
}

func newMockProcessor() *mockProcessor {
	m := new(mockProcessor)
	m.messages = make([]format.Message, 0)
	return m
}
func (p *mockProcessor) Process(message format.Message,
	_ receptionID.EphemeralIdentity, _ rounds.Round) {
	p.messages = append(p.messages, message)
}
func (p *mockProcessor) String() string { return "hello" }

func Test_asymmetricClient_Smoke(t *testing.T) {
	cMixHandler := newMockCmixHandler()
	rngGen := fastRNG.NewStreamGenerator(1000, 10, csprng.NewSystemRNG)
	pk, err := rsa.GenerateKey(rngGen.GetStream(), 4096)
	if err != nil {
		t.Fatalf("Failed to generate priv key: %+v", err)
	}
	cName := "MyChannel"
	cDesc := "This is my channel about stuff."
	cSalt := cMixCrypto.NewSalt(csprng.NewSystemRNG(), 32)
	cPubKey := pk.GetPublic()
	pubKeyHash := blake2b.Sum256(cPubKey.Bytes())
	secret := make([]byte, 32)
	_, err = rngGen.GetStream().Read(secret)
	if err != nil {
		t.Fatal(err)
	}
	cid, err := crypto.NewChannelID(
		cName, cDesc, cSalt, pubKeyHash[:], secret)
	if err != nil {
		t.Errorf("Failed to create channel ID: %+v", err)
	}
	channel := &crypto.Channel{
		ReceptionID:   cid,
		Name:          cName,
		Description:   cDesc,
		Salt:          cSalt,
		RsaPubKeyHash: pubKeyHash[:],
		Secret:        secret,
	}

	// must mutate cMixHandler such that it's processorMap contains a
	// message.Processor
	processor := newMockProcessor()
	cMixHandler.processorMap[*cid] = make(map[string][]message.Processor)
	cMixHandler.processorMap[*cid]["AsymmBcast"] = []message.Processor{processor}

	const n = 1
	cbChans := make([]chan []byte, n)
	clients := make([]Channel, n)
	for i := range clients {
		cbChan := make(chan []byte, 10)
		cb := func(payload []byte, _ receptionID.EphemeralIdentity,
			_ rounds.Round) {
			cbChan <- payload
		}

		s, err := NewBroadcastChannel(channel, newMockCmix(cMixHandler), rngGen)
		if err != nil {
			t.Errorf("Failed to create broadcast channel: %+v", err)
		}

		err = s.RegisterListener(cb, Asymmetric)
		if err != nil {
			t.Errorf("Failed to register listener: %+v", err)
		}

		cbChans[i] = cbChan
		clients[i] = s

		// Test that Get returns the expected channel
		if !reflect.DeepEqual(s.Get(), channel) {
			t.Errorf("Cmix %d returned wrong channel."+
				"\nexpected: %+v\nreceived: %+v", i, channel, s.Get())
		}
	}

	// Send broadcast from each client
	for i := range clients {
		payload := make([]byte, clients[i].MaxAsymmetricPayloadSize(cPubKey))
		copy(payload,
			fmt.Sprintf("Hello from client %d of %d.", i, len(clients)))

		// Start processes that waits for each client to receive broadcast
		var wg sync.WaitGroup
		for j := range cbChans {
			wg.Add(1)
			go func(i, j int, cbChan chan []byte) {
				defer wg.Done()
				select {
				case r := <-cbChan:
					if !bytes.Equal(payload, r) {
						t.Errorf("Cmix %d failed to receive expected payload "+
							"from client %d.\nexpected: %q\nreceived: %q",
							j, i, payload, r)
					}
				case <-time.After(time.Second):
					t.Errorf("Cmix %d timed out waiting for broadcast "+
						"payload from client %d.", j, i)
				}
			}(i, j, cbChans[j])
		}

		// Broadcast payload
		_, _, err = clients[i].BroadcastAsymmetric(
			pk, payload, cmix.GetDefaultCMIXParams())
		if err != nil {
			t.Errorf("Cmix %d failed to send broadcast: %+v", i, err)
		}

		// Wait for all clients to receive payload or time out
		wg.Wait()
	}

	// Stop each client
	for i := range clients {
		clients[i].Stop()
	}

	payload := make([]byte, clients[0].MaxAsymmetricPayloadSize(cPubKey))
	copy(payload, "This message should not get through.")

	// Start waiting on channels and error if anything is received
	var wg sync.WaitGroup
	for i := range cbChans {
		wg.Add(1)
		go func(i int, cbChan chan []byte) {
			defer wg.Done()
			select {
			case r := <-cbChan:
				t.Errorf("Cmix %d received message: %q", i, r)
			case <-time.After(25 * time.Millisecond):
			}
		}(i, cbChans[i])
	}

	// Broadcast payload
	_, _, err = clients[0].BroadcastAsymmetric(pk, payload, cmix.GetDefaultCMIXParams())
	if err != nil {
		t.Errorf("Cmix 0 failed to send broadcast: %+v", err)
	}

	wg.Wait()
}
