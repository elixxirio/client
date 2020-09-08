////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package api

import (
	"bufio"
	"crypto"
	gorsa "crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/bots"
	"gitlab.com/elixxir/client/cmixproto"
	clientcrypto "gitlab.com/elixxir/client/crypto"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/client/network"
	"gitlab.com/elixxir/client/keyStore"
	"gitlab.com/elixxir/client/parse"
	"gitlab.com/elixxir/client/rekey"
	"gitlab.com/elixxir/client/storage"
	"gitlab.com/elixxir/client/user"
	"gitlab.com/elixxir/client/userRegistry"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/large"
	"gitlab.com/elixxir/primitives/switchboard"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/crypto/tls"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/ndf"
	goio "io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type Client struct {
	storage             globals.Storage
	session             user.Session
	sessionV2           *storage.Session
	receptionManager    *network.ReceptionManager
	switchboard         *switchboard.Switchboard
	ndf                 *ndf.NetworkDefinition
	topology            *connect.Circuit
	opStatus            OperationProgressCallback
	rekeyChan           chan struct{}
	quitChan            chan struct{}
	registrationVersion string

	// Pointer to a send function, which allows testing to override the default
	// using NewTestClient
	sendFunc sender
}

// Type that defines what the default and any testing send functions should look like
type sender func(message parse.MessageInterface, rm *network.ReceptionManager, session user.Session, topology *connect.Circuit, host *connect.Host) error

//used to report the state of registration
type OperationProgressCallback func(int)

// Creates a new client with the default send function
func NewClient(s globals.Storage, locA, locB string, ndfJSON *ndf.NetworkDefinition) (*Client, error) {
	return newClient(s, locA, locB, ndfJSON, send)
}

// Creates a new test client with an overridden send function
func NewTestClient(s globals.Storage, locA, locB string, ndfJSON *ndf.NetworkDefinition, i interface{}, sendFunc sender) (*Client, error) {
	switch i.(type) {
	case *testing.T:
		break
	case *testing.M:
		break
	case *testing.B:
		break
	default:
		globals.Log.FATAL.Panicf("GenerateId is restricted to testing only. Got %T", i)
	}
	return newClient(s, locA, locB, ndfJSON, sendFunc)
}

// setStorage is a helper to initialize the new session storage
func (cl *Client) setStorage(locA, password string) error {
	// TODO: FIX ME
	// While the old session is still valid, we are using the LocA storage to initialize the session
	dirname := filepath.Dir(locA)
	//FIXME: We need to accept the user's password here!
	var err error
	network.SessionV2, err = storage.Init(dirname, password)
	if err != nil {
		return errors.Wrapf(err, "could not initialize v2 "+
			"storage at %s", locA)
	}
	clientcrypto.SessionV2 = network.SessionV2
	cl.sessionV2 = network.SessionV2

	// FIXME: Client storage must have regstate set
	_, err = cl.sessionV2.GetRegState()
	if os.IsNotExist(err) {
		cl.sessionV2.SetRegState(user.KeyGenComplete)
	}

	return nil
}

// Creates a new Client using the storage mechanism provided.
// If none is provided, a default storage using OS file access
// is created
// returns a new Client object, and an error if it fails
func newClient(s globals.Storage, locA, locB string, ndfJSON *ndf.NetworkDefinition, sendFunc sender) (*Client, error) {
	var store globals.Storage
	if s == nil {
		globals.Log.INFO.Printf("No storage provided," +
			" initializing Client with default storage")
		store = &globals.DefaultStorage{}
	} else {
		store = s
	}

	err := store.SetLocation(locA, locB)

	if err != nil {
		err = errors.New("Invalid Local Storage Location: " + err.Error())
		globals.Log.ERROR.Printf(err.Error())
		return nil, err
	}

	cl := new(Client)
	cl.storage = store
	cl.ndf = ndfJSON
	cl.sendFunc = sendFunc

	//Create the cmix group and init the registry
	cmixGrp := cyclic.NewGroup(
		large.NewIntFromString(cl.ndf.CMIX.Prime, 16),
		large.NewIntFromString(cl.ndf.CMIX.Generator, 16))
	userRegistry.InitUserRegistry(cmixGrp)

	cl.opStatus = func(int) {
		return
	}

	cl.switchboard = switchboard.NewSwitchboard()

	cl.rekeyChan = make(chan struct{}, 1)
	cl.quitChan = make(chan struct{}) // Blocking is intentional

	return cl, nil
}

