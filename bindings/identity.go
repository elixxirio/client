////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package bindings

import (
	"encoding/json"
	"gitlab.com/elixxir/client/v4/xxdk"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/elixxir/primitives/fact"
)

////////////////////////////////////////////////////////////////////////////////
// ReceptionIdentity                                                          //
////////////////////////////////////////////////////////////////////////////////

// ReceptionIdentity struct.
//
// JSON example:
//
//	 {
//	  "ID":"emV6aW1hAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD",
//	  "RSAPrivate":"LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBNU15dTdhYjBJOS9UL1BFUUxtd2x3ejZHV3FjMUNYemVIVXhoVEc4bmg1WWRWSXMxCmJ2THpBVjNOMDJxdXN6K2s4TVFEWjBtejMzdkswUmhPczZIY0NUSFdzTEpXRkE5WWpzWWlCRi9qTDd1bmd1ckIKL2tvK1JJSnNrWGFWaEZaazRGdERoRXhTNWY4RnR0Qmk1NmNLZmdJQlVKT3ozZi9qQllTMkxzMlJ6cWV5YXM3SApjV2RaME9TclBTT3BiYlViU1FPbS9LWnlweGZHU21yZ2oxRUZuU1dZZ2xGZTdUOTRPbHF5MG14QTV5clVXbHorCk9sK3hHbXpCNUp4WUFSMU9oMFQrQTk4RWMrTUZHNm43L1MraDdzRDgybGRnVnJmbStFTzRCdmFKeTRESGZGMWgKNnp6QnVnY25NUVFGc0dLeDFYWC9COTVMdUpPVjdyeXlDbzZGbHdJREFRQUJBb0lCQVFDaUh6OGNlcDZvQk9RTAphUzBVRitHeU5VMnlVcVRNTWtTWThoUkh1c09CMmFheXoybHZVb3RLUHBPbjZRSWRWVTJrcE4vY2dtY0lSb2x5CkhBMDRUOHJBWVNaRlVqaVlRajkzKzRFREpJYXd2Z0YyVEs1bFoyb3oxVTdreStncU82V0RMR2Z0Q0wvODVQWEIKa210aXhnUXpRV3g1RWcvemtHdm03eURBalQxeDloNytsRjJwNFlBam5kT2xTS0dmQjFZeTR1RXBQd0kwc1lWdgpKQWc0MEFxbllZUmt4emJPbmQxWGNjdEJFN2Z1VDdrWXhoeSs3WXYrUTJwVy9BYmh6NGlHOEY1MW9GMGZwV0czCmlISDhsVXZFTkp2SUZEVHZ0UEpESlFZalBRN3lUbGlGZUdrMXZUQkcyQkpQNExzVzhpbDZOeUFuRktaY1hOQ24KeHVCendiSlJBb0dCQVBUK0dGTVJGRHRHZVl6NmwzZmg3UjJ0MlhrMysvUmpvR3BDUWREWDhYNERqR1pVd1RGVQpOS2tQTTNjS29ia2RBYlBDb3FpL0tOOVBibk9QVlZ3R3JkSE9vSnNibFVHYmJGamFTUzJQMFZnNUVhTC9rT2dUCmxMMUdoVFpIUWk1VUlMM0p4M1Z3T0ZRQ3RQOU1UQlQ0UEQvcEFLbDg3VTJXN3JTY1dGV1ZGbFNkQW9HQkFPOFUKVmhHWkRpVGFKTWVtSGZIdVYrNmtzaUlsam9aUVVzeGpmTGNMZ2NjV2RmTHBqS0ZWTzJNN3NqcEJEZ0w4NmFnegorVk14ZkQzZ1l0SmNWN01aMVcwNlZ6TlNVTHh3a1dRY1hXUWdDaXc5elpyYlhCUmZRNUVjMFBlblVoWWVwVzF5CkpkTC8rSlpQeDJxSzVrQytiWU5EdmxlNWdpcjlDSGVzTlR5enVyckRBb0dCQUl0cTJnN1RaazhCSVFUUVNrZ24Kb3BkRUtzRW4wZExXcXlBdENtVTlyaWpHL2l2eHlXczMveXZDQWNpWm5VVEp0QUZISHVlbXVTeXplQ2g5QmRkegoyWkRPNUdqQVBxVHlQS3NudFlNZkY4UDczZ1NES1VSWWVFbHFDejdET0c5QzRzcitPK3FoN1B3cCtqUmFoK1ZiCkNuWllNMDlBVDQ3YStJYUJmbWRkaXpLbEFvR0JBSmo1dkRDNmJIQnNISWlhNUNJL1RZaG5YWXUzMkVCYytQM0sKMHF3VThzOCtzZTNpUHBla2Y4RjVHd3RuUU4zc2tsMk1GQWFGYldmeVFZazBpUEVTb0p1cGJzNXA1enNNRkJ1bwpncUZrVnQ0RUZhRDJweTVwM2tQbDJsZjhlZXVwWkZScGE0WmRQdVIrMjZ4eWYrNEJhdlZJeld3NFNPL1V4Q3crCnhqbTNEczRkQW9HQWREL0VOa1BjU004c1BCM3JSWW9MQ2twcUV2U0MzbVZSbjNJd3c1WFAwcDRRVndhRmR1ckMKYUhtSE1EekNrNEUvb0haQVhFdGZ2S2tRaUI4MXVYM2c1aVo4amdYUVhXUHRteTVIcVVhcWJYUTlENkxWc3B0egpKL3R4SWJLMXp5c1o2bk9IY1VoUUwyVVF6SlBBRThZNDdjYzVzTThEN3kwZjJ0QURTQUZNMmN3PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==",
//	  "Salt":"4kk02v0NIcGtlobZ/xkxqWz8uH/ams/gjvQm14QT0dI=",
//	  "DHKeyPrivate":"eyJWYWx1ZSI6NDU2MDgzOTEzMjA0OTIyODA5Njg2MDI3MzQ0MzM3OTA0MzAyODYwMjM2NDk2NDM5NDI4NTcxMTMwNDMzOTQwMzgyMTIyMjY4OTQzNTMyMjIyMzc1MTkzNTEzMjU4MjA4MDA0NTczMDY4MjEwNzg2NDI5NjA1MjA0OTA3MjI2ODI5OTc3NTczMDkxODY0NTY3NDExMDExNjQxNCwiRmluZ2VycHJpbnQiOjE2ODAxNTQxNTExMjMzMDk4MzYzfQ=="
//	  "E2eGrp": "eyJnZW4iOiIyIiwicHJpbWUiOiJlMmVlOTgzZDAzMWRjMWRiNmYxYTdhNjdkZjBlOWE4ZTU1NjFkYjhlOGQ0OTQxMzM5NGMwNDliN2E4YWNjZWRjMjk4NzA4ZjEyMTk1MWQ5Y2Y5MjBlYzVkMTQ2NzI3YWE0YWU1MzViMDkyMmM2ODhiNTViM2RkMmFlZGY2YzAxYzk0NzY0ZGFiOTM3OTM1YWE4M2JlMzZlNjc3NjA3MTNhYjQ0YTYzMzdjMjBlNzg2MTU3NWU3NDVkMzFmOGI5ZTlhZDg0MTIxMThjNjJhM2UyZTI5ZGY0NmIwODY0ZDBjOTUxYzM5NGE1Y2JiZGM2YWRjNzE4ZGQyYTNlMDQxMDIzZGJiNWFiMjNlYmI0NzQyZGU5YzE2ODdiNWIzNGZhNDhjMzUyMTYzMmM0YTUzMGU4ZmZiMWJjNTFkYWRkZjQ1M2IwYjI3MTdjMmJjNjY2OWVkNzZiNGJkZDVjOWZmNTU4ZTg4ZjI2ZTU3ODUzMDJiZWRiY2EyM2VhYzVhY2U5MjA5NmVlOGE2MDY0MmZiNjFlOGYzZDI0OTkwYjhjYjEyZWU0NDhlZWY3OGUxODRjNzI0MmRkMTYxYzc3MzhmMzJiZjI5YTg0MTY5ODk3ODgyNWI0MTExYjRiYzNlMWUxOTg0NTUwOTU5NTgzMzNkNzc2ZDhiMmJlZWVkM2ExYTFhMjIxYTZlMzdlNjY0YTY0YjgzOTgxYzQ2ZmZkZGMxYTQ1ZTNkNTIxMWFhZjhiZmJjMDcyNzY4YzRmNTBkN2Q3ODAzZDJkNGYyNzhkZTgwMTRhNDczMjM2MzFkN2UwNjRkZTgxYzBjNmJmYTQzZWYwZTY5OTg4NjBmMTM5MGI1ZDNmZWFjYWYxNjk2MDE1Y2I3OWMzZjljMmQ5M2Q5NjExMjBjZDBlNWYxMmNiYjY4N2VhYjA0NTI0MWY5Njc4OWMzOGU4OWQ3OTYxMzhlNjMxOWJlNjJlMzVkODdiMTA0OGNhMjhiZTM4OWI1NzVlOTk0ZGNhNzU1NDcxNTg0YTA5ZWM3MjM3NDJkYzM1ODczODQ3YWVmNDlmNjZlNDM4NzMifQ=="
//	}
type ReceptionIdentity struct {
	ID            []byte // User ID (base64)
	RSAPrivatePem []byte // RSA Private key (PEM format)
	Salt          []byte // Salt for identity (base64)
	DHKeyPrivate  []byte // DH Private key
	E2eGrp        []byte
}

