////////////////////////////////////////////////////////////////////////////////
// Copyright © 2019 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package user

import (
	"crypto/sha256"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/large"
	"gitlab.com/elixxir/primitives/id"
	"testing"
)

// TestUserRegistry tests the constructors/getters/setters
// surrounding the User struct and the Registry interface
func TestUserRegistry(t *testing.T) {
	InitUserRegistry(InitGroup())
	// Test if CountUsers correctly counts the hard-coded demo users
	if Users.CountUsers() != NumDemoUsers {
		t.Errorf("CountUsers: Start size of userRegistry not zero!")
	}
	// Test the integration of the LookupUser, UserHash and GetUser functions
	for i := 0; i < len(DemoUserNicks); i++ {
		currentID := id.NewIdFromUInt(uint64(i+1), id.User, t)
		reg, ok := Users.LookupUser(RegistrationCode(currentID))
		if !ok {
			t.Errorf("Couldn't lookup user %q with code %v", *currentID,
				RegistrationCode(currentID))
		}
		usr, ok := Users.GetUser(reg)
		if !ok {
			t.Logf("Reg codes of both: %v, %v", RegistrationCode(reg),
				RegistrationCode(currentID))
			t.Errorf("Couldn't get user %q corresponding to user %q",
				*reg, *currentID)
		}
		if usr.Username != DemoUserNicks[i] {
			t.Errorf("Nickname incorrectly set. Expected: %v Actual: %v",
				DemoUserNicks[i], usr.Username)
		}
	}
	// Test the NewUser function
	newID := id.NewIdFromUInt(2002, id.User, t)
	usr := Users.NewUser(newID, "Will I am")

	if usr.Username != "Will I am" {
		t.Errorf("User name should be 'Will I am', but is %v instead", usr.Username)
	}

	// Test that UpsertUser successfully adds a user to the usermap
	userCount := Users.CountUsers()
	Users.UpsertUser(usr)
	if Users.CountUsers() != userCount+1 {
		t.Errorf("Upsert did not add a new user. User count is incorrect")
	}
	newUsr, suc := Users.GetUser(newID)
	if !suc {
		t.Errorf("Upsert did not add the test user correctly. " +
			"The ID was not found by GetUser.")
	}
	if newUsr.Username != "Will I am" {
		t.Errorf("Upsert did not add the test user correctly. "+
			"The set nickname was incorrect. Expected: Will I am, "+
			"Actual: %v", newUsr.Username)
	}

	// Initialize group
	grp := InitGroup()

	// Test LookupKeys
	keys, suc := Users.LookupKeys(id.NewIdFromUInt(1, id.User, t))
	if !suc {
		t.Errorf("LookupKeys failed to find a valid user.")
	}
	h := sha256.New()
	h.Write([]byte(string(40001)))
	key := grp.NewIntFromBytes(h.Sum(nil))
	if keys.TransmissionKey.Text(16) != key.Text(16) {
		t.Errorf("LookupKeys returned an incorrect key. "+
			"Expected:%v \nActual%v", key.Text(16),
			keys.TransmissionKey.Text(16))
	}

	// Test delete user
	Users.DeleteUser(id.NewIdFromUInt(2, id.User, t))

	_, ok := Users.GetUser(id.NewIdFromUInt(2, id.User, t))
	if ok {
		t.Errorf("User %v has not been deleted succesfully!", usr.Username)
	}
}

// Doesn't actually do any testing, but can print the registration codes for
// the first several users
func TestPrintRegCodes(t *testing.T) {
	for i := 1; i <= NumDemoUsers; i++ {
		currentID := id.NewIdFromUInt(uint64(i), id.User, t)
		t.Logf("%v:\t%v", i, RegistrationCode(currentID))
	}
}

// InitGroup sets up the cryptographic constants for cMix
func InitGroup() *cyclic.Group {

	base := 16

	pString := "9DB6FB5951B66BB6FE1E140F1D2CE5502374161FD6538DF1648218642F0B5C48" +
		"C8F7A41AADFA187324B87674FA1822B00F1ECF8136943D7C55757264E5A1A44F" +
		"FE012E9936E00C1D3E9310B01C7D179805D3058B2A9F4BB6F9716BFE6117C6B5" +
		"B3CC4D9BE341104AD4A80AD6C94E005F4B993E14F091EB51743BF33050C38DE2" +
		"35567E1B34C3D6A5C0CEAA1A0F368213C3D19843D0B4B09DCB9FC72D39C8DE41" +
		"F1BF14D4BB4563CA28371621CAD3324B6A2D392145BEBFAC748805236F5CA2FE" +
		"92B871CD8F9C36D3292B5509CA8CAA77A2ADFC7BFD77DDA6F71125A7456FEA15" +
		"3E433256A2261C6A06ED3693797E7995FAD5AABBCFBE3EDA2741E375404AE25B"

	gString := "5C7FF6B06F8F143FE8288433493E4769C4D988ACE5BE25A0E24809670716C613" +
		"D7B0CEE6932F8FAA7C44D2CB24523DA53FBE4F6EC3595892D1AA58C4328A06C4" +
		"6A15662E7EAA703A1DECF8BBB2D05DBE2EB956C142A338661D10461C0D135472" +
		"085057F3494309FFA73C611F78B32ADBB5740C361C9F35BE90997DB2014E2EF5" +
		"AA61782F52ABEB8BD6432C4DD097BC5423B285DAFB60DC364E8161F4A2A35ACA" +
		"3A10B1C4D203CC76A470A33AFDCBDD92959859ABD8B56E1725252D78EAC66E71" +
		"BA9AE3F1DD2487199874393CD4D832186800654760E1E34C09E4D155179F9EC0" +
		"DC4473F996BDCE6EED1CABED8B6F116F7AD9CF505DF0F998E34AB27514B0FFE7"

	p := large.NewIntFromString(pString, base)
	g := large.NewIntFromString(gString, base)

	grp := cyclic.NewGroup(p, g)

	return grp
}