// LoadSession loads the session object for the UID
func (cl *Client) Login(password string) (*id.ID, error) {

	var session user.Session
	var err error
	done := make(chan struct{})

	// run session loading in a separate goroutine so if it panics it can
	// be caught and an error can be returned
	go func() {
		defer func() {
			if r := recover(); r != nil {
				globals.Log.ERROR.Println("Session file loading crashed")
				err = sessionFileError
				done <- struct{}{}
			}
		}()

		session, err = user.LoadSession(cl.storage, password)
		done <- struct{}{}
	}()

	//wait for session file loading to complete
	<-done

	if err != nil {
		return nil, errors.Wrap(err, "Login: Could not login")
	}

	if session == nil {
		return nil, errors.New("Unable to load session, no error reported")
	}

	cl.session = session
	locA, _ := cl.storage.GetLocation()
	err = cl.setStorage(locA, password)
	if err != nil {
		return nil, err
	}

	regState, err := cl.sessionV2.GetRegState()
	if err != nil {
		return nil, errors.Wrap(err,
			"Login: Could not login: Could not get regState")
	}
	userData, err := cl.sessionV2.GetUserData()
	if err != nil {
		return nil, errors.Wrap(err,
			"Login: Could not login: Could not get userData")
	}

	if regState <= user.KeyGenComplete {
		return nil, errors.New("Cannot log a user in which has not " +
			"completed registration ")
	}

	newRm, err := network.NewReceptionManager(cl.rekeyChan, cl.quitChan,
		userData.ThisUser.User,
		rsa.CreatePrivateKeyPem(userData.RSAPrivateKey),
		rsa.CreatePublicKeyPem(userData.RSAPublicKey),
		userData.Salt, cl.switchboard)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new reception manager")
	}
	newRm.Comms.Manager = cl.receptionManager.Comms.Manager
	cl.receptionManager = newRm
	cl.session.SetE2EGrp(userData.E2EGrp)
	cl.session.SetUser(userData.ThisUser.User)
	return userData.ThisUser.User, nil
}

// Logout closes the connection to the server and the messageReceiver and clears out the client values,
// so we can effectively shut everything down.  at this time it does
// nothing with the user id. In the future this will release resources
// and safely release any sensitive memory. Recommended time out is 500ms.
func (cl *Client) Logout(timeoutDuration time.Duration) error {
	if cl.session == nil {
		err := errors.New("Logout: Cannot Logout when you are not logged in")
		globals.Log.ERROR.Printf(err.Error())
		return err
	}

	// Here using a select statement and the fact that making
	// cl.ReceptionQuitChan is blocking, we can detect when
	// killing the reception manager is taking too long and we use
	// the time out to stop the attempt and return an error.
	timer := time.NewTimer(timeoutDuration)
	select {
	case cl.quitChan <- struct{}{}:
		cl.receptionManager.Comms.DisconnectAll()
	case <-timer.C:
		return errors.Errorf("Message receiver shut down timed out after %s ms", timeoutDuration)
	}

	// Store the user session files before logging out
	errStore := cl.session.StoreSession()
	if errStore != nil {
		err := errors.New(fmt.Sprintf("Logout: Store Failed: %s" +
			errStore.Error()))
		globals.Log.ERROR.Printf(err.Error())
		return err
	}

	// Clear all keys from ram
	errImmolate := cl.session.Immolate()
	cl.session = nil
	if errImmolate != nil {
		err := errors.New(fmt.Sprintf("Logout: Immolation Failed: %s" +
			errImmolate.Error()))
		globals.Log.ERROR.Printf(err.Error())
		return err
	}

	// Here we clear away all state in the client struct that should not be persistent
	cl.session = nil
	cl.receptionManager = nil
	cl.topology = nil
	cl.registrationVersion = ""

	return nil
}

