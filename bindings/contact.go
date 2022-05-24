package bindings

import (
	"encoding/json"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/diffieHellman"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/crypto/xx"
	"gitlab.com/xx_network/primitives/id"
)

type Identity struct {
	ID            []byte
	RSAPrivatePem []byte
	Salt          []byte
	DHKeyPrivate  []byte
}

type Fact struct {
	Fact string
	Type string
}

// MakeIdentity generates a new cryptographic identity for receving
// messages
func (c *Client) MakeIdentity() ([]byte, error) {
	stream := c.api.GetRng().GetStream()
	defer stream.Close()

	//make RSA Key
	rsaKey, err := rsa.GenerateKey(stream,
		rsa.DefaultRSABitLen)
	if err != nil {
		return nil, err
	}

	//make salt
	salt := make([]byte, 32)
	_, err = stream.Read(salt)

	//make dh private key
	privkey := diffieHellman.GeneratePrivateKey(
		len(c.api.GetStorage().GetE2EGroup().GetPBytes()),
		c.api.GetStorage().GetE2EGroup(), stream)

	//make the ID
	id, err := xx.NewID(rsaKey.GetPublic(),
		salt, id.User)

	if err != nil {
		return nil, err
	}

	//create the identity object
	I := Identity{
		ID:            id.Marshal(),
		RSAPrivatePem: rsa.CreatePrivateKeyPem(rsaKey),
		Salt:          salt,
		DHKeyPrivate:  privkey.Bytes(),
	}

	return json.Marshal(&I)
}

func GetContactFromIdentity(identity string) []byte {
	I := Identity{}
}

func unmarshalIdentity(marshaled []byte) (*id.ID, *rsa.PrivateKey, []byte,
	*cyclic.Int, error) {
	return nil, nil, nil, nil, nil
}

// SetFactsOnContact replaces the facts on the contact with the passed in facts
// pass in empty facts in order to clear the facts
func SetFactsOnContact(contact []byte, facts []byte) []byte {
	I := Identity{}
}

func GetIDFromContact(contact []byte) []byte {

}

func GetPubkeyFromContact(contact []byte) []byte {

}

func GetFactsFromContact(contact []byte) []byte {

}
