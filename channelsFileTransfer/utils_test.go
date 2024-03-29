////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package channelsFileTransfer

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	clientNotif "gitlab.com/elixxir/client/v4/notifications"

	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix"
	"gitlab.com/elixxir/client/v4/cmix/gateway"
	"gitlab.com/elixxir/client/v4/cmix/identity"
	"gitlab.com/elixxir/client/v4/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/v4/cmix/message"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/client/v4/e2e"
	"gitlab.com/elixxir/client/v4/stoppable"
	"gitlab.com/elixxir/client/v4/storage"
	userStorage "gitlab.com/elixxir/client/v4/storage/user"
	"gitlab.com/elixxir/client/v4/xxdk"
	"gitlab.com/elixxir/comms/network"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/fastRNG"
	ftCrypto "gitlab.com/elixxir/crypto/fileTransfer"
	cryptoMessage "gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/crypto/rsa"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/elixxir/primitives/states"
	"gitlab.com/elixxir/primitives/version"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/crypto/large"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"gitlab.com/xx_network/primitives/ndf"
	"gitlab.com/xx_network/primitives/netTime"
)

// newFile generates a file with random data of size numParts * partSize.
// Returns the full file and the file parts. If the partSize allows, each part
// starts with a "|<[PART_001]" and ends with a ">|".
func newFile(numParts uint16, partSize int, prng io.Reader, t *testing.T) (
	[]byte, [][]byte) {
	const (
		prefix = "|<[PART_%3d]"
		suffix = ">|"
	)
	// Create file buffer of the expected size
	fileBuff := bytes.NewBuffer(make([]byte, 0, int(numParts)*partSize))
	partList := make([][]byte, numParts)

	// Create new rand.Rand with the seed generated from the io.Reader
	b := make([]byte, 8)
	_, err := prng.Read(b)
	if err != nil {
		t.Errorf("Failed to generate random seed: %+v", err)
	}
	seed := binary.LittleEndian.Uint64(b)
	randPrng := rand.New(rand.NewSource(int64(seed)))

	for partNum := range partList {
		s := randStringBytes(partSize, randPrng)
		if len(s) >= (len(prefix) + len(suffix)) {
			partList[partNum] = []byte(
				prefix + s[:len(s)-(len(prefix)+len(suffix))] + suffix)
		} else {
			partList[partNum] = []byte(s)
		}

		fileBuff.Write(partList[partNum])
	}

	return fileBuff.Bytes(), partList
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randStringBytes generates a random string of length n consisting of the
// characters in letterBytes.
func randStringBytes(n int, prng *rand.Rand) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[prng.Intn(len(letterBytes))]
	}
	return string(b)
}

////////////////////////////////////////////////////////////////////////////////
// Mock xxdk.E2e                                                              //
////////////////////////////////////////////////////////////////////////////////

type mockE2e struct {
	rid xxdk.ReceptionIdentity
	c   cmix.Client
	s   storage.Session
	rng *fastRNG.StreamGenerator
}

func newMockE2e(rid *id.ID, c cmix.Client, s storage.Session,
	rng *fastRNG.StreamGenerator) *mockE2e {
	return &mockE2e{
		rid: xxdk.ReceptionIdentity{ID: rid},
		c:   c,
		s:   s,
		rng: rng,
	}
}

func (m *mockE2e) GetStorage() storage.Session                  { return m.s }
func (m *mockE2e) GetReceptionIdentity() xxdk.ReceptionIdentity { return m.rid }
func (m *mockE2e) GetCmix() cmix.Client                         { return m.c }
func (m *mockE2e) GetRng() *fastRNG.StreamGenerator             { return m.rng }
func (m *mockE2e) GetE2E() e2e.Handler                          { return nil }

// //////////////////////////////////////////////////////////////////////////////
// Mock cMix                                                                  //
// //////////////////////////////////////////////////////////////////////////////
type cmixMsg struct {
	rid         id.Round
	targetedMsg cmix.TargetedCmixMessage
	msg         format.Message
}

type mockCmixHandler struct {
	sync.Mutex
	processorMap map[format.Fingerprint]message.Processor
	messageList  map[format.Fingerprint]cmixMsg
}

