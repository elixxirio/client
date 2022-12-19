///////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/crypto/codename"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
)

// EventModelBuilder builds an event model
type DMReceiverBuilder interface {
	Build(path string) DMReceiver
}

// DMClient is the bindings level interface for the direct messaging
// client.  It implements all the dm.Client functions but converts
// from basic types that are friendly to gomobile and javascript wasm
// interfaces (e.g., []byte, int, and string).
//
// Users of the bindings api can create multiple DMClients, which are
// tracked via a private singleton.
type DMClient struct {
	api dm.Client
	// NOTE: This matches the integer in the dmClientTracker singleton
	id int
}

// NewDMClientWithGoEventModel creates a new [DMClient] from a
// private identity ([dm.PrivateIdentity]). This is not compatible with
// GoMobile Bindings because it receives the go event model.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [GenerateDMIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// Parameters:
//   - cmixID - The tracked Cmix object ID. This can be retrieved using
//     [Cmix.GetID].
//   - privateIdentity - Bytes of a private identity
//     ([codename.PrivateIdentity]) that is generated by
//     [GenerateIdentity].
//   - receiverBuild - A function that initialises and returns the
//     Receiver event model that is not compatible with GoMobile
//     bindings.
func NewDMClientWithGoEventModel(cmixID int, privateIdentity []byte,
	receiver dm.EventModel) (*DMClient, error) {
	pi, err := codename.UnmarshalPrivateIdentity(privateIdentity)
	if err != nil {
		return nil, err
	}

	// Get user from singleton
	user, err := cmixTrackerSingleton.get(cmixID)
	if err != nil {
		return nil, err
	}

	nickMgr := dm.NewNicknameManager(user.api.GetStorage().GetReceptionID(),
		user.api.GetStorage().GetKV())

	sendTracker := dm.NewSendTracker(user.api.GetStorage().GetKV())

	m := dm.NewDMClient(&pi, receiver, sendTracker, nickMgr,
		user.api.GetCmix(), user.api.GetRng())
	if err != nil {
		return nil, err
	}

	// Add channel to singleton and return
	return dmClients.add(m), nil
}

