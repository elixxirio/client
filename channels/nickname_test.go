package channels

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/elixxir/client/v4/collective"
	"gitlab.com/elixxir/client/v4/storage/versioned"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/xx_network/primitives/id"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Unit test. Tests that once you set a nickname with SetNickname, you can
// retrieve the nickname using GetNickname.
func TestNicknameManager_SetGetNickname(t *testing.T) {
	rkv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs)
	nm := loadOrNewNicknameManager(rkv)

	for i := 0; i < numTests; i++ {
		chId := id.NewIdFromUInt(uint64(i), id.User, t)
		nickname := "nickname#" + strconv.Itoa(i)
		err := nm.SetNickname(nickname, chId)
		if err != nil {
			t.Fatalf("SetNickname error when setting %s: %+v", nickname, err)
		}

		received, _ := nm.GetNickname(chId)
		if received != nickname {
			t.Fatalf("GetNickname did not return expected values."+
				"\nExpected: %s"+
				"\nReceived: %s", nickname, received)
		}
	}
}

// Unit test. Tests that once you set a nickname with SetNickname, you can
// retrieve the nickname using GetNickname after a reload.
func TestNicknameManager_SetGetNickname_Reload(t *testing.T) {
	rkv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs)
	nm := loadOrNewNicknameManager(rkv)

	for i := 0; i < numTests; i++ {
		chId := id.NewIdFromUInt(uint64(i), id.User, t)
		nickname := "nickname#" + strconv.Itoa(i)
		err := nm.SetNickname(nickname, chId)
		if err != nil {
			t.Fatalf("SetNickname error when setting %s: %+v", nickname, err)
		}
	}

	nm2 := loadOrNewNicknameManager(rkv)

	fmt.Println(nm2.byChannel)

	for i := 0; i < numTests; i++ {
		chId := id.NewIdFromUInt(uint64(i), id.User, t)
		nick, exists := nm2.GetNickname(chId)
		if !exists {
			t.Fatalf("Nickname %d not found  ", i)
		}
		expected := "nickname#" + strconv.Itoa(i)
		if nick != expected {
			t.Fatalf("Nickname %d not found, expected: %s, received: %s ",
				i, expected, nick)
		}
	}
}

// Error case: Tests that nicknameManager.GetNickname returns a false boolean
// if no nickname has been set with the channel ID.
func TestNicknameManager_GetNickname_Error(t *testing.T) {
	rkv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs)
	nm := loadOrNewNicknameManager(rkv)

	for i := 0; i < numTests; i++ {
		chId := id.NewIdFromUInt(uint64(i), id.User, t)
		_, exists := nm.GetNickname(chId)
		if exists {
			t.Fatalf("GetNickname expected error case: " +
				"This should not retrieve nicknames for channel IDs " +
				"that are not set.")
		}
	}
}

func TestNicknameManager_DeleteNickname(t *testing.T) {
	kv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs)
	nm := loadOrNewNicknameManager(kv)

	for i := 0; i < numTests; i++ {
		chId := id.NewIdFromUInt(uint64(i), id.User, t)
		nickname := "nickname#" + strconv.Itoa(i)
		err := nm.SetNickname(nickname, chId)
		if err != nil {
			t.Fatalf("SetNickname error when setting %s: %+v", nickname, err)
		}

		err = nm.DeleteNickname(chId)
		if err != nil {
			t.Fatalf("DeleteNickname error: %+v", err)
		}

		_, exists := nm.GetNickname(chId)
		if exists {
			t.Fatalf("GetNickname expected error case: " +
				"This should not retrieve nicknames for channel IDs " +
				"that are not set.")
		}
	}
}

func TestNicknameManager_mapUpdate(t *testing.T) {
	kv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs)
	nm := loadOrNewNicknameManager(kv)

	numIDs := 100

	wg := &sync.WaitGroup{}
	wg.Add(numIDs)

	expectedUpdates := make(map[id.ID]NicknameUpdate, numIDs)
	edits := make(map[string]versioned.ElementEdit, numIDs)

	rng := rand.New(rand.NewSource(69))

	for i := 0; i < numIDs; i++ {
		cid := &id.ID{}
		cid[0] = byte(i)

		nicknameBytes := make([]byte, 10)
		rng.Read(nicknameBytes)
		nickname := base64.StdEncoding.EncodeToString(nicknameBytes)

		// make 1/3 chance it will be deleted
		existsChoice := make([]byte, 1)
		rng.Read(existsChoice)
		op := versioned.KeyOperation(int(existsChoice[0]) % 3)
		data, _ := json.Marshal(&nickname)

		nu := NicknameUpdate{
			ChannelId:      cid,
			Nickname:       nickname,
			NicknameExists: true,
		}

		if op == versioned.Deleted { // set the nickname if it is to be deleted so we can test deletion works
			if err := nm.SetNickname(nickname, cid); err != nil {
				t.Fatalf("Failed to set nickname for deletion: %+v", err)
			}
			data = nil
			nu.Nickname = ""
			nu.NicknameExists = false
		} else if op == versioned.Updated {
			rng.Read(nicknameBytes)
			nicknameOld := base64.StdEncoding.EncodeToString(nicknameBytes)
			if err := nm.SetNickname(nicknameOld, cid); err != nil {
				t.Fatalf("Failed to set nickname for updating: %+v", err)
			}
		}

		expectedUpdates[*cid] = nu

		//create the edit that will be processed

		edits[marshalChID(cid)] = versioned.ElementEdit{
			OldElement: nil,
			NewElement: &versioned.Object{
				Version:   0,
				Timestamp: time.Now(),
				Data:      data,
			},
			Operation: op,
		}
	}

	testingCB := func(update NicknameUpdate) {
		expectedNU, exists := expectedUpdates[*update.ChannelId]
		if !exists {
			t.Errorf("Update not found in list of updates: %+v", update)
		} else if !expectedNU.Equals(update) {
			t.Errorf("updates do not match: %+v vs %+v", update, expectedNU)
		}

		wg.Done()
	}

	nm.callback = testingCB

	nm.mapUpdate(nicknameMapName, edits)

	wg.Wait()

}
