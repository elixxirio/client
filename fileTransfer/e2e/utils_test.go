////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package e2e

import (
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/message"
	"gitlab.com/elixxir/client/cmix/rounds"
	"gitlab.com/elixxir/client/e2e"
	"gitlab.com/elixxir/client/e2e/receive"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/elixxir/crypto/cyclic"
	e2eCrypto "gitlab.com/elixxir/crypto/e2e"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/crypto/large"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"gitlab.com/xx_network/primitives/netTime"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// Mock cMix                                                           //
////////////////////////////////////////////////////////////////////////////////

type mockCmixHandler struct {
	sync.Mutex
	processorMap map[format.Fingerprint]message.Processor
}

func newMockCmixHandler() *mockCmixHandler {
	return &mockCmixHandler{
		processorMap: make(map[format.Fingerprint]message.Processor),
	}
}

type mockCmix struct {
	myID          *id.ID
	numPrimeBytes int
	health        bool
	handler       *mockCmixHandler
	healthCBs     map[uint64]func(b bool)
	healthIndex   uint64
	sync.Mutex
}

func newMockCmix(myID *id.ID, handler *mockCmixHandler, storage *mockStorage) *mockCmix {
	return &mockCmix{
		myID:          myID,
		numPrimeBytes: storage.GetCmixGroup().GetP().ByteLen(),
		health:        true,
		handler:       handler,
		healthCBs:     make(map[uint64]func(b bool)),
		healthIndex:   0,
	}
}

func (m *mockCmix) GetMaxMessageLength() int {
	msg := format.NewMessage(m.numPrimeBytes)
	return msg.ContentsSize()
}

func (m *mockCmix) SendMany(messages []cmix.TargetedCmixMessage,
	_ cmix.CMIXParams) (id.Round, []ephemeral.Id, error) {
	m.handler.Lock()
	for _, targetedMsg := range messages {
		msg := format.NewMessage(m.numPrimeBytes)
		msg.SetContents(targetedMsg.Payload)
		msg.SetMac(targetedMsg.Mac)
		msg.SetKeyFP(targetedMsg.Fingerprint)
		m.handler.processorMap[targetedMsg.Fingerprint].Process(msg,
			receptionID.EphemeralIdentity{Source: targetedMsg.Recipient},
			rounds.Round{ID: 42})
	}
	m.handler.Unlock()
	return 42, []ephemeral.Id{}, nil
}

func (m *mockCmix) AddFingerprint(_ *id.ID, fp format.Fingerprint, mp message.Processor) error {
	m.Lock()
	defer m.Unlock()
	m.handler.processorMap[fp] = mp
	return nil
}

func (m *mockCmix) DeleteFingerprint(_ *id.ID, fp format.Fingerprint) {
	m.handler.Lock()
	delete(m.handler.processorMap, fp)
	m.handler.Unlock()
}

func (m *mockCmix) CheckInProgressMessages() {}

func (m *mockCmix) IsHealthy() bool {
	return m.health
}

func (m *mockCmix) WasHealthy() bool { return true }

func (m *mockCmix) AddHealthCallback(f func(bool)) uint64 {
	m.Lock()
	defer m.Unlock()
	m.healthIndex++
	m.healthCBs[m.healthIndex] = f
	go f(true)
	return m.healthIndex
}

func (m *mockCmix) RemoveHealthCallback(healthID uint64) {
	m.Lock()
	defer m.Unlock()
	if _, exists := m.healthCBs[healthID]; !exists {
		jww.FATAL.Panicf("No health callback with ID %d exists.", healthID)
	}
	delete(m.healthCBs, healthID)
}

func (m *mockCmix) GetRoundResults(_ time.Duration,
	roundCallback cmix.RoundEventCallback, _ ...id.Round) error {
	go roundCallback(true, false, map[id.Round]cmix.RoundResult{42: {}})
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Mock E2E Handler                                                           //
////////////////////////////////////////////////////////////////////////////////
func newMockListener(hearChan chan receive.Message) *mockListener {
	return &mockListener{hearChan: hearChan}
}

func (l *mockListener) Hear(item receive.Message) {
	l.hearChan <- item
}

func (l *mockListener) Name() string {
	return "mockListener"
}

type mockE2eHandler struct {
	msgMap    map[id.ID]map[catalog.MessageType][][]byte
	listeners map[catalog.MessageType]receive.Listener
}

func newMockE2eHandler() *mockE2eHandler {
	return &mockE2eHandler{
		msgMap:    make(map[id.ID]map[catalog.MessageType][][]byte),
		listeners: make(map[catalog.MessageType]receive.Listener),
	}
}

type mockE2e struct {
	myID    *id.ID
	handler *mockE2eHandler
}

type mockListener struct {
	hearChan chan receive.Message
}

func newMockE2e(myID *id.ID, handler *mockE2eHandler) *mockE2e {
	return &mockE2e{
		myID:    myID,
		handler: handler,
	}
}

// SendE2E adds the message to the e2e handler map.
func (m *mockE2e) SendE2E(mt catalog.MessageType, recipient *id.ID, payload []byte,
	_ e2e.Params) ([]id.Round, e2eCrypto.MessageID, time.Time, error) {

	m.handler.listeners[mt].Hear(receive.Message{
		MessageType: mt,
		Payload:     payload,
		Sender:      m.myID,
		RecipientID: recipient,
	})

	return []id.Round{42}, e2eCrypto.MessageID{}, netTime.Now(), nil
}

func (m *mockE2e) RegisterListener(_ *id.ID, mt catalog.MessageType,
	listener receive.Listener) receive.ListenerID {
	m.handler.listeners[mt] = listener
	return receive.ListenerID{}
}

////////////////////////////////////////////////////////////////////////////////
// Mock Storage Session                                                       //
////////////////////////////////////////////////////////////////////////////////

type mockStorage struct {
	kv        *versioned.KV
	cmixGroup *cyclic.Group
}

func newMockStorage() *mockStorage {
	b := make([]byte, 768)
	rng := fastRNG.NewStreamGenerator(1000, 10, csprng.NewSystemRNG).GetStream()
	_, _ = rng.Read(b)
	rng.Close()

	return &mockStorage{
		kv:        versioned.NewKV(ekv.MakeMemstore()),
		cmixGroup: cyclic.NewGroup(large.NewIntFromBytes(b), large.NewInt(2)),
	}
}

func (m *mockStorage) GetKV() *versioned.KV        { return m.kv }
func (m *mockStorage) GetCmixGroup() *cyclic.Group { return m.cmixGroup }
