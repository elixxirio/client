////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package cypher

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/client/v4/storage/utility"
	ftCrypto "gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/xx_network/primitives/netTime"
)

// Storage keys and versions.
const (
	cypherManagerPrefix          = "CypherManagerStore"
	cypherManagerFpVectorKey     = "CypherManagerFingerprintVector"
	cypherManagerKeyStoreKey     = "CypherManagerKey"
	cypherManagerKeyStoreVersion = 0
)

// Error messages.
const (
	// NewManager
	errNewFpVector = "failed to create new state vector for fingerprints: %+v"
	errSaveKey     = "failed to save transfer key: %+v"

	// LoadManager
	errLoadKey      = "failed to load transfer key: %+v"
	errLoadFpVector = "failed to load state vector: %+v"

	// Manager.PopCypher
	errGetNextFp = "used all %d fingerprints"

	// Manager.Delete
	errDeleteKey      = "failed to delete transfer key: %+v"
	errDeleteFpVector = "failed to delete fingerprint state vector: %+v"
)

// Manager the creation
type Manager struct {
	// The transfer key is a randomly generated key created by the sender and
	// used to generate MACs and fingerprints
	key *ftCrypto.TransferKey

	// Stores the state of a fingerprint (used/unused) in a bitstream format
	// (has its own storage backend)
	fpVector *utility.StateVector

	kv versioned.KV
}

// NewManager returns a new cypher Manager initialised with the given number of
// fingerprints.
func NewManager(key *ftCrypto.TransferKey, numFps uint16, kv versioned.KV) (
	*Manager, error) {

	kv, err := kv.Prefix(cypherManagerPrefix)
	if err != nil {
		return nil, err
	}

	fpVector, err := utility.NewStateVector(
		uint32(numFps), false, cypherManagerFpVectorKey, kv)
	if err != nil {
		return nil, errors.Errorf(errNewFpVector, err)
	}

	err = saveKey(key, kv)
	if err != nil {
		return nil, errors.Errorf(errSaveKey, err)
	}

	tfp := &Manager{
		key:      key,
		fpVector: fpVector,
		kv:       kv,
	}

	return tfp, nil
}

// PopCypher returns a new Cypher with next available fingerprint number. This
// marks the fingerprint as used. Returns false if no more fingerprints are
// available.
func (m *Manager) PopCypher() (Cypher, error) {
	fpNum, err := m.fpVector.Next()
	if err != nil {
		return Cypher{}, errors.Errorf(errGetNextFp, m.fpVector.GetNumKeys())
	}

	c := Cypher{
		Manager: m,
		fpNum:   uint16(fpNum),
	}

	return c, nil
}

// GetUnusedCyphers returns a list of cyphers with unused fingerprints numbers.
func (m *Manager) GetUnusedCyphers() []Cypher {
	fpNums := m.fpVector.GetUnusedKeyNums()
	cypherList := make([]Cypher, len(fpNums))

	for i, fpNum := range fpNums {
		cypherList[i] = Cypher{
			Manager: m,
			fpNum:   uint16(fpNum),
		}
	}

	return cypherList
}

////////////////////////////////////////////////////////////////////////////////
// Storage Functions                                                          //
////////////////////////////////////////////////////////////////////////////////

// LoadManager loads the Manager from storage.
func LoadManager(kv versioned.KV) (*Manager, error) {
	kv, err := kv.Prefix(cypherManagerPrefix)
	if err != nil {
		return nil, err
	}
	key, err := loadKey(kv)
	if err != nil {
		return nil, errors.Errorf(errLoadKey, err)
	}

	fpVector, err := utility.LoadStateVector(kv, cypherManagerFpVectorKey)
	if err != nil {
		return nil, errors.Errorf(errLoadFpVector, err)
	}

	tfp := &Manager{
		key:      key,
		fpVector: fpVector,
		kv:       kv,
	}

	return tfp, nil
}

// Delete removes all saved entries from storage.
func (m *Manager) Delete() error {
	// Delete transfer key
	err := m.kv.Delete(cypherManagerKeyStoreKey, cypherManagerKeyStoreVersion)
	if err != nil {
		return errors.Errorf(errDeleteKey, err)
	}

	// Delete StateVector
	err = m.fpVector.Delete()
	if err != nil {
		return errors.Errorf(errDeleteFpVector, err)
	}

	return nil
}

// saveKey saves the transfer key to storage.
func saveKey(key *ftCrypto.TransferKey, kv versioned.KV) error {
	obj := &versioned.Object{
		Version:   cypherManagerKeyStoreVersion,
		Timestamp: netTime.Now(),
		Data:      key.Bytes(),
	}

	return kv.Set(cypherManagerKeyStoreKey, obj)
}

// loadKey loads the transfer key from storage.
func loadKey(kv versioned.KV) (*ftCrypto.TransferKey, error) {
	obj, err := kv.Get(cypherManagerKeyStoreKey, cypherManagerKeyStoreVersion)
	if err != nil {
		return nil, err
	}

	key := ftCrypto.UnmarshalTransferKey(obj.Data)
	return &key, nil
}
