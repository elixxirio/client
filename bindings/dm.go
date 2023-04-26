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
	"gitlab.com/elixxir/client/v4/dm/storage"
	"sync"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/client/v4/storage/utility"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/codename"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
)

// DMReceiverBuilder builds an event model.
type DMReceiverBuilder interface {
	Build(path string) DMReceiver
}

// DMClient is the bindings level interface for the direct messaging client. It
// implements all the [dm.Client] functions but converts from basic types that
// are friendly to gomobile and Javascript WASM interfaces (e.g., []byte, int,
// and string).
//
// Users of the bindings API can create multiple DMClient objects, which are
// tracked via a private singleton.
type DMClient struct {
	api dm.Client
	// NOTE: This matches the integer in the dmClientTracker singleton
	id int
}

// NewDMClient creates a new [DMClient] from a private identity
// ([codename.PrivateIdentity]), used for direct messaging.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [codename.GenerateIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// Parameters:
//   - cmixID - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - privateIdentity - Bytes of a private identity
//     ([codename.PrivateIdentity]) that is generated by
//     [codename.GenerateIdentity].
//   - receiverBuilder - An interface that contains a function that initialises
//     and returns an [EventModel] that is bindings-compatible.
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

	receptionID := dm.DeriveReceptionID(pi.PubKey, pi.GetDMToken())

	nickMgr := dm.NewNicknameManager(receptionID,
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

// NewDMClientWithGoEventModel creates a new [DMClient] from a private identity
// ([codename.PrivateIdentity]). This is not compatible with GoMobile bindings
// because it receives the go event model.
//
// This is for instantiating a manager for an identity. For generating
// a new identity, use [codename.GenerateIdentity]. You should instantiate
// every load as there is no load function and associated state in
// this module.
//
// Parameters:
//   - cmixID - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - privateIdentity - Bytes of a private identity
//     ([codename.PrivateIdentity]) that is generated by
//     [codename.GenerateIdentity].
//   - receiver - The [dm.EventModel].
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

	receptionID := dm.DeriveReceptionID(pi.PubKey, pi.GetDMToken())

	nickMgr := dm.NewNicknameManager(receptionID,
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

// DmReceivedCallback is called any time a DM is received or updated.
//
// messageUpdate is true if the DM message already exists and was edited.
// conversationUpdate is true if the DM conversation was created or modified.
type DmReceivedCallback interface {
	Callback(uuid int64, pubKey []byte,
		messageUpdate, conversationUpdate bool)
}

// NewDmManagerMobile loads an existing [DMClient] for the given storage
// tag backed with SqlLite for mobile use.
//
// Parameters:
//   - cmixID - ID of [Cmix] object in tracker. This can be retrieved using
//     [Cmix.GetID].
//   - privateIdentity - Bytes of a private identity
//     ([codename.PrivateIdentity]) that is generated by
//     [codename.GenerateIdentity].
//   - dbFilePath - absolute string path to the SqlLite database file
//   - cipherID - ID of [ChannelDbCipher] object in tracker.
//   - msgCb - Callback that is invoked whenever DM message is received/updated.
func NewDmManagerMobile(cmixID int, privateIdentity []byte,
	dbFilePath string, cipherID int, msgCb DmReceivedCallback) (*DMClient, error) {

	// Get user from singleton
	user, err := cmixTrackerSingleton.get(cmixID)
	if err != nil {
		return nil, err
	}

	cipher, err := channelDbCipherTrackerSingleton.get(cipherID)
	if err != nil {
		return nil, err
	}

	newMsgCb := func(uuid uint64, pubKey ed25519.PublicKey,
		messageUpdate, conversationUpdate bool) {
		msgCb.Callback(int64(uuid), pubKey, messageUpdate, conversationUpdate)
	}

	model, err := storage.NewEventModel(dbFilePath, cipher, newMsgCb)
	if err != nil {
		return nil, err
	}

	pi, err := codename.UnmarshalPrivateIdentity(privateIdentity)
	if err != nil {
		return nil, err
	}

	receptionID := dm.DeriveReceptionID(pi.PubKey, pi.GetDMToken())

	nickMgr := dm.NewNicknameManager(receptionID,
		user.api.GetStorage().GetKV())

	sendTracker := dm.NewSendTracker(user.api.GetStorage().GetKV())

	m := dm.NewDMClient(&pi, model, sendTracker, nickMgr,
		user.api.GetCmix(), user.api.GetRng())
	if err != nil {
		return nil, err
	}

	// Add channel to singleton and return
	return dmClients.add(m), nil
}

////////////////////////////////////////////////////////////////////////////////
// DM Share URL                                                          //
////////////////////////////////////////////////////////////////////////////////

// DMShareURL is returned from [DMClient.GetShareURL]. It includes the
// user's share URL.
//
// JSON example for a user:
//
//	{
//	 "url": "https://internet.speakeasy.tech/?l=32&m=5&p=EfDzQDa4fQ5BoqNIMbECFDY9ckRr_fadd8F1jE49qJc%3D&t=4231817746&v=1",
//	 "password": "hunter2",
//	}
type DMShareURL struct {
	URL      string `json:"url"`
	Password string `json:"password"`
}

// DMUser is returned from [DecodeDMShareURL]. It includes the token
// and public key of the user who created the URL.
//
// JSON example for a user:
//
//	{
//	 "token": 4231817746,
//	 "publicKey": "EfDzQDa4fQ5BoqNIMbECFDY9ckRr/fadd8F1jE49qJc="
//	}
type DMUser struct {
	Token     int32  `json:"token"`
	PublicKey []byte `json:"publicKey"`
}

// GetShareURL generates a URL that can be used to share a URL to initiate d
// direct messages with this user.
//
// Parameters:
//   - host - The URL to append the DM info to.
//
// Returns:
//   - JSON of [DMShareURL].
func (dmc *DMClient) GetShareURL(host string) ([]byte, error) {
	// todo: in a later ticket, RNG will be utilized for password protected DMs
	//  This note is for this ticketholder: RNG is part of the DMClient, but
	//  there is no accessor. Simply add the accessor to the interface and call
	//   dmc.GetRNG().GetStream.
	url, err := dm.ShareURL(
		host, 0, int32(dmc.api.GetToken()), dmc.api.GetPublicKey(), nil)
	if err != nil {
		return nil, err
	}

	su := DMShareURL{
		URL: url,
	}

	return json.Marshal(su)
}

// DecodeDMShareURL decodes the user's URL into a DMUser.
//
// Parameters:
//   - url - The user's share URL. Should be received from another user or
//     generated via [DMClient.GetShareURL].
//
// Returns:
//   - JSON of DMUser.
func DecodeDMShareURL(url string) ([]byte, error) {
	token, pubKey, err := dm.DecodeShareURL(url, "")
	if err != nil {
		return nil, err
	}

	dmShareReport := &DMUser{
		Token:     token,
		PublicKey: pubKey.Bytes(),
	}

	return json.Marshal(dmShareReport)
}

// GetID returns the tracker ID for the DMClient object.
func (dmc *DMClient) GetID() int {
	return dmc.id
}

// GetPublicKey returns the bytes of the public key for this client.
func (dmc *DMClient) GetPublicKey() []byte {
	return dmc.api.GetPublicKey().Bytes()
}

// GetToken returns the DM token of this client.
func (dmc *DMClient) GetToken() uint32 {
	return dmc.api.GetToken()
}

// GetIdentity returns the public identity associated with this client.
func (dmc *DMClient) GetIdentity() []byte {
	return dmc.api.GetIdentity().Marshal()
}

// ExportPrivateIdentity encrypts and exports the private identity to a portable
// string.
func (dmc *DMClient) ExportPrivateIdentity(password string) ([]byte, error) {
	return dmc.api.ExportPrivateIdentity(password)
}

// GetNickname gets the nickname associated with this DM user.
func (dmc *DMClient) GetNickname() (string, error) {
	nick, exists := dmc.api.GetNickname()
	if !exists {
		return "", errors.New("no nickname found")
	}
	return nick, nil
}

// SetNickname sets the nickname to use for this user.
func (dmc *DMClient) SetNickname(nick string) {
	dmc.api.SetNickname(nick)
}

// IsBlocked indicates if the given sender is blocked.
// Blocking is controlled by the receiver/EventModel.
func (dmc *DMClient) IsBlocked(senderPubKeyBytes []byte) bool {
	senderPubKey := ed25519.PublicKey(senderPubKeyBytes)
	return dmc.api.IsBlocked(senderPubKey)
}

// GetBlockedSenders returns the public keys of all senders who are blocked by
// this user. Blocking is controlled by the receiver/EventModel.
func (dmc *DMClient) GetBlockedSenders() []byte {
	blocked := dmc.api.GetBlockedSenders()
	blockedJSON, err := json.Marshal(blocked)
	if err != nil {
		jww.ERROR.Printf("Couldn't marshal blocked senders: %+v", err)
	}
	return blockedJSON
}

// SendText is used to send a formatted direct message to a user.
//
// Parameters:
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - message - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the [channels.Manager]
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - JSON of [xxdk.CMIXParams]. If left empty, then
//     [GetDefaultCMixParams] will be used internally.
//
// Returns:
//   - []byte - SON of [ChannelSendReport].
func (dmc *DMClient) SendText(partnerPubKeyBytes []byte, partnerToken int32,
	message string, leaseTimeMS int64, cmixParamsJSON []byte) ([]byte, error) {
	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Send message
	msgID, rnd, ephID, err := dmc.api.SendText(&partnerPubKey, uint32(partnerToken),
		message, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendReply is used to send a formatted direct message reply.
//
// If the message ID that the reply is sent to does not exist, then the other
// side will post the message as a normal message and not as a reply.
//
// The message will auto delete leaseTime after the round it is sent in, lasting
// forever if [ValidForever] is used.
//
// Parameters:
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - replyMessage - The contents of the reply message. The message should be
//     at most 510 bytes. This is expected to be Unicode, and thus a string data
//     type is expected
//   - replyToBytes - The bytes of the [message.ID] of the message you wish to
//     reply to. This may be found in the [ChannelSendReport] if replying to
//     your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the [ChannelMessageReceptionCallback] registered using
//     [ChannelsManager.RegisterReceiveHandler].
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the [channels.Manager]
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - JSON of [xxdk.CMIXParams]. If left empty, then
//     [GetDefaultCMixParams] will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport
func (dmc *DMClient) SendReply(partnerPubKeyBytes []byte, partnerToken int32,
	replyMessage string, replyToBytes []byte, leaseTimeMS int64,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Unmarshal message ID
	replyTo, err := message.UnmarshalID(replyToBytes)
	if err != nil {
		return nil, err
	}

	// Send Reply
	msgID, rnd, ephID, err := dmc.api.SendReply(&partnerPubKey,
		uint32(partnerToken), replyMessage, replyTo, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendReaction is used to send a reaction to a direct message.
// The reaction must be a single emoji with no other characters,
// and will be rejected otherwise.
//
// Clients will drop the reaction if they do not recognize the reactTo message.
//
// Parameters:
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - reaction - The user's reaction. This should be a single emoji with no
//     other characters. As such, a Unicode string is expected.
//   - reactToBytes - The bytes of the [message.ID] of the message you wish to
//     react to. This may be found in the [ChannelSendReport] if replying to
//     your own. Alternatively, if reacting to another user's message, you may
//     retrieve it via the [ChannelMessageReceptionCallback] registered using
//     [ChannelsManager.RegisterReceiveHandler].
//   - cmixParamsJSON - JSON of [xxdk.CMIXParams]. If left empty, then
//     [GetDefaultCMixParams] will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport.
func (dmc *DMClient) SendReaction(partnerPubKeyBytes []byte, partnerToken int32,
	reaction string, reactToBytes []byte,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Unmarshal message ID
	reactTo, err := message.UnmarshalID(reactToBytes)
	if err != nil {
		return nil, err
	}

	// Send reaction
	msgID, rnd, ephID, err := dmc.api.SendReaction(&partnerPubKey,
		uint32(partnerToken), reaction, reactTo, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendInvite is used to send to a DM partner an invitation to another
// channel.
//
// Parameters:
//   - channelsManagerId - ID of [ChannelsManager] object in tracker. This can
//     be retrieved using [ChannelsManager.GetID].
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - inviteToChannelBytes - Marshalled bytes of the channel the user is
//     inviting another user to.
//   - message - The contents of the message. The message should be at most 510
//     bytes. This is expected to be Unicode, and thus a string data type is
//     expected.
//   - host - The URL to append the channel info to.
//   - maxUses - The maximum number of uses the link can be used (0 for
//     unlimited).
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
func (dmc *DMClient) SendInvite(channelsManagerId int, partnerPubKeyBytes []byte,
	partnerToken int32, inviteToChannelBytes []byte, message string,
	host string, maxUses int, cmixParamsJSON []byte) ([]byte, error) {

	chanMan, err := channelManagerTrackerSingleton.get(channelsManagerId)
	if err != nil {
		return nil, err
	}

	// Unmarshal channel ID
	inviteToID, err := id.Unmarshal(inviteToChannelBytes)
	if err != nil {
		return nil, err
	}

	// Unmarshal cmix params
	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}

	// Retrieve channel from manager
	inviteTo, err := chanMan.api.GetChannel(inviteToID)
	if err != nil {
		return nil, err
	}

	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	// Send invite
	msgID, rnd, ephID, err := dmc.api.SendInvite(&partnerPubKey,
		uint32(partnerToken), message, inviteTo, host, maxUses, params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// SendSilent is used to send to a channel a message with no notifications.
// Its primary purpose is to communicate new nicknames without calling
// SendMessage.
//
// It takes no payload intentionally as the message should be very
// lightweight.
//
// Parameters:
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - cmixParamsJSON - A JSON marshalled [xxdk.CMIXParams]. This may be empty,
//     and GetDefaultCMixParams will be used internally.
func (dmc *DMClient) SendSilent(partnerPubKeyBytes []byte,
	partnerToken int32, cmixParamsJSON []byte) ([]byte, error) {
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}

	msgID, rnd, ephID, err := dmc.api.SendSilent(&partnerPubKey, uint32(partnerToken),
		params.CMIX)
	if err != nil {
		return nil, err
	}

	// Construct send report
	return constructDMSendReport(msgID, rnd.ID, ephID)
}

// Send is used to send a raw message. In general, it
// should be wrapped in a function that defines the wire protocol.
//
// If the final message, before being sent over the wire, is too long, this will
// return an error. Due to the underlying encoding using compression, it is not
// possible to define the largest payload that can be sent, but it will always
// be possible to send a payload of 802 bytes at minimum.
//
// The meaning of leaseTimeMS depends on the use case.
//
// Parameters:
//   - partnerPubKeyBytes - The bytes of the public key of the partner's ED25519
//     signing key.
//   - partnerToken - The token used to derive the reception ID for the partner.
//   - messageType - The message type of the message. This will be a valid
//     [dm.MessageType].
//   - plaintext - The contents of the message. This need not be of data type
//     string, as the message could be a specified format that the channel may
//     recognize.
//   - leaseTimeMS - The lease of the message. This will be how long the message
//     is valid until, in milliseconds. As per the [channels.Manager]
//     documentation, this has different meanings depending on the use case.
//     These use cases may be generic enough that they will not be enumerated
//     here.
//   - cmixParamsJSON - JSON of [xxdk.CMIXParams]. If left empty, then
//     [GetDefaultCMixParams] will be used internally.
//
// Returns:
//   - []byte - A JSON marshalled ChannelSendReport.
func (dmc *DMClient) Send(partnerPubKeyBytes []byte,
	partnerToken int32, messageType int, plaintext []byte, leaseTimeMS int64,
	cmixParamsJSON []byte) ([]byte, error) {

	params, err := parseCMixParams(cmixParamsJSON)
	if err != nil {
		return nil, err
	}
	partnerPubKey := ed25519.PublicKey(partnerPubKeyBytes)
	msgTy := dm.MessageType(messageType)

	// Send message
	msgID, rnd, ephID, err := dmc.api.Send(&partnerPubKey,
		uint32(partnerToken), msgTy, plaintext, params.CMIX)
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
		MessageID:  dmMsgID.Bytes(),
		RoundsList: makeRoundsList(roundId),
		EphId:      ephId.Int64(),
	}

	// Marshal send report
	return json.Marshal(sendReport)
}

func GetDMInstance(instanceID int) (*DMClient, error) {
	instance, ok := dmClients.tracked[instanceID]
	if !ok {
		return nil, errors.Errorf("no dm instance id: %d", instanceID)
	}
	return instance, nil
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

////////////////////////////////////////////////////////////////////////////////
// DM DMDbCipher                                                             //
////////////////////////////////////////////////////////////////////////////////

// DMDbCipher is the bindings layer representation of the [DM.Cipher].
type DMDbCipher struct {
	api  cryptoChannel.Cipher
	salt []byte
	id   int
}

// DMDbCipherTrackerSingleton is used to track DMDbCipher objects
// so that they can be referenced by ID back over the bindings.
var DMDbCipherTrackerSingleton = &DMDbCipherTracker{
	tracked: make(map[int]*DMDbCipher),
	count:   0,
}

// DMDbCipherTracker is a singleton used to keep track of extant
// DMDbCipher objects, preventing race conditions created by passing it
// over the bindings.
type DMDbCipherTracker struct {
	tracked map[int]*DMDbCipher
	count   int
	mux     sync.RWMutex
}

// create creates a DMDbCipher from a [DM.Cipher], assigns it a unique
// ID, and adds it to the DMDbCipherTracker.
func (ct *DMDbCipherTracker) create(c cryptoChannel.Cipher) *DMDbCipher {
	ct.mux.Lock()
	defer ct.mux.Unlock()

	chID := ct.count
	ct.count++

	ct.tracked[chID] = &DMDbCipher{
		api: c,
		id:  chID,
	}

	return ct.tracked[chID]
}

// get an DMDbCipher from the DMDbCipherTracker given its ID.
func (ct *DMDbCipherTracker) get(id int) (*DMDbCipher, error) {
	ct.mux.RLock()
	defer ct.mux.RUnlock()

	c, exist := ct.tracked[id]
	if !exist {
		return nil, errors.Errorf(
			"Cannot get DMDbCipher for ID %d, does not exist", id)
	}

	return c, nil
}

// delete removes a DMDbCipher from the DMDbCipherTracker.
func (ct *DMDbCipherTracker) delete(id int) {
	ct.mux.Lock()
	defer ct.mux.Unlock()

	delete(ct.tracked, id)
}

// GetDMDbCipherTrackerFromID returns the DMDbCipher with the
// corresponding ID in the tracker.
func GetDMDbCipherTrackerFromID(id int) (*DMDbCipher, error) {
	return DMDbCipherTrackerSingleton.get(id)
}

// NewDMsDatabaseCipher constructs a DMDbCipher object.
//
// Parameters:
//   - cmixID - The tracked [Cmix] object ID.
//   - password - The password for storage. This should be the same password
//     passed into [NewCmix].
//   - plaintTextBlockSize - The maximum size of a payload to be encrypted.
//     A payload passed into [DMDbCipher.Encrypt] that is larger than
//     plaintTextBlockSize will result in an error.
func NewDMsDatabaseCipher(cmixID int, password []byte,
	plaintTextBlockSize int) (*DMDbCipher, error) {
	// Get user from singleton
	user, err := cmixTrackerSingleton.get(cmixID)
	if err != nil {
		return nil, err
	}

	// Generate RNG
	stream := user.api.GetRng().GetStream()

	// Load or generate a salt
	salt, err := utility.NewOrLoadSalt(
		user.api.GetStorage().GetKV(), stream)
	if err != nil {
		return nil, err
	}

	// Construct a cipher
	c, err := cryptoChannel.NewCipher(password, salt,
		plaintTextBlockSize, stream)
	if err != nil {
		return nil, err
	}

	// Return a cipher
	return DMDbCipherTrackerSingleton.create(c), nil
}

// GetID returns the ID for this DMDbCipher in the DMDbCipherTracker.
func (c *DMDbCipher) GetID() int {
	return c.id
}

// Encrypt will encrypt the raw data. It will return a ciphertext. Padding is
// done on the plaintext so all encrypted data looks uniform at rest.
//
// Parameters:
//   - plaintext - The data to be encrypted. This must be smaller than the block
//     size passed into [NewDMsDatabaseCipher]. If it is larger, this will
//     return an error.
func (c *DMDbCipher) Encrypt(plaintext []byte) ([]byte, error) {
	return c.api.Encrypt(plaintext)
}

// Decrypt will decrypt the passed in encrypted value. The plaintext will
// be returned by this function. Any padding will be discarded within
// this function.
//
// Parameters:
//   - ciphertext - the encrypted data returned by [DMDbCipher.Encrypt].
func (c *DMDbCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return c.api.Decrypt(ciphertext)
}

// MarshalJSON marshals the cipher into valid JSON. This function adheres to the
// json.Marshaler interface.
func (c *DMDbCipher) MarshalJSON() ([]byte, error) {
	return c.api.MarshalJSON()
}

// UnmarshalJSON unmarshalls JSON into the cipher. This function adheres to the
// json.Unmarshaler interface.
//
// Note that this function does not transfer the internal RNG. Use
// NewCipherFromJSON to properly reconstruct a cipher from JSON.
func (c *DMDbCipher) UnmarshalJSON(data []byte) error {
	return c.api.UnmarshalJSON(data)
}