// VerifyNDF verifies the signature of the network definition file (NDF) and
// returns the structure. Panics when the NDF string cannot be decoded and when
// the signature cannot be verified. If the NDF public key is empty, then the
// signature verification is skipped and warning is printed.
func VerifyNDF(ndfString, ndfPub string) *ndf.NetworkDefinition {
	// If there is no public key, then skip verification and print warning
	if ndfPub == "" {
		globals.Log.WARN.Printf("Running without signed network " +
			"definition file")
	} else {
		ndfReader := bufio.NewReader(strings.NewReader(ndfString))
		ndfData, err := ndfReader.ReadBytes('\n')
		ndfData = ndfData[:len(ndfData)-1]
		if err != nil {
			globals.Log.FATAL.Panicf("Could not read NDF: %v", err)
		}
		ndfSignature, err := ndfReader.ReadBytes('\n')
		if err != nil {
			globals.Log.FATAL.Panicf("Could not read NDF Sig: %v",
				err)
		}
		ndfSignature, err = base64.StdEncoding.DecodeString(
			string(ndfSignature[:len(ndfSignature)-1]))
		if err != nil {
			globals.Log.FATAL.Panicf("Could not read NDF Sig: %v",
				err)
		}
		// Load the TLS cert given to us, and from that get the RSA public key
		cert, err := tls.LoadCertificate(ndfPub)
		if err != nil {
			globals.Log.FATAL.Panicf("Could not load public key: %v", err)
		}
		pubKey := &rsa.PublicKey{PublicKey: *cert.PublicKey.(*gorsa.PublicKey)}

		// Hash NDF JSON
		rsaHash := sha256.New()
		rsaHash.Write(ndfData)

		globals.Log.INFO.Printf("%s \n::\n %s",
			ndfSignature, ndfData)

		// Verify signature
		err = rsa.Verify(
			pubKey, crypto.SHA256, rsaHash.Sum(nil), ndfSignature, nil)

		if err != nil {
			globals.Log.FATAL.Panicf("Could not verify NDF: %v", err)
		}
	}

	ndfJSON, _, err := ndf.DecodeNDF(ndfString)
	if err != nil {
		globals.Log.FATAL.Panicf("Could not decode NDF: %v", err)
	}
	return ndfJSON
}

func (cl *Client) GetRegistrationVersion() string { // on client
	return cl.registrationVersion
}

//GetNDF returns the clients ndf
func (cl *Client) GetNDF() *ndf.NetworkDefinition {
	return cl.ndf
}

func (cl *Client) SetOperationProgressCallback(rpc OperationProgressCallback) {
	cl.opStatus = func(i int) { go rpc(i) }
}

// Populates a text message and returns its wire representation
// TODO support multi-type messages or telling if a message is too long?
func FormatTextMessage(message string) []byte {
	textMessage := cmixproto.TextMessage{
		Color:   -1,
		Message: message,
		Time:    time.Now().Unix(),
	}

	wireRepresentation, _ := proto.Marshal(&textMessage)
	return wireRepresentation
}

var sessionFileError = errors.New("Session file cannot be loaded and " +
	"is possibly corrupt. Please contact support@xxmessenger.io")

func (cl *Client) InitListeners() error {
	transmitGateway, err := id.Unmarshal(cl.ndf.Gateways[0].ID)
	if err != nil {
		globals.Log.DEBUG.Printf("%s: Gateways are: %+v", err.Error(),
			cl.ndf.Gateways)
		return err
	}
	transmissionHost, ok := cl.receptionManager.Comms.GetHost(transmitGateway)
	if !ok {
		return errors.New("Failed to retrieve host for transmission")
	}

	// Initialize UDB and nickname "bot" stuff here
	bots.InitBots(cl.session, *cl.sessionV2, cl.receptionManager,
		cl.topology, transmissionHost)
	// Initialize Rekey listeners
	rekey.InitRekey(cl.session, *cl.sessionV2, cl.receptionManager,
		cl.topology, transmissionHost, cl.rekeyChan)
	return nil
}

// Logs in user and sets session on client object
// returns the nickname or error if login fails
func (cl *Client) StartMessageReceiver(callback func(error)) error {
	pollWaitTimeMillis := 500 * time.Millisecond
	// TODO Don't start the message receiver if it's already started.
	// Should be a pretty rare occurrence except perhaps for mobile.
	receptionGateway, err := id.Unmarshal(cl.ndf.Gateways[len(cl.ndf.Gateways)-1].ID)
	if err != nil {
		return err
	}
	receptionHost, ok := cl.receptionManager.Comms.GetHost(receptionGateway)
	if !ok {
		return errors.New("Failed to retrieve host for transmission")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				globals.Log.ERROR.Println("Message Receiver Panicked: ", r)
				time.Sleep(1 * time.Second)
				go func() {
					callback(errors.New(fmt.Sprintln("Message Receiver Panicked", r)))
				}()
			}
		}()
		cl.receptionManager.MessageReceiver(cl.session, pollWaitTimeMillis, receptionHost, callback)
	}()

	return nil
}

// Default send function, can be overridden for testing
func (cl *Client) Send(message parse.MessageInterface) error {
	transmitGateway, err := id.Unmarshal(cl.ndf.Gateways[0].ID)
	if err != nil {
		return err
	}
	transmitGateway.SetType(id.Gateway)
	host, ok := cl.receptionManager.Comms.GetHost(transmitGateway)
	if !ok {
		return errors.New("Failed to retrieve host for transmission")
	}

	return cl.sendFunc(message, cl.receptionManager, cl.session, cl.topology, host)
}