func newMockCmixHandler() *mockCmixHandler {
	return &mockCmixHandler{
		processorMap: make(map[format.Fingerprint]message.Processor),
		messageList:  make(map[format.Fingerprint]cmixMsg),
	}
}

type mockCmix struct {
	failSomeSends bool

	myID          *id.ID
	numPrimeBytes int
	health        bool
	handler       *mockCmixHandler
	healthCBs     map[uint64]func(b bool)
	healthIndex   uint64
	round         id.Round
	prng          *rand.Rand
	sync.Mutex
}

func newMockCmix(
	myID *id.ID, handler *mockCmixHandler, storage *mockStorage) *mockCmix {
	return &mockCmix{
		myID:          myID,
		numPrimeBytes: storage.GetCmixGroup().GetP().ByteLen(),
		health:        true,
		handler:       handler,
		healthCBs:     make(map[uint64]func(b bool)),
		healthIndex:   0,
		round:         0,
		prng:          rand.New(rand.NewSource(42)),
	}
}

func (m *mockCmix) Follow(cmix.ClientErrorReport) (stoppable.Stoppable, error) { panic("implement me") }
func (m *mockCmix) SetTrackNetworkPeriod(time.Duration)                        { panic("implement me") }

func (m *mockCmix) GetMaxMessageLength() int {
	msg := format.NewMessage(m.numPrimeBytes)
	return msg.ContentsSize()
}

func (m *mockCmix) Send(*id.ID, format.Fingerprint, cmix.Service, []byte,
	[]byte, cmix.CMIXParams) (rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}

func (m *mockCmix) SendMany(messages []cmix.TargetedCmixMessage,
	_ cmix.CMIXParams) (rounds.Round, []ephemeral.Id, error) {
	m.handler.Lock()
	defer m.handler.Unlock()
	rid := m.round
	m.round++
	for _, targetedMsg := range messages {
		msg := format.NewMessage(m.numPrimeBytes)
		msg.SetContents(targetedMsg.Payload)
		msg.SetMac(targetedMsg.Mac)
		msg.SetKeyFP(targetedMsg.Fingerprint)
		m.handler.messageList[targetedMsg.Fingerprint] =
			cmixMsg{rid, targetedMsg, msg}

		// Fail to process some messages so that resending can be tested
		if m.failSomeSends && m.prng.Intn(10) == 5 {
			continue
		}

		mp, exists := m.handler.processorMap[targetedMsg.Fingerprint]
		if exists {
			go func(mp message.Processor, rid id.Round,
				targetedMsg cmix.TargetedCmixMessage, msg format.Message) {
				mp.Process(msg, nil, nil, receptionID.EphemeralIdentity{
					Source: targetedMsg.Recipient}, rounds.Round{ID: rid},
				)
			}(mp, rid, targetedMsg, msg)
		}
	}
	return rounds.Round{ID: rid}, []ephemeral.Id{}, nil
}

func (m *mockCmix) SendWithAssembler(*id.ID, cmix.MessageAssembler, cmix.CMIXParams) (rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockCmix) SendManyWithAssembler([]*id.ID, cmix.ManyMessageAssembler, cmix.CMIXParams) (rounds.Round, []ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockCmix) AddIdentity(*id.ID, time.Time, bool, message.Processor) {}
func (m *mockCmix) AddIdentityWithHistory(*id.ID, time.Time, time.Time, bool, message.Processor) {
	panic("implement me")
}
func (m *mockCmix) RemoveIdentity(*id.ID)                          {}
func (m *mockCmix) GetIdentity(*id.ID) (identity.TrackedID, error) { panic("implement me") }
func (m *mockCmix) AddFingerprint(_ *id.ID, fp format.Fingerprint, mp message.Processor) error {
	m.handler.Lock()
	defer m.handler.Unlock()
	m.handler.processorMap[fp] = mp

	p, exists := m.handler.messageList[fp]
	if exists {
		go mp.Process(
			p.msg, nil, nil,
			receptionID.EphemeralIdentity{Source: p.targetedMsg.Recipient},
			rounds.Round{ID: p.rid},
		)
	}

	return nil
}

func (m *mockCmix) DeleteFingerprint(_ *id.ID, fp format.Fingerprint) {
	m.handler.Lock()
	defer m.handler.Unlock()
	delete(m.handler.processorMap, fp)
}