// StoreReceptionIdentity stores the given identity in Cmix storage with the
// given key. This is the ideal way to securely store identities, as the caller
// of this function is only required to store the given key separately rather
// than the keying material.
func StoreReceptionIdentity(key string, identity []byte, cmixId int) error {
	cmix, err := cmixTrackerSingleton.get(cmixId)
	if err != nil {
		return err
	}
	receptionIdentity, err := xxdk.UnmarshalReceptionIdentity(identity)
	if err != nil {
		return err
	}
	return xxdk.StoreReceptionIdentity(key, receptionIdentity, cmix.api)
}

// LoadReceptionIdentity loads the given identity in Cmix storage with the given
// key.
func LoadReceptionIdentity(key string, cmixId int) ([]byte, error) {
	cmix, err := cmixTrackerSingleton.get(cmixId)
	if err != nil {
		return nil, err
	}
	storageObj, err := cmix.api.GetStorage().Get(key)
	if err != nil {
		return nil, err
	}

	return storageObj.Data, nil
}

// MakeReceptionIdentity generates a new cryptographic identity for receiving
// messages.
func (c *Cmix) MakeReceptionIdentity() ([]byte, error) {
	ident, err := xxdk.MakeReceptionIdentity(c.api)
	if err != nil {
		return nil, err
	}

	return ident.Marshal()
}

