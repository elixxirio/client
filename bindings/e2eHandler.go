////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/catalog"
	"gitlab.com/elixxir/client/v4/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/primitives/id"
)

// E2ESendReport is the bindings' representation of the return values of
// SendE2E.
//
// E2ESendReport Example JSON:
//
//	 {
//			"Rounds": [ 1, 4, 9],
//	     "RoundURL":"https://dashboard.xx.network/rounds/25?xxmessenger=true",
//			"MessageID": "iM34yCIr4Je8ZIzL9iAAG1UWAeDiHybxMTioMAaezvs=",
//			"Timestamp": 1661532254302612000,
//			"KeyResidue": "9q2/A69EAuFM1hFAT7Bzy5uGOQ4T6bPFF72h5PlgCWE="
//	 }
type E2ESendReport struct {
	RoundsList
	RoundURL   string
	MessageID  []byte
	Timestamp  int64
	KeyResidue []byte
}

// GetReceptionID returns the marshalled default IDs.
//
// Returns:
//   - []byte - the marshalled bytes of the id.ID object.
func (e *E2e) GetReceptionID() []byte {
	return e.api.GetE2E().GetReceptionID().Marshal()
}

// DeleteContact removes a partner from E2e's storage.
//
// Parameters:
//   - partnerID - the marshalled bytes of id.ID.
func (e *E2e) DeleteContact(partnerID []byte) error {
	partner, err := id.Unmarshal(partnerID)
	if err != nil {
		return err
	}

	return e.api.DeleteContact(partner)
}

// GetAllPartnerIDs returns a list of all partner IDs that the user has an E2E
// relationship with.
//
// Returns:
//   - []byte - the marshalled bytes of []*id.ID.
func (e *E2e) GetAllPartnerIDs() ([]byte, error) {
	return json.Marshal(e.api.GetE2E().GetAllPartnerIDs())
}

// PayloadSize returns the max payload size for a partitionable E2E message.
func (e *E2e) PayloadSize() int {
	return int(e.api.GetE2E().PayloadSize())
}

// SecondPartitionSize returns the max partition payload size for all payloads
// after the first payload.
func (e *E2e) SecondPartitionSize() int {
	return int(e.api.GetE2E().SecondPartitionSize())
}

// PartitionSize returns the partition payload size for the given payload index.
// The first payload is index 0.
func (e *E2e) PartitionSize(payloadIndex int) int {
	return int(e.api.GetE2E().PartitionSize(uint(payloadIndex)))
}

// FirstPartitionSize returns the max partition payload size for the first
// payload.
func (e *E2e) FirstPartitionSize() int {
	return int(e.api.GetE2E().FirstPartitionSize())
}

// GetHistoricalDHPrivkey returns the user's marshalled historical DH private
// key.
//
// Returns:
//   - []byte - the marshalled bytes of the cyclic.Int object.
func (e *E2e) GetHistoricalDHPrivkey() ([]byte, error) {
	return e.api.GetE2E().GetHistoricalDHPrivkey().MarshalJSON()
}

// GetHistoricalDHPubkey returns the user's marshalled historical DH public key.
//
// Returns:
//   - []byte - the marshalled bytes of the cyclic.Int object.
func (e *E2e) GetHistoricalDHPubkey() ([]byte, error) {
	return e.api.GetE2E().GetHistoricalDHPubkey().MarshalJSON()
}

// HasAuthenticatedChannel returns true if an authenticated channel with the
// partner exists, otherwise returns false.
//
// Parameters:
//   - partnerId - the marshalled bytes of the id.ID object.
func (e *E2e) HasAuthenticatedChannel(partnerId []byte) (bool, error) {
	partner, err := id.Unmarshal(partnerId)
	if err != nil {
		return false, err
	}
	return e.api.GetE2E().HasAuthenticatedChannel(partner), nil
}

// RemoveService removes all services for the given tag.
func (e *E2e) RemoveService(tag string) error {
	return e.api.GetE2E().RemoveService(tag)
}