func (m *mockCmix) DeleteClientFingerprints(*id.ID) {
	m.handler.Lock()
	defer m.handler.Unlock()
	m.handler.processorMap = make(map[format.Fingerprint]message.Processor)
}

func (m *mockCmix) AddService(*id.ID, message.Service, message.Processor) { panic("implement me") }
func (m *mockCmix) UpsertCompressedService(*id.ID, message.CompressedService, message.Processor) {
	panic("implement me")
}
func (m *mockCmix) DeleteCompressedService(*id.ID, message.CompressedService, message.Processor) {
	panic("implement me")
}
func (m *mockCmix) PauseNodeRegistrations(time.Duration) error               { panic("implement me") }
func (m *mockCmix) ChangeNumberOfNodeRegistrations(int, time.Duration) error { panic("implement me") }
func (m *mockCmix) DeleteService(*id.ID, message.Service, message.Processor) { panic("implement me") }
func (m *mockCmix) DeleteClientService(*id.ID)                               { panic("implement me") }
func (m *mockCmix) TrackServices(message.ServicesTracker)                    { panic("implement me") }
func (m *mockCmix) GetServices() (message.ServiceList, message.CompressedServiceList) {
	panic("implement me")
}
func (m *mockCmix) CheckInProgressMessages() {}
func (m *mockCmix) IsHealthy() bool          { return m.health }
func (m *mockCmix) WasHealthy() bool         { return true }

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

func (m *mockCmix) HasNode(*id.ID) bool            { panic("implement me") }
func (m *mockCmix) NumRegisteredNodes() int        { panic("implement me") }
func (m *mockCmix) TriggerNodeRegistration(*id.ID) { panic("implement me") }

func (m *mockCmix) GetRoundResults(_ time.Duration,
	roundCallback cmix.RoundEventCallback, rids ...id.Round) {
	go roundCallback(true, false, map[id.Round]cmix.RoundResult{
		rids[0]: {
			Status: cmix.Succeeded,
			Round: rounds.Round{
				Timestamps: map[states.Round]time.Time{
					states.COMPLETED: netTime.Now(),
				},
			},
		}})
}

func (m *mockCmix) LookupHistoricalRound(id.Round, rounds.RoundResultCallback) error {
	panic("implement me")
}
func (m *mockCmix) SendToAny(func(host *connect.Host) (interface{}, error),
	*stoppable.Single) (interface{}, error) {
	panic("implement me")
}
func (m *mockCmix) SendToPreferred([]*id.ID, gateway.SendToPreferredFunc,
	*stoppable.Single, time.Duration) (interface{}, error) {
	panic("implement me")
}
func (m *mockCmix) GetHostParams() connect.HostParams { panic("implement me") }
func (m *mockCmix) GetAddressSpace() uint8            { panic("implement me") }
func (m *mockCmix) RegisterAddressSpaceNotification(string) (chan uint8, error) {
	panic("implement me")
}
func (m *mockCmix) UnregisterAddressSpaceNotification(string) { panic("implement me") }
func (m *mockCmix) GetInstance() *network.Instance            { panic("implement me") }
func (m *mockCmix) GetVerboseRounds() string                  { panic("implement me") }

////////////////////////////////////////////////////////////////////////////////
// Mock Storage Session                                                       //
////////////////////////////////////////////////////////////////////////////////