// Send prepares and sends a message to the cMix network
func send(message parse.MessageInterface, rm *network.ReceptionManager, session user.Session, topology *connect.Circuit, host *connect.Host) error {
	recipientID := message.GetRecipient()
	cryptoType := message.GetCryptoType()
	return rm.SendMessage(session, topology, recipientID, cryptoType, message.Pack(), host)
}

// DisableBlockingTransmission turns off blocking transmission, for
// use with the channel bot and dummy bot
func (cl *Client) DisableBlockingTransmission() {
	cl.receptionManager.DisableBlockingTransmission()
}

// SetRateLimiting sets the minimum amount of time between message
// transmissions just for testing, probably to be removed in production
func (cl *Client) SetRateLimiting(limit uint32) {
	cl.receptionManager.SetRateLimit(time.Duration(limit) * time.Millisecond)
}

func (cl *Client) Listen(user *id.ID, messageType int32, newListener switchboard.Listener) string {
	listenerId := cl.GetSwitchboard().
		Register(user, messageType, newListener)
	globals.Log.INFO.Printf("Listening now: user %v, message type %v, id %v",
		user, messageType, listenerId)
	return listenerId
}

func (cl *Client) StopListening(listenerHandle string) {
	cl.GetSwitchboard().Unregister(listenerHandle)
}

func (cl *Client) GetSwitchboard() *switchboard.Switchboard {
	return cl.switchboard
}

func (cl *Client) GetUsername() string {
	userData, _ := cl.sessionV2.GetUserData()
	return userData.ThisUser.Username
}

func (cl *Client) GetCurrentUser() *id.ID {
	userData, _ := cl.sessionV2.GetUserData()
	return userData.ThisUser.User
}

// FIXME: This is not exactly thread safe but is unlikely
// to cause issues as username is not changed after
// registration. Needs serious design considerations.
func (cl *Client) ChangeUsername(username string) error {
	userData, err := cl.sessionV2.GetUserData()
	if err != nil {
		return err
	}
	userData.ThisUser.Username = username
	return cl.sessionV2.CommitUserData(userData)
}

func (cl *Client) GetKeyParams() *keyStore.KeyParams {
	return cl.session.GetKeyStore().GetKeyParams()
}

// Returns the local version of the client repo
func GetLocalVersion() string {
	return globals.SEMVER
}

type SearchCallback interface {
	Callback(userID, pubKey []byte, err error)
}

// UDB Search API
// Pass a callback function to extract results
func (cl *Client) SearchForUser(emailAddress string,
	cb SearchCallback, timeout time.Duration) {
	//see if the user has been searched before, if it has, return it
	contact, err := cl.sessionV2.GetContactByEmail(emailAddress)

	// if we successfully got the contact, return it.
	// errors can include the email address not existing,
	// so errors from the GetContact call are ignored
	if contact != nil && err == nil {
		cb.Callback(contact.Id.Bytes(), contact.PublicKey, nil)
		return
	}

	valueType := "EMAIL"
	go func() {
		contact, err := bots.Search(valueType, emailAddress, cl.opStatus, timeout)
		if err == nil && contact.Id != nil && contact.PublicKey != nil {
			cl.opStatus(globals.UDB_SEARCH_BUILD_CREDS)
			err = cl.registerUserE2E(contact)
			if err != nil {
				cb.Callback(contact.Id.Bytes(), contact.PublicKey, err)
				return
			}

			// FIXME: remove this once key manager is moved to new
			//        session
			err = cl.session.StoreSession()
			if err != nil {
				cb.Callback(contact.Id.Bytes(),
					contact.PublicKey, err)
				return
			}
			//store the user so future lookups can find it
			err = cl.sessionV2.SetContactByEmail(emailAddress,
				contact)

			// If there is something in the channel then send it; otherwise,
			// skip over it
			select {
			case cl.rekeyChan <- struct{}{}:
			default:
			}

			cb.Callback(contact.Id.Bytes(), contact.PublicKey, err)

		} else {
			if err == nil {
				globals.Log.INFO.Printf("UDB Search for email %s failed: user not found", emailAddress)
				err = errors.New("user not found in UDB")
				cb.Callback(nil, nil, err)
			} else {
				globals.Log.INFO.Printf("UDB Search for email %s failed: %+v", emailAddress, err)
				cb.Callback(nil, nil, err)
			}

		}
	}()
}