// SendE2E send a message containing the payload to the recipient of the passed
// message type, per the given parameters--encrypted with end-to-end encryption.
//
// Parameters:
//   - recipientId - the marshalled bytes of the id.ID object.
//   - e2eParams - the marshalled bytes of the e2e.Params object.
//
// Returns:
//   - []byte - the JSON marshalled bytes of the E2ESendReport object, which can
//     be passed into Cmix.WaitForRoundResult to see if the send succeeded.
func (e *E2e) SendE2E(messageType int, recipientId, payload,
	e2eParamsJSON []byte) ([]byte, error) {
	if len(e2eParamsJSON) == 0 {
		jww.WARN.Printf("e2e params not specified, using defaults...")
		e2eParamsJSON = GetDefaultE2EParams()
	}
	params, err := parseE2EParams(e2eParamsJSON)
	if err != nil {
		return nil, err
	}

	recipient, err := id.Unmarshal(recipientId)
	if err != nil {
		return nil, err
	}

	sendReport, err := e.api.GetE2E().SendE2E(
		catalog.MessageType(messageType), recipient, payload,
		params.Base)
	if err != nil {
		return nil, err
	}

	result := E2ESendReport{
		RoundsList: makeRoundsList(sendReport.RoundList...),
		RoundURL:   getRoundURL(sendReport.RoundList[0]),
		MessageID:  sendReport.MessageId.Marshal(),
		Timestamp:  sendReport.SentTime.UnixNano(),
		KeyResidue: sendReport.KeyResidue.Marshal(),
	}
	return json.Marshal(result)
}

// AddService adds a service for all partners of the given tag, which will call
// back on the given processor. These can be sent to using the tag fields in the
// Params object.
//
// Passing nil for the processor allows you to create a service that is never
// called but will be visible by notifications. Processes added this way are
// generally not end-to-end encrypted messages themselves, but other protocols
// that piggyback on e2e relationships to start communication.
func (e *E2e) AddService(tag string, processor Processor) error {
	return e.api.GetE2E().AddService(
		tag, &messageProcessor{bindingsCbs: processor})
}

// RegisterListener registers a new listener.
//
// Parameters:
//   - senderId - the user ID who sends messages to this user that
//     this function will register a listener for.
//   - messageType - message type from the sender you want to listen for.
//   - newListener: A provider for a callback to hear a message.
//     Do not pass nil to this.
func (e *E2e) RegisterListener(
	senderID []byte, messageType int, newListener Listener) error {
	jww.INFO.Printf("RegisterListener(%v, %d)", senderID, messageType)

	// Convert senderID to id.Id object
	var uid *id.ID
	if len(senderID) == 0 {
		uid = &id.ID{}
	} else {
		var err error
		uid, err = id.Unmarshal(senderID)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to "+
				"ResgisterListener: %+v", err))
		}
	}

	// Register listener
	// todo: when implementing an unregister function, return and provide a way
	//  track this listener ID
	_ = e.api.GetE2E().RegisterListener(uid,
		catalog.MessageType(messageType), listener{l: newListener})

	return nil
}

// Processor is the bindings-specific interface for message.Processor methods.
type Processor interface {
	Process(
		message, tags, metadata, receptionId []byte, ephemeralId, roundId int64)
	fmt.Stringer
}

// messageProcessor implements Processor as a way of obtaining a
// message.Processor over the bindings.
type messageProcessor struct {
	bindingsCbs Processor
}

// convertProcessor turns the input of a message.Processor to the
// binding-layer primitives equivalents within the Processor.Process.
func convertProcessor(msg format.Message, tags []string, metadata []byte,
	receptionID receptionID.EphemeralIdentity, round rounds.Round) (
	message []byte, tagsOut []byte, metadataOut []byte, receptionId []byte, ephemeralId int64, roundId int64) {

	tagsOut, _ = json.Marshal(tags)
	message = msg.Marshal()
	receptionId = receptionID.Source.Marshal()
	ephemeralId = int64(receptionID.EphId.UInt64())
	roundId = int64(round.ID)
	metadataOut = metadata[:]
	return
}

// Process decrypts and hands off the message to its internal down stream
// message processing system.
//
// CRITICAL: Fingerprints should never be used twice. Process must denote, in
// long-term storage, usage of a fingerprint and that fingerprint must not be
// added again during application load. It is a security vulnerability to reuse
// a fingerprint. It leaks privacy and can lead to compromise of message
// contents and integrity.
func (m *messageProcessor) Process(msg format.Message, tags []string, metadata []byte,
	receptionID receptionID.EphemeralIdentity, roundId rounds.Round) {
	m.bindingsCbs.Process(convertProcessor(msg, tags, metadata, receptionID, roundId))
}

// String prints a name for debugging.
func (m *messageProcessor) String() string {
	return m.bindingsCbs.String()
}