type mockStorage struct {
	kv        versioned.KV
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

func (m *mockStorage) GetClientVersion() version.Version     { panic("implement me") }
func (m *mockStorage) Get(string) (*versioned.Object, error) { panic("implement me") }
func (m *mockStorage) Set(string, *versioned.Object) error   { panic("implement me") }
func (m *mockStorage) Delete(string) error                   { panic("implement me") }
func (m *mockStorage) GetKV() versioned.KV                   { return m.kv }
func (m *mockStorage) GetCmixGroup() *cyclic.Group           { return m.cmixGroup }
func (m *mockStorage) GetE2EGroup() *cyclic.Group            { panic("implement me") }
func (m *mockStorage) ForwardRegistrationStatus(storage.RegistrationStatus) error {
	panic("implement me")
}
func (m *mockStorage) RegStatus() storage.RegistrationStatus                  { panic("implement me") }
func (m *mockStorage) SetRegCode(string)                                      { panic("implement me") }
func (m *mockStorage) GetRegCode() (string, error)                            { panic("implement me") }
func (m *mockStorage) SetNDF(*ndf.NetworkDefinition)                          { panic("implement me") }
func (m *mockStorage) GetNDF() *ndf.NetworkDefinition                         { panic("implement me") }
func (m *mockStorage) GetTransmissionID() *id.ID                              { panic("implement me") }
func (m *mockStorage) GetTransmissionSalt() []byte                            { panic("implement me") }
func (m *mockStorage) GetReceptionID() *id.ID                                 { panic("implement me") }
func (m *mockStorage) GetReceptionSalt() []byte                               { panic("implement me") }
func (m *mockStorage) GetReceptionRSA() rsa.PrivateKey                        { panic("implement me") }
func (m *mockStorage) GetTransmissionRSA() rsa.PrivateKey                     { panic("implement me") }
func (m *mockStorage) IsPrecanned() bool                                      { panic("implement me") }
func (m *mockStorage) SetUsername(string) error                               { panic("implement me") }
func (m *mockStorage) GetUsername() (string, error)                           { panic("implement me") }
func (m *mockStorage) PortableUserInfo() userStorage.Info                     { panic("implement me") }
func (m *mockStorage) GetTransmissionRegistrationValidationSignature() []byte { panic("implement me") }
func (m *mockStorage) GetReceptionRegistrationValidationSignature() []byte    { panic("implement me") }
func (m *mockStorage) GetRegistrationTimestamp() time.Time                    { panic("implement me") }
func (m *mockStorage) SetTransmissionRegistrationValidationSignature([]byte)  { panic("implement me") }
func (m *mockStorage) SetReceptionRegistrationValidationSignature([]byte)     { panic("implement me") }
func (m *mockStorage) SetRegistrationTimestamp(int64)                         { panic("implement me") }

////////////////////////////////////////////////////////////////////////////////
// Mock Event Model                                                           //
////////////////////////////////////////////////////////////////////////////////

// Ensure that mockEventModel adheres to EventModel.
var _ EventModel = (*mockEventModel)(nil)

// mockEventModel mocks the EventModel for testing.
type mockEventModel struct {
	fileCB    func(ModelFile)
	msgCB     func(modelMessage channels.ModelMessage)
	files     map[ftCrypto.ID]ModelFile
	messages  map[uint64]channels.ModelMessage
	messageID uint64
	t         testing.TB
	sync.RWMutex
}

func newMockEventModel(fileCB func(ModelFile),
	msgCB func(modelMessage channels.ModelMessage), t testing.TB) *mockEventModel {
	return &mockEventModel{
		fileCB:    fileCB,
		msgCB:     msgCB,
		files:     make(map[ftCrypto.ID]ModelFile),
		messages:  make(map[uint64]channels.ModelMessage),
		messageID: 0,
		t:         t,
	}
}

func (m *mockEventModel) ReceiveFile(fileID ftCrypto.ID, fileLink,
	fileData []byte, timestamp time.Time, status Status) error {
	m.Lock()
	defer m.Unlock()

	m.files[fileID] = ModelFile{fileID, fileLink, fileData, timestamp, status}

	go m.fileCB(m.files[fileID])
	return nil
}

func (m *mockEventModel) UpdateFile(fileID ftCrypto.ID, fileLink,
	fileData []byte, timestamp *time.Time, status *Status) error {
	m.Lock()
	defer m.Unlock()

	f, exists := m.files[fileID]
	if !exists {
		return channels.NoMessageErr
	}

	if fileLink != nil {
		f.Link = fileLink
	}
	if fileData != nil {
		f.Data = fileData
	}
	if timestamp != nil {
		f.Timestamp = *timestamp
	}
	if status != nil {
		f.Status = *status
	}

	m.files[fileID] = f

	go m.fileCB(m.files[fileID])

	return nil
}

func (m *mockEventModel) GetFile(fileID ftCrypto.ID) (ModelFile, error) {
	m.Lock()
	defer m.Unlock()

	f, exists := m.files[fileID]
	if !exists {
		return ModelFile{}, channels.NoMessageErr
	}

	return f, nil
}

func (m *mockEventModel) DeleteFile(fileID ftCrypto.ID) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.files[fileID]; !exists {
		return channels.NoMessageErr
	}

	delete(m.files, fileID)
	return nil
}