// MakeLegacyReceptionIdentity generates the legacy identity for receiving
// messages. As with all legacy calls, this should primarily be used
// for the xx messenger team.
func (c *Cmix) MakeLegacyReceptionIdentity() ([]byte, error) {
	ident, err := xxdk.MakeLegacyReceptionIdentity(c.api)
	if err != nil {
		return nil, err
	}

	return ident.Marshal()
}

// GetReceptionRegistrationValidationSignature returns the signature provided by
// the xx network.
func (c *Cmix) GetReceptionRegistrationValidationSignature() []byte {
	regSig := c.api.GetStorage().GetReceptionRegistrationValidationSignature()
	return regSig
}

////////////////////////////////////////////////////////////////////////////////
// Contact Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// GetIDFromContact returns the ID in the [contact.Contact] object.
//
// Parameters:
//   - marshaledContact - JSON marshalled bytes of [contact.Contact]
//
// Returns:
//   - []byte - bytes of the [id.ID] object
func GetIDFromContact(marshaledContact []byte) ([]byte, error) {
	cnt, err := contact.Unmarshal(marshaledContact)
	if err != nil {
		return nil, err
	}

	return cnt.ID.Marshal(), nil
}

// GetPubkeyFromContact returns the DH public key in the [contact.Contact]
// object.
//
// Parameters:
//   - marshaledContact - JSON marshalled bytes of [contact.Contact]
//
// Returns:
//   - []byte - JSON marshalled bytes of the [cyclic.Int] object
func GetPubkeyFromContact(marshaledContact []byte) ([]byte, error) {
	cnt, err := contact.Unmarshal(marshaledContact)
	if err != nil {
		return nil, err
	}

	return json.Marshal(cnt.DhPubKey)
}

////////////////////////////////////////////////////////////////////////////////
// Fact Functions                                                             //
////////////////////////////////////////////////////////////////////////////////

// SetFactsOnContact replaces the facts on the contact with the passed in facts
// pass in empty facts in order to clear the facts.
//
// Parameters:
//   - marshaledContact - the JSON marshalled bytes of [contact.Contact]
//   - factListJSON - the JSON marshalled bytes of [fact.FactList]
//
// Returns:
//   - []byte - marshalled bytes of the modified [contact.Contact]
func SetFactsOnContact(marshaledContact []byte, factListJSON []byte) ([]byte, error) {
	cnt, err := contact.Unmarshal(marshaledContact)
	if err != nil {
		return nil, err
	}

	var factsList fact.FactList
	err = json.Unmarshal(factListJSON, &factsList)
	if err != nil {
		return nil, err
	}

	cnt.Facts = factsList

	return cnt.Marshal(), nil
}

// GetFactsFromContact returns the fact list in the [contact.Contact] object.
//
// Parameters:
//   - marshaledContact - the JSON marshalled bytes of [contact.Contact]
//
// Returns:
//   - []byte - the JSON marshalled bytes of [fact.FactList]
func GetFactsFromContact(marshaledContact []byte) ([]byte, error) {
	cnt, err := contact.Unmarshal(marshaledContact)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&cnt.Facts)
}
