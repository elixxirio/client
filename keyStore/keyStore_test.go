package keyStore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"gitlab.com/elixxir/primitives/id"
	"reflect"
	"testing"
)

// Test GOB Encode/Decode of KeyStore
// and compare if all keys match originals
func TestKeyStore_Gob(t *testing.T) {
	grp := initGroup()
	baseKey := grp.NewInt(57)
	privKey := grp.NewInt(5)
	pubKey := grp.NewInt(42)
	partner := id.NewUserFromUint(14, t)
	userID := id.NewUserFromUint(18, t)

	ks := NewStore()
	km := NewManager(baseKey, privKey, pubKey,
		partner, true, 12, 10, 10)

	// Generate Send Keys
	km.GenerateKeys(grp, userID, ks)

	km2 := NewManager(baseKey, privKey, pubKey,
		partner, false, 12, 10, 10)

	// Generate Receive Keys
	km2.GenerateKeys(grp, userID, ks)

	// Now that some KeyManagers are in the keystore, Gob Encode it
	var byteBuf bytes.Buffer

	enc := gob.NewEncoder(&byteBuf)
	dec := gob.NewDecoder(&byteBuf)

	err := enc.Encode(ks)

	if err != nil {
		t.Errorf("Error GOB Encoding KeyStore: %s", err)
	}

	outKs := &KeyStore{}

	err = dec.Decode(&outKs)

	if err != nil {
		t.Errorf("Error GOB Decoding KeyStore: %s", err)
	}

	// Need to reconstruct keys after decoding
	outKs.ReconstructKeys(grp, userID)

	// Get KeyManagers and compare keys
	outKm := outKs.GetSendManager(partner)

	for i := 0; i < 12; i++ {
		origKey, _ := km.PopKey()
		actualKey, _ := outKm.PopKey()

		if origKey.GetOuterType() != actualKey.GetOuterType() {
			t.Errorf("Send Key type mistmatch after GOB Encode/Decode")
		} else if origKey.key.Cmp(actualKey.key) != 0 {
			t.Errorf("Send Key mistmatch after GOB Encode/Decode")
		}
	}

	for i := 0; i < 10; i++ {
		origKey, _ := km.PopRekey()
		actualKey, _ := outKm.PopRekey()

		if origKey.GetOuterType() != actualKey.GetOuterType() {
			t.Errorf("Send Key type mistmatch after GOB Encode/Decode")
		} else if origKey.key.Cmp(actualKey.key) != 0 {
			t.Errorf("Send Key mistmatch after GOB Encode/Decode")
		}
	}
}

// Tests that GobDecode() for Key Store throws an error for a
// malformed byte array
func TestKeyStore_GobDecodeErrors(t *testing.T) {
	ksTest := KeyStore{}
	err := ksTest.GobDecode([]byte{})

	if !reflect.DeepEqual(err, errors.New("EOF")) {
		t.Errorf("GobDecode() did not produce the expected error\n\treceived: %v"+
			"\n\texpected: %v", err, errors.New("EOF"))
	}
}