func (m *mockEventModel) JoinChannel(*cryptoBroadcast.Channel) { panic("implement me") }
func (m *mockEventModel) LeaveChannel(*id.ID)                  { panic("implement me") }
func (m *mockEventModel) ReceiveMessage(channelID *id.ID, messageID cryptoMessage.ID, nickname,
	text string, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, lease time.Duration, round rounds.Round,
	messageType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	m.Lock()
	defer m.Unlock()

	newID := m.messageID
	m.messageID++
	m.messages[newID] = channels.ModelMessage{
		UUID:            newID,
		Nickname:        nickname,
		MessageID:       messageID,
		ChannelID:       channelID,
		ParentMessageID: cryptoMessage.ID{},
		Timestamp:       timestamp,
		Lease:           lease,
		Status:          status,
		Hidden:          hidden,
		Pinned:          false,
		Content:         []byte(text),
		Type:            messageType,
		Round:           round.ID,
		PubKey:          pubKey,
		CodesetVersion:  codeset,
		DmToken:         dmToken,
	}

	go m.msgCB(m.messages[newID])

	return newID
}
func (m *mockEventModel) ReceiveReply(*id.ID, cryptoMessage.ID, cryptoMessage.ID, string, string, ed25519.PublicKey, uint32, uint8, time.Time, time.Duration, rounds.Round, channels.MessageType, channels.SentStatus, bool) uint64 {
	panic("implement me")
}
func (m *mockEventModel) ReceiveReaction(*id.ID, cryptoMessage.ID, cryptoMessage.ID, string, string, ed25519.PublicKey, uint32, uint8, time.Time, time.Duration, rounds.Round, channels.MessageType, channels.SentStatus, bool) uint64 {
	panic("implement me")
}
func (m *mockEventModel) UpdateFromUUID(uint64, *cryptoMessage.ID, *time.Time, *rounds.Round, *bool, *bool, *channels.SentStatus) error {
	panic("implement me")
}
func (m *mockEventModel) UpdateFromMessageID(cryptoMessage.ID, *time.Time, *rounds.Round, *bool, *bool, *channels.SentStatus) (uint64, error) {
	panic("implement me")
}
func (m *mockEventModel) GetMessage(cryptoMessage.ID) (channels.ModelMessage, error) {
	panic("implement me")
}
func (m *mockEventModel) DeleteMessage(cryptoMessage.ID) error     { panic("implement me") }
func (m *mockEventModel) MuteUser(*id.ID, ed25519.PublicKey, bool) { panic("implement me") }

////////////////////////////////////////////////////////////////////////////////
// Mock Channels Manager                                                      //
////////////////////////////////////////////////////////////////////////////////

// Ensure that mockChannelsManager adheres to channels.Manager.
var _ channels.Manager = (*mockChannelsManager)(nil)

// mockChannelsManager mocks the channels.Manager for testing.
type mockChannelsManager struct {
	me  cryptoChannel.PrivateIdentity
	emh []channels.ExtensionMessageHandler
}

func newMockChannelsManager(
	me cryptoChannel.PrivateIdentity) (*mockChannelsManager, error) {
	m := &mockChannelsManager{
		me:  me,
		emh: []channels.ExtensionMessageHandler{},
	}

	return m, nil
}

func (m *mockChannelsManager) addEMH(emh ...channels.ExtensionMessageHandler) {
	m.emh = append(m.emh, emh...)
}

