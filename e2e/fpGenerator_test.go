////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package e2e

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"gitlab.com/elixxir/client/v4/cmix"
	"gitlab.com/elixxir/client/v4/cmix/gateway"
	"gitlab.com/elixxir/client/v4/cmix/identity"
	"gitlab.com/elixxir/client/v4/cmix/message"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/e2e/ratchet/partner/session"
	"gitlab.com/elixxir/client/v4/stoppable"
	"gitlab.com/elixxir/comms/network"
	"gitlab.com/elixxir/crypto/e2e"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
)

// Adds a list of cyphers with different fingerprints with fpGenerator.AddKey
// and then checks that they were added to the mock cMix fingerprint tracker.
func Test_fpGenerator_AddKey(t *testing.T) {
	prng := rand.New(rand.NewSource(42))
	net := newMockFpgCmix()
	fpg := &fpGenerator{&manager{
		net:  net,
		myID: id.NewIdFromString("myID", id.User, t),
	}}

	fps := make([]format.Fingerprint, 20)
	for i := range fps {
		prng.Read(fps[i][:])
		fpg.AddKey(mockSessionCypher{fps[i]})
	}

	for i, fp := range fps {
		if _, exists := net.processors[*fpg.m.myID][fp]; !exists {
			t.Errorf("Fingerprint #%d does not exist.", i)
		} else {
			delete(net.processors[*fpg.m.myID], fp)
		}
	}

	if len(net.processors[*fpg.m.myID]) != 0 {
		t.Errorf("%d extra fingerprints found: %+v",
			len(net.processors[*fpg.m.myID]), net.processors[*fpg.m.myID])
	}
}

// Adds a list of cyphers with different fingerprints and then deletes all of
// them with fpGenerator.DeleteKey and checks that all keys were deleted.
func Test_fpGenerator_DeleteKey(t *testing.T) {
	prng := rand.New(rand.NewSource(42))
	net := newMockFpgCmix()
	fpg := &fpGenerator{&manager{
		net:  net,
		myID: id.NewIdFromString("myID", id.User, t),
	}}

	fps := make([]format.Fingerprint, 20)
	for i := range fps {
		prng.Read(fps[i][:])
		fpg.AddKey(mockSessionCypher{fps[i]})
	}

	for _, fp := range fps {
		fpg.DeleteKey(mockSessionCypher{fp})
	}

	if len(net.processors[*fpg.m.myID]) != 0 {
		t.Errorf("%d extra fingerprints found: %+v",
			len(net.processors[*fpg.m.myID]), net.processors[*fpg.m.myID])
	}
}

////////////////////////////////////////////////////////////////////////////////
// Mock Session Cypher                                                        //
////////////////////////////////////////////////////////////////////////////////

type mockSessionCypher struct {
	fp format.Fingerprint
}

func (m mockSessionCypher) GetSession() *session.Session    { return nil }
func (m mockSessionCypher) Fingerprint() format.Fingerprint { return m.fp }
func (m mockSessionCypher) Encrypt([]byte) (ecrContents, mac []byte, residue e2e.KeyResidue) {
	return nil, nil, e2e.KeyResidue{}
}
func (m mockSessionCypher) Decrypt(format.Message) ([]byte, e2e.KeyResidue, error) {
	return nil, e2e.KeyResidue{}, nil
}
func (m mockSessionCypher) Use() {}

////////////////////////////////////////////////////////////////////////////////
// Mock cMix                                                           //
////////////////////////////////////////////////////////////////////////////////

type mockFpgCmix struct {
	processors map[id.ID]map[format.Fingerprint]message.Processor
	sync.Mutex
}

func (m *mockFpgCmix) SetTrackNetworkPeriod(d time.Duration) {
	//TODO implement me
	panic("implement me")
}

func newMockFpgCmix() *mockFpgCmix {
	return &mockFpgCmix{
		processors: make(map[id.ID]map[format.Fingerprint]message.Processor),
	}
}

