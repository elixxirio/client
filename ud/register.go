package ud

import (
	"github.com/pkg/errors"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/crypto/factID"
	"gitlab.com/elixxir/crypto/hash"
	"gitlab.com/elixxir/primitives/fact"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/crypto/signature/rsa"
)

type registerUserComms interface {
	SendRegisterUser(*connect.Host, *pb.UDBUserRegistration) (*messages.Ack, error)
}

// Register registers a user with user discovery.
func (m *Manager) Register(username string) error {
	return m.register(username, m.comms)
}

// register registers a user with user discovery with a specified comm for
// easier testing.
func (m *Manager) register(username string, comm registerUserComms) error {
	var err error
	user := m.storage.User()
	cryptoUser := m.storage.User().GetCryptographicIdentity()
	rng := m.rng.GetStream()

	// Construct the user registration message
	msg := &pb.UDBUserRegistration{
		PermissioningSignature: user.GetRegistrationValidationSignature(),
		RSAPublicPem:           string(rsa.CreatePublicKeyPem(cryptoUser.GetRSA().GetPublic())),
		IdentityRegistration: &pb.Identity{
			Username: username,
			DhPubKey: m.storage.E2e().GetDHPublicKey().Bytes(),
			Salt:     cryptoUser.GetSalt(),
		},
		UID: cryptoUser.GetUserID().Marshal(),
	}

	// Sign the identity data and add to user registration message
	identityDigest := msg.IdentityRegistration.Digest()
	msg.IdentitySignature, err = rsa.Sign(rng, cryptoUser.GetRSA(),
		hash.CMixHash, identityDigest, nil)
	if err != nil {
		return errors.Errorf("Failed to sign user's IdentityRegistration: %+v", err)
	}

	// Create new username fact
	usernameFact, err := fact.NewFact(fact.Username, username)
	if err != nil {
		return errors.Errorf("Failed to create new username fact: %+v", err)
	}

	// Hash and sign fact
	hashedFact := factID.Fingerprint(usernameFact)
	signedFact, err := rsa.Sign(rng, cryptoUser.GetRSA(), hash.CMixHash, hashedFact, nil)

	// Add username fact register request to the user registration message
	msg.Frs = &pb.FactRegisterRequest{
		UID: cryptoUser.GetUserID().Marshal(),
		Fact: &pb.Fact{
			Fact:     username,
			FactType: 0,
		},
		FactSig: signedFact,
	}

	// Register user with user discovery
	_, err = comm.SendRegisterUser(m.host, msg)

	return err
}