// NewDMClient creates a new [DMClient] from a private
// identity ([channel.PrivateIdentity]).
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [GenerateDMIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// Parameters:
//   - cmixID - The tracked Cmix object ID. This can be retrieved using
//     [Cmix.GetID].
//   - privateIdentity - Bytes of a private identity
//     ([codename.PrivateIdentity]) that is generated by
//     [codename.GenerateIdentity].
//   - event - An interface that contains a function that initialises
//     and returns the event model that is bindings-compatible.
func NewDMClient(cmixID int, privateIdentity []byte,
	receiverBuilder DMReceiverBuilder) (*DMClient, error) {
	pi, err := codename.UnmarshalPrivateIdentity(privateIdentity)
	if err != nil {
		return nil, err
	}

	// Get user from singleton
	user, err := cmixTrackerSingleton.get(cmixID)
	if err != nil {
		return nil, err
	}

	eb := func(path string) (dm.EventModel, error) {
		return NewDMReceiver(receiverBuilder.Build(path)), nil
	}

	// We path to the string of the public key for this user
	dmPath := base64.RawStdEncoding.EncodeToString(pi.PubKey[:])
	receiver, err := eb(dmPath)
	if err != nil {
		return nil, err
	}

	nickMgr := dm.NewNicknameManager(user.api.GetStorage().GetReceptionID(),
		user.api.GetStorage().GetKV())

	sendTracker := dm.NewSendTracker(user.api.GetStorage().GetKV())

	m := dm.NewDMClient(&pi, receiver, sendTracker, nickMgr,
		user.api.GetCmix(), user.api.GetRng())
	if err != nil {
		return nil, err
	}

	// Add channel to singleton and return
	return dmClients.add(m), nil
}

// GetID returns the tracker ID for the DMClient object.
func (cm *DMClient) GetID() int {
	return cm.id
}

// GetPublicKey returns the public key bytes for this client
func (cm *DMClient) GetPublicKey() []byte {
	return cm.api.GetPublicKey().Bytes()
}

// GetToken returns the dm token for this client
func (cm *DMClient) GetToken() uint32 {
	return cm.api.GetToken()
}

// GetIdentity returns the public identity associated with this DMClient
func (cm *DMClient) GetIdentity() []byte {
	return cm.api.GetIdentity().Marshal()
}

// ExportPrivateIdentity encrypts and exports the private identity to a
// portable string.
func (cm *DMClient) ExportPrivateIdentity(password string) ([]byte, error) {
	return cm.api.ExportPrivateIdentity(password)
}

// GetNickname gets a nickname associated with this DM partner
// (reception) ID.
func (cm *DMClient) GetNickname(idBytes []byte) (string, error) {
	chid, err := id.Unmarshal(idBytes)
	if err != nil {
		return "", err
	}
	nick, exists := cm.api.GetNickname(chid)
	if !exists {
		return "", errors.New("no nickname found")
	}

	return nick, nil
}

// SetNickname sets the nickname to use
func (cm *DMClient) SetNickname(nick string) {
	cm.api.SetNickname(nick)
}

// SendText is used to send a formatted direct message.
//
// Parameters:
//   - partnerPubKeyBytes - the bytes of the public key of the partner's ed25519
//     signing key.
//   - dmToken - the token used to derive the reception id for the partner.
//   - message - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the channels.Manager
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be
//     empty, and GetDefaultCMixParams will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport
func (cm *DMClient) SendText(partnerPubKeyBytes []byte, dmToken uint32,
	message string, leaseTimeMS int64, cmixParamsJSON []byte) ([]byte,
	error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Send message
	msgID, rnd, ephID, err := cm.api.SendText(&partnerPubKey, dmToken,
		message, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendReply is used to send a formatted direct message.
//
// If the message ID the reply is sent to is nonexistent, the other side will
// post the message as a normal message and not a reply. The message will auto
// delete validUntil after the round it is sent in, lasting forever if
// [channels.ValidForever] is used.
//
// Parameters:
//   - partnerPubKeyBytes - the bytes of the public key of the partner's ed25519
//     signing key.
//   - dmToken - the token used to derive the reception id for the partner.
//   - message - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected.
//   - messageToReactTo - The marshalled [channel.MessageID] of the message you
//     wish to reply to. This may be found in the ChannelSendReport if replying
//     to your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the ChannelMessageReceptionCallback registered using
//     RegisterReceiveHandler.
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the channels.Manager
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport
func (cm *DMClient) SendReply(partnerPubKeyBytes []byte, dmToken uint32,
	content string, messageToReactTo []byte, leaseTimeMS int64,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Unmarshal message ID
	msgId := message.ID{}
	copy(msgId[:], messageToReactTo)

	// Send Reply
	msgID, rnd, ephID, err := cm.api.SendReply(&partnerPubKey, dmToken,
		content, msgId, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendReaction is used to send a reaction to a direct message
// The reaction must be a single emoji with no other characters, and will
// be rejected otherwise.
// Users will drop the reaction if they do not recognize the reactTo message.
//
// Parameters:
//   - partnerPubKeyBytes - the bytes of the public key of the partner's ed25519
//     signing key.
//   - dmToken - the token used to derive the reception id for the partner.
//   - reaction - The user's reaction. This should be a single emoji with no
//     other characters. As such, a Unicode string is expected.
//   - messageToReactTo - The marshalled [channel.MessageID] of the message you
//     wish to reply to. This may be found in the ChannelSendReport if replying
//     to your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the ChannelMessageReceptionCallback registered using
//     RegisterReceiveHandler.
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport.
func (dmc *DMClient) SendReaction(partnerPubKeyBytes []byte, dmToken uint32,
	reaction string, messageToReactTo []byte,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Unmarshal message ID
	msgId := message.ID{}
	copy(msgId[:], messageToReactTo)

	// Send reaction
	msgID, rnd, ephID, err := dmc.api.SendReaction(&partnerPubKey,
		dmToken, reaction, msgId, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// Send is used to send a raw message via DM. In general, it
// should be wrapped in a function that defines the wire protocol. If the final
// message, before being sent over the wire, is too long, this will return an
// error. Due to the underlying encoding using compression, it isn't possible to
// define the largest payload that can be sent, but it will always be possible
// to send a payload of 802 bytes at minimum. The meaning of validUntil depends
// on the use case.
//
// Parameters:
//   - messageType - The message type of the message. This will be a valid
//     [channels.MessageType].
//   - partnerPubKeyBytes - the bytes of the public key of the partner's ed25519
//     signing key.
//   - dmToken - the token used to derive the reception id for the partner.
//   - message - The contents of the message. This need not be of data type
//     string, as the message could be a specified format that the channel may
//     recognize.
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the channels.Manager
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport.
func (dmc *DMClient) Send(messageType int, partnerPubKeyBytes []byte,
	dmToken uint32, message []byte, leaseTimeMS int64,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)
	msgTy := dm.MessageType(messageType)

	// Send message
	msgID, rnd, ephID, err := dmc.api.Send(&partnerPubKey,
		dmToken, msgTy, message, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// constructChannelSendReport is a helper function which returns a JSON
// marshalled ChannelSendReport.
func constructDMSendReport(dmMsgID message.ID,
	roundId id.Round, ephId ephemeral.Id) ([]byte, error) {
	// Construct send report
	sendReport := ChannelSendReport{
		MessageId:  dmMsgID.Bytes(),
		RoundsList: makeRoundsList(roundId),
		EphId:      ephId.Int64(),
	}

	// Marshal send report
	return json.Marshal(sendReport)
}

// Simple mux'd map list of clients.
var dmClients = &dmClientTracker{
	tracked: make(map[int]*DMClient),
	count:   0,
}

type dmClientTracker struct {
	tracked map[int]*DMClient
	count   int
	sync.RWMutex
}

func (dct *dmClientTracker) add(c dm.Client) *DMClient {
	dct.Lock()
	defer dct.Unlock()

	dmID := dct.count
	dct.count++

	dct.tracked[dmID] = &DMClient{
		api: c,
		id:  dmID,
	}

	return dct.tracked[dmID]
}
func (dct *dmClientTracker) get(id int) (*DMClient, error) {
	dct.RLock()
	defer dct.RUnlock()

	c, exist := dct.tracked[id]
	if !exist {
		return nil, errors.Errorf("DMClient ID %d does not exist",
			id)
	}

	return c, nil
}
func (dct *dmClientTracker) delete(id int) {
	dct.Lock()
	defer dct.Unlock()

	delete(dct.tracked, id)
	dct.count--
}
