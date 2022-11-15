///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package ud

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/crypto/partnerships/crust"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/crypto/signature/rsa"
)

// userValidationComms is a sub-interface of the Comms interface for
// username validation.
type userValidationComms interface {
	SendUsernameValidation(host *connect.Host,
		message *pb.UsernameValidationRequest) (*pb.UsernameValidation, error)
}

// GetUsername returns the username from the Manager's store.
func (m *Manager) GetUsername() (string, error) {
	return m.storage.GetUd().GetUsername()
}

// GetUsernameValidationSignature will lazily load a username validation
// signature. If it is not already present within the Manager object, it
// will query the signature from the UD service.
func (m *Manager) GetUsernameValidationSignature() ([]byte, error) {
	m.usernameValidationMux.Lock()
	defer m.usernameValidationMux.Unlock()
	var err error

	// If validation signature is not present, request it from
	// UD
	if m.usernameValidationSignature == nil {
		m.usernameValidationSignature, err = m.getUsernameValidationSignature(m.comms)
		if err != nil {
			return nil, errors.Errorf("Failed to retrieve signature from "+
				"UD: %v", err)
		}
	}

	return m.usernameValidationSignature, nil
}

// getUsernameValidationSignature is the helper function which queries
// the signature from the UD service.
func (m *Manager) getUsernameValidationSignature(
	comms userValidationComms) ([]byte, error) {

	// Construct request for username validation
	request := &pb.UsernameValidationRequest{
		UserId: m.myID.Marshal(),
	}

	// Get UD host
	host, err := m.getHost()
	if err != nil {
		return nil, err
	}

	jww.INFO.Printf("[CRUST] Retrieving username validation from UD...")

	// Send request
	response, err := comms.SendUsernameValidation(host, request)
	if err != nil {
		jww.INFO.Printf("[CRUST] Received error from UsernameValidation: %+v", err)
		return nil, err
	}

	jww.INFO.Printf("[CRUST] Retrieved username validation from UD...")

	publicKey, err := rsa.LoadPublicKeyFromPem(response.ReceptionPublicKeyPem)
	if err != nil {
		return nil, err
	}

	// Verify response is valid
	err = crust.VerifyVerificationSignature(host.GetPubKey(),
		crust.HashUsername(response.Username),
		publicKey, response.Signature)
	if err != nil {
		return nil, err
	}

	// Return response
	return response.Signature, nil
}