func (m *mockChannelsManager) GenerateChannel(string, string,
	cryptoBroadcast.PrivacyLevel) (*cryptoBroadcast.Channel, error) {
	panic("implement me")
}
func (m *mockChannelsManager) JoinChannel(*cryptoBroadcast.Channel) error { panic("implement me") }
func (m *mockChannelsManager) LeaveChannel(*id.ID) error                  { panic("implement me") }
func (m *mockChannelsManager) EnableDirectMessages(*id.ID) error          { panic("implement me") }
func (m *mockChannelsManager) DisableDirectMessages(*id.ID) error         { panic("implement me") }
func (m *mockChannelsManager) AreDMsEnabled(*id.ID) bool                  { panic("implement me") }
func (m *mockChannelsManager) ReplayChannel(*id.ID) error                 { panic("implement me") }
func (m *mockChannelsManager) GetChannels() []*id.ID                      { panic("implement me") }
func (m *mockChannelsManager) GetChannel(*id.ID) (*cryptoBroadcast.Channel, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SendSilent(*id.ID, time.Duration, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SendInvite(*id.ID, string, *cryptoBroadcast.Channel, string, time.Duration, cmix.CMIXParams, []ed25519.PublicKey) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) GetNotificationStatus(*id.ID) (clientNotif.NotificationState, error) {
	panic("implement me")
}

func (m *mockChannelsManager) SendGeneric(channelID *id.ID,
	messageType channels.MessageType, msg []byte, validUntil time.Duration,
	_ bool, _ cmix.CMIXParams, _ map[channels.PingType][]ed25519.PublicKey) (
	cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {

	msgID := cryptoMessage.DeriveChannelMessageID(channelID, 0, msg)
	for _, emh := range m.emh {
		emh.Handle(channelID, msgID, messageType, "", msg, nil, m.me.PubKey,
			m.me.GetDMToken(), 0, netTime.Now(), netTime.Now(), validUntil, 0,
			rounds.Round{}, channels.Delivered, false, false)
	}

	return msgID, rounds.Round{}, ephemeral.Id{}, nil
}

func (m *mockChannelsManager) SendMessage(*id.ID, string, time.Duration, cmix.CMIXParams, []ed25519.PublicKey) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SendReply(*id.ID, string, cryptoMessage.ID, time.Duration, cmix.CMIXParams, []ed25519.PublicKey) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SendReaction(*id.ID, string, cryptoMessage.ID, time.Duration, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SendAdminGeneric(*id.ID, channels.MessageType, []byte, time.Duration, bool, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) DeleteMessage(*id.ID, cryptoMessage.ID, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) PinMessage(*id.ID, cryptoMessage.ID, bool, time.Duration, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) MuteUser(*id.ID, ed25519.PublicKey, bool, time.Duration, cmix.CMIXParams) (cryptoMessage.ID, rounds.Round, ephemeral.Id, error) {
	panic("implement me")
}
func (m *mockChannelsManager) GetIdentity() cryptoChannel.Identity          { panic("implement me") }
func (m *mockChannelsManager) ExportPrivateIdentity(string) ([]byte, error) { panic("implement me") }
func (m *mockChannelsManager) GetStorageTag() string                        { panic("implement me") }
func (m *mockChannelsManager) RegisterReceiveHandler(channels.MessageType, *channels.ReceiveMessageHandler) error {
	panic("implement me")
}
func (m *mockChannelsManager) SetNickname(string, *id.ID) error { panic("implement me") }
func (m *mockChannelsManager) DeleteNickname(*id.ID) error      { panic("implement me") }

func (m *mockChannelsManager) GetNickname(*id.ID) (nickname string, exists bool) {
	return "", false
}

func (m *mockChannelsManager) Muted(*id.ID) bool                        { panic("implement me") }
func (m *mockChannelsManager) GetMutedUsers(*id.ID) []ed25519.PublicKey { panic("implement me") }
func (m *mockChannelsManager) GetNotificationLevel(*id.ID) (channels.NotificationLevel, error) {
	panic("implement me")
}
func (m *mockChannelsManager) SetMobileNotificationsLevel(*id.ID, channels.NotificationLevel, clientNotif.NotificationState) error {
	panic("implement me")
}
func (m *mockChannelsManager) IsChannelAdmin(*id.ID) bool { panic("implement me") }
func (m *mockChannelsManager) ExportChannelAdminKey(*id.ID, string) ([]byte, error) {
	panic("implement me")
}
func (m *mockChannelsManager) VerifyChannelAdminKey(*id.ID, string, []byte) (bool, error) {
	panic("implement me")
}
func (m *mockChannelsManager) ImportChannelAdminKey(*id.ID, string, []byte) error {
	panic("implement me")
}
func (m *mockChannelsManager) DeleteChannelAdminKey(*id.ID) error { panic("implement me") }
