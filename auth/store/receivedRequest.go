package store

import (
	"github.com/cloudflare/circl/dh/sidh"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	util "gitlab.com/elixxir/client/storage/utility"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/xx_network/primitives/id"
	"sync"
)

type ReceivedRequest struct {
	kv *versioned.KV

	aid authIdentity

	// ID received on
	myID *id.ID

	// contact of partner
	partner contact.Contact

	//sidHPublic key of partner
	theirSidHPubKeyA *sidh.PublicKey

	//lock to make sure only one operator at a time
	mux *sync.Mutex
}

func newReceivedRequest(kv *versioned.KV, myID *id.ID, c contact.Contact,
	key *sidh.PublicKey) *ReceivedRequest {

	aid := makeAuthIdentity(c.ID, myID)
	kv = kv.Prefix(makeReceiveRequestPrefix(aid))

	if err := util.StoreContact(kv, c); err != nil {
		jww.FATAL.Panicf("Failed to save contact for partner %s", c.ID.String())
	}

	storeKey := util.MakeSIDHPublicKeyKey(c.ID)
	if err := util.StoreSIDHPublicKey(kv, key, storeKey); err != nil {
		jww.FATAL.Panicf("Failed to save contact pubKey for partner %s",
			c.ID.String())
	}

	return &ReceivedRequest{
		kv:               kv,
		aid:              aid,
		myID:             myID,
		partner:          c,
		theirSidHPubKeyA: key,
	}
}

func loadReceivedRequest(kv *versioned.KV, partner *id.ID, myID *id.ID) (
	*ReceivedRequest, error) {

	// try the load with both the new prefix and the old, which one is
	// successful will determine which file structure the sent request will use
	// a change was made when auth was upgraded to handle auths for multiple
	// outgoing IDs and it became possible to have multiple auths for the same
	// partner at a time, so it now needed to be keyed on the touple of
	// partnerID,MyID. Old receivedByID always have the same myID so they can be left
	// at their own paths
	aid := makeAuthIdentity(partner, myID)
	newKV := kv
	oldKV := kv.Prefix(makeReceiveRequestPrefix(aid))

	c, err := util.LoadContact(newKV, partner)

	//loading with the new prefix path failed, try with the new
	if err != nil {
		c, err = util.LoadContact(newKV, partner)
		if err != nil {
			return nil, errors.WithMessagef(err, "Failed to Load "+
				"Received Auth Request Contact with %s and to %s",
				partner, myID)
		} else {
			kv = oldKV
		}
	} else {
		kv = newKV
	}

	key, err := util.LoadSIDHPublicKey(kv,
		util.MakeSIDHPublicKeyKey(c.ID))
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to Load "+
			"Received Auth Request Partner SIDHkey with %s and to %s",
			partner, myID)
	}

	return &ReceivedRequest{
		aid:              aid,
		kv:               kv,
		myID:             myID,
		partner:          c,
		theirSidHPubKeyA: key,
	}, nil
}

func (rr *ReceivedRequest) GetMyID() *id.ID {
	return rr.myID
}

func (rr *ReceivedRequest) GetContact() contact.Contact {
	return rr.partner
}

func (rr *ReceivedRequest) GetTheirSidHPubKeyA() *sidh.PublicKey {
	return rr.theirSidHPubKeyA
}

func (rr *ReceivedRequest) delete() {
	if err := util.DeleteContact(rr.kv, rr.partner.ID); err != nil {
		jww.FATAL.Panicf("Failed to delete received request "+
			"contact for %s to %s", rr.partner.ID, rr.myID)
	}
	if err := util.DeleteSIDHPublicKey(rr.kv,
		util.MakeSIDHPublicKeyKey(rr.partner.ID)); err != nil {
		jww.FATAL.Panicf("Failed to delete received request "+
			"SIDH pubkey for %s to %s", rr.partner.ID, rr.myID)
	}
}

func (rr *ReceivedRequest) getType() RequestType {
	return Receive
}

func (rr *ReceivedRequest) isTemporary() bool {
	return rr.kv.IsMemStore()
}