func (m *mockFpgCmix) Follow(cmix.ClientErrorReport) (stoppable.Stoppable, error) { return nil, nil }
func (m *mockFpgCmix) GetMaxMessageLength() int                                   { return 0 }
func (m *mockFpgCmix) Send(*id.ID, format.Fingerprint, cmix.Service, []byte, []byte, cmix.CMIXParams) (rounds.Round, ephemeral.Id, error) {
	return rounds.Round{}, ephemeral.Id{}, nil
}
func (m *mockFpgCmix) SendWithAssembler(recipient *id.ID, assembler cmix.MessageAssembler,
	cmixParams cmix.CMIXParams) (rounds.Round, ephemeral.Id, error) {
	return rounds.Round{}, ephemeral.Id{}, nil
}
func (m *mockFpgCmix) SendMany(messages []cmix.TargetedCmixMessage, params cmix.CMIXParams) (rounds.Round, []ephemeral.Id, error) {
	return rounds.Round{}, nil, nil
}
func (m *mockFpgCmix) SendManyWithAssembler(recipients []*id.ID, assembler cmix.ManyMessageAssembler, params cmix.CMIXParams) (rounds.Round, []ephemeral.Id, error) {
	return rounds.Round{}, nil, nil
}
func (m *mockFpgCmix) AddIdentity(*id.ID, time.Time, bool, message.Processor) {}
func (m *mockFpgCmix) AddIdentityWithHistory(id *id.ID, validUntil,
	beginning time.Time, persistent bool, _ message.Processor) {
}
func (m *mockFpgCmix) RemoveIdentity(*id.ID) {}
func (m *mockFpgCmix) GetIdentity(*id.ID) (identity.TrackedID, error) {
	return identity.TrackedID{}, nil
}

func (m *mockFpgCmix) AddFingerprint(uid *id.ID, fp format.Fingerprint, mp message.Processor) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.processors[*uid]; !exists {
		m.processors[*uid] =
			map[format.Fingerprint]message.Processor{fp: mp}
	} else if _, exists = m.processors[*uid][fp]; !exists {
		m.processors[*uid][fp] = mp
	}

	return nil
}

func (m *mockFpgCmix) DeleteFingerprint(uid *id.ID, fp format.Fingerprint) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.processors[*uid]; exists {
		delete(m.processors[*uid], fp)
	}
}

func (m *mockFpgCmix) UpsertCompressedService(clientID *id.ID, newService message.CompressedService,
	response message.Processor) {
}
func (m *mockFpgCmix) DeleteCompressedService(clientID *id.ID, toDelete message.CompressedService,
	processor message.Processor) {

}

func (m *mockFpgCmix) DeleteClientFingerprints(*id.ID)                       {}
func (m *mockFpgCmix) AddService(*id.ID, message.Service, message.Processor) {}
func (m *mockFpgCmix) IncreaseParallelNodeRegistration(int) func() (stoppable.Stoppable, error) {
	return nil
}
func (m *mockFpgCmix) DeleteService(*id.ID, message.Service, message.Processor) {}
func (m *mockFpgCmix) DeleteClientService(*id.ID)                               {}
func (m *mockFpgCmix) TrackServices(message.ServicesTracker)                    {}
func (m *mockFpgCmix) GetServices() (message.ServiceList, message.CompressedServiceList) {
	return message.ServiceList{}, message.CompressedServiceList{}
}
func (m *mockFpgCmix) CheckInProgressMessages()            {}
func (m *mockFpgCmix) IsHealthy() bool                     { return false }
func (m *mockFpgCmix) WasHealthy() bool                    { return false }
func (m *mockFpgCmix) AddHealthCallback(func(bool)) uint64 { return 0 }
func (m *mockFpgCmix) RemoveHealthCallback(uint64)         {}
func (m *mockFpgCmix) HasNode(*id.ID) bool                 { return false }
func (m *mockFpgCmix) NumRegisteredNodes() int             { return 0 }
func (m *mockFpgCmix) TriggerNodeRegistration(*id.ID)      {}
func (m *mockFpgCmix) GetRoundResults(time.Duration, cmix.RoundEventCallback, ...id.Round) {
}
func (m *mockFpgCmix) LookupHistoricalRound(id.Round, rounds.RoundResultCallback) error { return nil }
func (m *mockFpgCmix) SendToAny(func(host *connect.Host) (interface{}, error), *stoppable.Single) (interface{}, error) {
	return nil, nil
}
func (m *mockFpgCmix) SendToPreferred([]*id.ID, gateway.SendToPreferredFunc, *stoppable.Single, time.Duration) (interface{}, error) {
	return nil, nil
}
func (m *mockFpgCmix) SetGatewayFilter(gateway.Filter)                             {}
func (m *mockFpgCmix) GetHostParams() connect.HostParams                           { return connect.HostParams{} }
func (m *mockFpgCmix) GetAddressSpace() uint8                                      { return 0 }
func (m *mockFpgCmix) RegisterAddressSpaceNotification(string) (chan uint8, error) { return nil, nil }
func (m *mockFpgCmix) UnregisterAddressSpaceNotification(string)                   {}
func (m *mockFpgCmix) GetInstance() *network.Instance                              { return nil }
func (m *mockFpgCmix) GetVerboseRounds() string                                    { return "" }
func (m *mockFpgCmix) PauseNodeRegistrations(timeout time.Duration) error          { return nil }
func (m *mockFpgCmix) ChangeNumberOfNodeRegistrations(toRun int, timeout time.Duration) error {
	return nil
}