type NickLookupCallback interface {
	Callback(nick string, err error)
}

func (cl *Client) DeleteUser(u *id.ID) (string, error) {

	//delete from session
	// FIXME: I believe this used to return the user name of the deleted
	// user and the way we are calling this won't work since it is based on
	// user name and not User ID.
	user := cl.sessionV2.GetContactByID(u)
	err1 := cl.sessionV2.DeleteContactByID(u)

	email := ""
	if user != nil {
		email = user.Email
	}

	//delete from keystore
	err2 := cl.session.GetKeyStore().DeleteContactKeys(u)

	if err1 == nil && err2 == nil {
		return email, nil
	}

	if err1 != nil && err2 == nil {
		return email, errors.Wrap(err1, "Failed to remove from value store")
	}

	if err1 == nil && err2 != nil {
		return email, errors.Wrap(err2, "Failed to remove from key store")
	}

	if err1 != nil && err2 != nil {
		return email, errors.Wrap(fmt.Errorf("%s\n%s", err1, err2),
			"Failed to remove from key store and value store")
	}

	return email, nil

}

// Nickname lookup API
// Non-blocking, once the API call completes, the callback function
// passed as argument is called
func (cl *Client) LookupNick(user *id.ID,
	cb NickLookupCallback) {
	go func() {
		nick, err := bots.LookupNick(user)
		if err != nil {
			globals.Log.INFO.Printf("Lookup for nickname for user %+v failed", user)
		}
		cb.Callback(nick, err)
	}()
}

//Message struct adherent to interface in bindings for data return from ParseMessage
type ParsedMessage struct {
	Typed   int32
	Payload []byte
}

func (p ParsedMessage) GetSender() []byte {
	return []byte{}
}

func (p ParsedMessage) GetPayload() []byte {
	return p.Payload
}

func (p ParsedMessage) GetRecipient() []byte {
	return []byte{}
}

func (p ParsedMessage) GetMessageType() int32 {
	return p.Typed
}

func (p ParsedMessage) GetTimestampNano() int64 {
	return 0
}

func (p ParsedMessage) GetTimestamp() int64 {
	return 0
}

// Parses a passed message.  Allows a message to be aprsed using the interal parser
// across the API
func ParseMessage(message []byte) (ParsedMessage, error) {
	tb, err := parse.Parse(message)

	pm := ParsedMessage{}

	if err != nil {
		return pm, err
	}

	pm.Payload = tb.Body
	pm.Typed = int32(tb.MessageType)

	return pm, nil
}

func (cl *Client) GetSessionData() ([]byte, error) {
	return cl.session.GetSessionData()
}

// Set the output of the
func SetLogOutput(w goio.Writer) {
	globals.Log.SetLogOutput(w)
}

// GetSession returns the session object for external access.  Access at yourx
// own risk
func (cl *Client) GetSession() user.Session {
	return cl.session
}

// GetSessionV2 returns the session object for external access.  Access at yourx
// own risk
func (cl *Client) GetSessionV2() *storage.Session {
	return cl.sessionV2
}

// ReceptionManager returns the comm manager object for external access.  Access
// at your own risk
func (cl *Client) GetCommManager() *network.ReceptionManager {
	return cl.receptionManager
}

// LoadSessionText: load the encrypted session as a string
func (cl *Client) LoadEncryptedSession() (string, error) {
	encryptedSession, err := cl.GetSession().LoadEncryptedSession(cl.storage)
	if err != nil {
		return "", err
	}
	//Encode session to bas64 for useability
	encodedSession := base64.StdEncoding.EncodeToString(encryptedSession)

	return encodedSession, nil
}

//WriteToSession: Writes an arbitrary string to the session file
// Takes in a string that is base64 encoded (meant to be output of LoadEncryptedSession)
func (cl *Client) WriteToSessionFile(replacement string, store globals.Storage) error {
	//This call must not occur prior to a newClient call, thus check that client has been initialized
	if cl.ndf == nil || cl.topology == nil {
		errMsg := errors.Errorf("Cannot write to session if client hasn't been created yet")
		return errMsg
	}
	//Decode the base64 encoded replacement string (assumed to be encoded form LoadEncryptedSession)
	decodedSession, err := base64.StdEncoding.DecodeString(replacement)
	if err != nil {
		errMsg := errors.Errorf("Failed to decode replacment string: %+v", err)
		return errMsg
	}
	//Write the new session data to both locations
	err = user.WriteToSession(decodedSession, store)
	if err != nil {
		errMsg := errors.Errorf("Failed to store session: %+v", err)
		return errMsg
	}

	return nil
}
