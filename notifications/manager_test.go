package notifications

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.com/elixxir/client/v4/collective"
	"gitlab.com/elixxir/client/v4/collective/versioned"
	"gitlab.com/elixxir/client/v4/xxdk"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/elixxir/crypto/rsa"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestManager_RegisterUpdateCallback(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	//build the groups
	numGroups := 2000
	groups := make([]string, numGroups)
	for i := 0; i < numGroups; i++ {
		groups[i] = fmt.Sprintf("Group-%d", i)
	}

	update1Response := make([]bool, numGroups)

	// register them all in parallel, showing the mux works (would segfault if
	//it doesn't on concurrent map access)
	wg := &sync.WaitGroup{}
	wg.Add(numGroups)
	for i := 0; i < numGroups; i++ {
		localI := i
		update := func(group Group, created, edits, deletions []*id.ID, max NotificationState) {
			update1Response[localI] = true
		}
		go func() {
			m.RegisterUpdateCallback(groups[localI], update)
			wg.Done()
		}()
	}

	wg.Wait()

	//check that the correct got in
	if len(mInternal.callbacks) != numGroups {
		t.Fatalf("Not all groups got added, only %d/%d",
			len(mInternal.group), numGroups)
	}

	// check that each func is unique by calling and seeign that the correct
	// bool was set
	for gKey := range mInternal.callbacks {
		update := mInternal.callbacks[gKey]
		update(nil, nil, nil, nil, Push)
	}

	for idx, wasSet := range update1Response {
		if !wasSet {
			t.Errorf("index %d was not set", idx)
		}
	}

	//do a second one, setting every even one showing that it overwrites

	update2Response := make([]bool, numGroups)

	wg3 := &sync.WaitGroup{}
	wg3.Add(numGroups / 2)
	for i := 0; i < numGroups; i = i + 2 {
		localI := i
		update := func(group Group, created, edits, deletions []*id.ID,
			max NotificationState) {
			update2Response[localI] = true
		}
		go func() {
			m.RegisterUpdateCallback(groups[localI], update)
			wg3.Done()
		}()
	}

	wg3.Wait()

	//check that the correct number are still set
	if len(mInternal.callbacks) != numGroups {
		t.Fatalf("Not all groups got added, only %d/%d",
			len(mInternal.group), numGroups)
	}

	// check that each func is unique by calling and seeign that the correct
	// bool was set
	for gKey := range mInternal.callbacks {
		update := mInternal.callbacks[gKey]
		update(nil, nil, nil, nil, Push)
	}

	for idx, wasSet := range update2Response {
		if idx%2 == 0 {
			if !wasSet {
				t.Errorf("index %d was not set", idx)
			}
		} else {
			if wasSet {
				t.Errorf("index %d was not set when it shouldnt be", idx)
			}
		}

	}
}

func TestManager_mapUpdate(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	// build the map and the data to check the results against
	numGroups := 4
	numElementsPerGroup := 5
	numTests := numGroups * numElementsPerGroup

	//set up groups and callback recipients
	wg := &sync.WaitGroup{}
	groups := make([]string, numGroups)
	expectedCallbacks := make(map[string]expectedCallback, numGroups)
	calledBack := make([]bool, numGroups)

	for i := 0; i < numGroups; i++ {
		groupName := fmt.Sprintf("group%d", i)
		groups[i] = groupName
		expectedCallbacks[groupName] = expectedCallback{group: make(Group)}
		// even groups will call back
		localI := i
		if i%2 == 0 {
			wg.Add(1)
			cb := func(group Group, created, edits, deletions []*id.ID,
				max NotificationState) {
				received := expectedCallback{
					group:     group,
					created:   created,
					edits:     edits,
					deletions: deletions,
				}
				//check that maps match
				expected := expectedCallbacks[groupName]
				if !reflect.DeepEqual(received.group, expected.group) {
					t.Errorf("group %s did not match, \n %+v \nvs \n %+v",
						groupName, received, expectedCallbacks[groupName])
				}

				//check that all lists match
				require.ElementsMatch(t, received.created, expected.created)
				require.ElementsMatch(t, received.edits, expected.edits)
				require.ElementsMatch(t, received.deletions, expected.deletions)

				calledBack[localI] = true
				wg.Done()
			}
			mInternal.RegisterUpdateCallback(groupName, cb)
		}
	}

	// build map to call, setting up initial state and expected outcome data
	// structures
	edits := make(map[string]versioned.ElementEdit, numTests)
	expected := make(map[id.ID]*registration, numTests)
	groupList := make(map[id.ID]string, numTests)

	for i := 0; i < numTests; i++ {
		nid := id.NewIdFromUInt(uint64(i), id.User, t)

		b := []byte{byte(i)}

		groupName := groups[getGroup(i, numGroups)]

		expectedReg, obj := makeRegAndObj(t, groupName, b, NotificationState(i%3))

		var newObj *versioned.Object
		var oldObj *versioned.Object
		var oldReg *registration

		op := versioned.KeyOperation(i % 3)

		exCb := expectedCallbacks[groupName]

		switch op {
		case versioned.Created:
			newObj = obj
			exCb.created = append(exCb.created, nid)
			exCb.group[*nid] = expectedReg.State
		case versioned.Updated:
			newObj = obj
			oldReg, oldObj = makeRegAndObj(t, groupName, []byte{0}, NotificationState((i+2)%3))
			exCb.edits = append(exCb.edits, nid)
			exCb.group[*nid] = expectedReg.State
		case versioned.Deleted:
			oldObj = obj
			oldReg = expectedReg
			expectedReg = nil
			if i != 2 {
				exCb.deletions = append(exCb.deletions, nid)
			}
		}

		expectedCallbacks[groupName] = exCb

		// for the first delete do not insert to test a silent deletion works
		if i != 2 && (op == versioned.Updated || op == versioned.Deleted) {
			mInternal.upsertNotificationUnsafeRAM(nid, *oldReg)
		}

		elementName := makeElementName(nid)

		edits[elementName] = versioned.ElementEdit{
			OldElement: oldObj,
			NewElement: newObj,
			Operation:  op,
		}

		expected[*nid] = expectedReg
		groupList[*nid] = groupName
	}

	// run the map update
	mInternal.mapUpdate(edits)

	wg.Wait()

	//check the update happened correctly
	for nid, reg := range expected {
		if reg == nil {
			checkElementDeleted(mInternal, &nid, groupList[nid], t)
		} else {
			checkElement(mInternal, &nid, *reg, t)
		}
	}

	//check that callbacks happened correctly
	for i := 0; i < numGroups; i++ {
		groupName := groups[i]
		if i%2 == 0 {
			if !calledBack[i] {
				t.Errorf("Callback for %s (%d) was not called", groupName, i)
			}
		} else {
			if calledBack[i] {
				t.Errorf("Callback for %s (%d) was called when it "+
					"shouldnt have been", groupName, i)
			}
		}
	}
}

type expectedCallback struct {
	group Group
	created,
	edits,
	deletions []*id.ID
}

func makeRegAndObj(t *testing.T, group string, metadata []byte,
	status NotificationState) (*registration, *versioned.Object) {
	reg := &registration{
		Group: group,
		State: State{
			Metadata: metadata,
			Status:   status,
		},
	}
	regBytes, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("Failed to marshal registration %+v", err)
	}

	obj := &versioned.Object{
		Version:   0,
		Timestamp: time.Now(),
		Data:      regBytes,
	}

	return reg, obj
}

func TestManager_loadNotificationsUnsafe(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	numGroups := 4
	numElementsPerGroup := 5
	numTests := numGroups * numElementsPerGroup

	groups := make([]string, numGroups)
	for i := 0; i < numGroups; i++ {
		groups[i] = fmt.Sprintf("group%d", i)
	}

	mapObj := make(map[string]*versioned.Object, numTests)

	for i := 0; i < numTests; i++ {
		nid := id.NewIdFromUInt(uint64(i), id.User, t)

		b := []byte{byte(i)}

		_, obj := makeRegAndObj(t, groups[getGroup(i, numGroups)], b, NotificationState(i%3))

		elementName := makeElementName(nid)
		mapObj[elementName] = obj
	}

	mInternal.loadNotificationsUnsafe(mapObj)

	//check the groups are correct
	for i := 0; i < numGroups; i++ {
		groupName := groups[i]
		g, exists := mInternal.group[groupName]
		if !exists {
			t.Fatalf("group %s does not exist", groupName)
		} else if len(g) != numElementsPerGroup {
			t.Fatalf("Group has the wrong number of elements: %s",
				groupName)
		}
	}

	// check that the elements are correct
	for i := 0; i < numTests; i++ {
		nid := id.NewIdFromUInt(uint64(i), id.User, t)

		b := []byte{byte(i)}

		regExpected, _ := makeRegAndObj(t, groups[getGroup(i, numGroups)],
			b, NotificationState(i%3))

		checkElement(mInternal, nid, *regExpected, t)
	}
}

func checkElement(m *manager, nid *id.ID, regExpected registration, t *testing.T) {

	// check the object is listed
	regReceived, exists := m.notifications[*nid]
	if !exists {
		t.Errorf("registration does not exist for %s", nid)
	}

	if !reflect.DeepEqual(regExpected, regReceived) {
		t.Errorf("Received registrations oes not match expected")
	}

	//check that its group exists
	g, exists := m.group[regExpected.Group]
	if !exists {
		t.Errorf("Group doesnt exist for existing registration")
		return
	}

	stateReceived, exists := g[*nid]
	if !exists {
		t.Errorf("registration does not exist in group %s for %s",
			regExpected.Group, nid)
	}
	if !reflect.DeepEqual(regExpected.State, stateReceived) {
		t.Errorf("Received states does not match expected")
	}
}

func checkElementDeleted(m *manager, nid *id.ID, group string, t *testing.T) {
	// check the object is listed
	_, exists := m.notifications[*nid]
	if exists {
		t.Errorf("registration exist for %s, should be deleted", nid)
	}

	//check that its group exists
	g, exists := m.group[group]
	if !exists {
		t.Errorf("Group doesnt exist for deleted registration, this is valid")
		return
	}

	// if the group does exist, check the element isnt in it
	_, exists = g[*nid]
	if exists {
		t.Errorf("registration does exists in group %s for %s when it "+
			"should be deleted",
			group, nid)
	}
}

func getGroup(i, numGroups int) int {
	return i % numGroups
}

func TestManager_maxStateUpdate(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)
	resultCh := make(chan bool, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- true
			}
		}()
		mInternal.maxStateUpdate(nil, nil, versioned.Deleted)
		resultCh <- false
	}()

	result := <-resultCh
	if result == false {
		t.Errorf("Failed to get panic on delete")
	}

	numGroups := 4
	numCB := numGroups / 2
	wg := &sync.WaitGroup{}

	didRun := make([]bool, numGroups)
	expectedRun := []bool{true, false, true, false}

	var setMax NotificationState

	skip := true
	for i := 0; i < numGroups; i++ {
		groupName := fmt.Sprintf("grp_%d", i)
		nid := id.NewIdFromUInt(uint64(i), id.User, t)
		if i%2 == 0 {

			localI := i
			cb := func(group Group, created, edits, deletions []*id.ID,
				maxState NotificationState) {
				if skip {
					return
				}
				if created != nil || edits != nil || deletions != nil {
					t.Errorf("actions are not nil")
				}
				if maxState != setMax {
					t.Errorf("max state incorrect, expected %s, "+
						"received %s", setMax, maxState)
				}
				didRun[localI] = true
				wg.Done()
			}
			m.RegisterUpdateCallback(groupName, cb)
		}
		m.Set(nid, groupName, []byte{0}, NotificationState(i%3))
	}
	skip = false

	for i := Mute; i <= Push; i++ {
		setMax = i
		for j := versioned.Created; j <= versioned.Loaded; j++ {
			ch := make(chan bool)
			didRun = make([]bool, numGroups)
			if j != versioned.Deleted {
				wg.Add(numCB)
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						ch <- false
					}
				}()
				mInternal.maxStateUpdate(nil, makeMaxStateObj(i, t), j)
				ch <- true
			}()
			result := <-ch
			if j == versioned.Deleted && result {
				t.Errorf("did not panic on deletion")
			} else if j != versioned.Deleted && !result {
				t.Errorf("panicked on non deletion")
			}
			if j != versioned.Deleted {
				wg.Wait()
				if mInternal.maxState != i {
					t.Errorf("max state not set to %s, is %s", i, mInternal.maxState)
				}
				if !reflect.DeepEqual(expectedRun, didRun) {
					t.Errorf("wrong callbacks were called, %+v", didRun)
				}
			}

		}
	}

}

func makeMaxStateObj(max NotificationState, t *testing.T) *versioned.Object {
	b, err := json.Marshal(&max)
	if err != nil {
		t.Fatalf("Failed to marshal max state: %+v", err)
	}
	return &versioned.Object{
		Version:   maxStateKetVersion,
		Timestamp: time.Now(),
		Data:      b,
	}
}

func TestManager_setLoadMaxState(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	mInternal.maxState = -1

	tests := []NotificationState{Mute, WhenOpen, Push}

	for _, ns := range tests {
		mInternal.maxState = -1
		mInternal.setMaxStateUnsafe(ns)
		if mInternal.maxState != ns {
			t.Errorf("In ram max state did not take with set")
		}
		mInternal.maxState = -1
		obj, err := mInternal.remote.Get(maxStateKey, maxStateKetVersion)
		if err != nil {
			t.Errorf("Failed to get max state %s from ekv: %+v", ns, err)
		}
		mInternal.loadMaxStateUnsafe(obj)
		if mInternal.maxState != ns {
			t.Errorf("In ram max state did not take with load of %s",
				string(obj.Data))
		}
	}
}

func TestManager_upsertNotificationUnsafeRAM(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	numTests := 100
	numGroups := numTests / 5
	groups := make([]string, numGroups)
	for i := 0; i < numGroups; i++ {
		groups[i] = fmt.Sprintf("Group-%d", i)
	}

	regs := make([]registration, numTests)
	nids := make([]*id.ID, numTests)
	for i := 0; i < numTests; i++ {
		nid := id.NewIdFromUInt(uint64(i), id.User, t)
		nids[i] = nid
		reg := registration{
			Group: groups[i%numGroups],
			State: State{
				Metadata: nil,
				Status:   1,
			},
		}
		regs[i] = reg
		mInternal.upsertNotificationUnsafeRAM(nid, reg)
	}

	// test that every notification was added
	if len(mInternal.notifications) != numTests {
		t.Errorf("Wrong under of notifications inserted: %d vs %d",
			len(mInternal.notifications), numTests)
	}

	for _, nid := range nids {
		if _, exists := mInternal.notifications[*nid]; !exists {
			t.Errorf("Registration %s not present when it should be", nid)
		}
	}

	//test that groups exist
	if len(mInternal.group) != numGroups {
		t.Errorf("Groups are missing")
	}

	//test that every element is in the right group
	for nid, reg := range mInternal.notifications {
		g := mInternal.group[reg.Group]
		if _, exists := g[nid]; !exists {
			t.Errorf("Group %s missing nid %s", reg.Group, nid)
		}
	}
}

func TestManager_deleteNotificationUnsafeRAM(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	numTests := 100
	numGroups := numTests / 5
	groups := make([]string, numGroups)
	for i := 0; i < numGroups; i++ {
		groups[i] = fmt.Sprintf("Group-%d", i)
	}

	regs := make([]registration, numTests)
	nids := make([]*id.ID, numTests)
	for i := 0; i < numTests; i++ {
		nid := id.NewIdFromUInt(uint64(i), id.User, t)
		nids[i] = nid
		reg := registration{
			Group: groups[i%numGroups],
			State: State{
				Metadata: nil,
				Status:   1,
			},
		}
		regs[i] = reg
		mInternal.upsertNotificationUnsafeRAM(nid, reg)
	}

	// test that every notification was added
	if len(mInternal.notifications) != numTests {
		t.Errorf("Wrong under of notifications inserted: %d vs %d",
			len(mInternal.notifications), numTests)
	}

	for _, nid := range nids {
		if _, exists := mInternal.notifications[*nid]; !exists {
			t.Errorf("Registration %s not present when it should be", nid)
		}
	}

	// delete the odd ones and ones that are dividable by 4, to fully remove from
	// some groups and partially from other
	for i := 0; i < numTests; i++ {
		if (i%2) == 1 || i < numTests/2 {
			grp := mInternal.deleteNotificationUnsafeRAM(nids[i])
			if _, exists := mInternal.notifications[*nids[i]]; exists {
				t.Errorf("Failed to delete %s", nids[i])
			}
			if grp != regs[i].Group {
				t.Errorf("group does not match, %s vs %s", grp, regs[i].Group)
			}
		}

	}

	//test that groups are correct
	for i := 0; i < len(groups); i++ {
		g := groups[i]
		_, exists := mInternal.group[g]
		if (i % 2) == 1 {
			if exists {
				t.Errorf("Group %d should not exist, it is odd", i)
			}
		} else {
			if !exists {
				t.Errorf("Group %d should exist, it is even", i)
			}
		}
	}
}

func TestManager_token(t *testing.T) {
	m, _, _ := buildTestingManager(t)
	mInternal := m.(*manager)

	tokenExpected1 := "blarg"
	appExpected1 := "pswii60"

	tokenExpected2 := "flexnard"
	appExpected2 := "hackentosh"

	//call load when there is no token, token should say empty
	mInternal.loadTokenUnsafe()

	if mInternal.token.Token != "" || mInternal.token.App != "" {
		t.Errorf("Token loaded when not present, token: '%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	// set the token
	setBefore := mInternal.setTokenUnsafe(tokenExpected1, appExpected1)

	if setBefore == true {
		t.Errorf("token was set before when it shouldnt have been")
	}

	// check that it is correct on the object
	if mInternal.token.Token != tokenExpected1 || mInternal.token.App != appExpected1 {
		t.Errorf("Token stored as wrong values, token: '%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	// check that load works, clear the tokens then load them
	mInternal.token.Token = ""
	mInternal.token.App = ""
	mInternal.loadTokenUnsafe()

	if mInternal.token.Token != tokenExpected1 || mInternal.token.App != appExpected1 {
		t.Errorf("Token loaded incorrectly, token: '%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	//set the token to new things
	setBefore = mInternal.setTokenUnsafe(tokenExpected2, appExpected2)

	if setBefore == false {
		t.Errorf("token was not set before when it should have been")
	}

	// check that it is correct on the object
	if mInternal.token.Token != tokenExpected2 || mInternal.token.App != appExpected2 {
		t.Errorf("Token stored as wrong values, token: '%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	// check that load works, clear the tokens then load them
	mInternal.token.Token = ""
	mInternal.token.App = ""
	mInternal.loadTokenUnsafe()

	if mInternal.token.Token != tokenExpected2 ||
		mInternal.token.App != appExpected2 {
		t.Errorf("Token loaded incorrectly, token: '%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	// delete the tokens and check it took in ram
	mInternal.deleteTokenUnsafe()

	if mInternal.token.Token != "" || mInternal.token.App != "" {
		t.Errorf("Token deleted incorrectly from ram, token: "+
			"'%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}

	// load the token and see if it changes to check if the token
	mInternal.loadTokenUnsafe()

	if mInternal.token.Token != "" || mInternal.token.App != "" {
		t.Errorf("Token deleted incorrectly from disk, token: "+
			"'%s', app: '%s",
			mInternal.token.Token, mInternal.token.App)
	}
}

func buildTestingManager(t *testing.T) (Manager, xxdk.TransmissionIdentity, *commsMock) {
	kv := collective.TestingKV(t, ekv.MakeMemstore(), collective.StandardPrefexs, nil)
	comms := initCommsMock()
	rng := fastRNG.NewStreamGenerator(1, 1,
		csprng.NewSystemRNG)

	nid := id.NewIdFromUInt(5, id.User, t)
	scheme := rsa.GetScheme()

	stream := rng.GetStream()
	defer stream.Close()
	priv, err := scheme.Generate(stream, 1024)
	if err != nil {
		t.Fatalf("failed to generate keys: %+v", err)
	}
	ti := xxdk.TransmissionIdentity{
		ID:                    nid,
		RSAPrivate:            priv,
		Salt:                  make([]byte, 32),
		RegistrationTimestamp: time.Now().UTC().UnixNano(),
	}

	_, err = stream.Read(ti.Salt)
	if err != nil {
		t.Fatalf("failed to generate salt: %+v", err)
	}

	regSig := make([]byte, 32)
	_, err = stream.Read(regSig)
	if err != nil {
		t.Fatalf("failed to generate salt: %+v", err)
	}

	m := NewOrLoadManager(ti, regSig, kv, comms, rng)

	return m, ti, comms
}

// object used to track what comms return
type commsMock struct {
	receivedHost    *connect.Host
	receivedMessage interface{}
	returnedMessage *messages.Ack
	returnedError   error
	numRuns         int
}

func initCommsMock() *commsMock {
	cm := &commsMock{}
	cm.reset()
	return cm
}

func (cm *commsMock) reset() {
	cm.receivedHost = nil
	cm.receivedMessage = nil
	cm.returnedMessage = &messages.Ack{}
	cm.returnedError = nil
}

func (cm *commsMock) RegisterToken(host *connect.Host,
	message *pb.RegisterTokenRequest) (*messages.Ack, error) {
	cm.receivedHost = host
	cm.receivedMessage = message
	cm.numRuns++
	return cm.returnedMessage, cm.returnedError
}

func (cm *commsMock) UnregisterToken(host *connect.Host,
	message *pb.UnregisterTokenRequest) (*messages.Ack, error) {
	cm.receivedHost = host
	cm.receivedMessage = message
	cm.numRuns++
	return cm.returnedMessage, cm.returnedError
}

func (cm *commsMock) RegisterTrackedID(host *connect.Host,
	message *pb.RegisterTrackedIdRequest) (*messages.Ack, error) {
	cm.receivedHost = host
	cm.receivedMessage = message
	cm.numRuns++
	return cm.returnedMessage, cm.returnedError
}

func (cm *commsMock) UnregisterTrackedID(host *connect.Host,
	message *pb.UnregisterTrackedIdRequest) (*messages.Ack, error) {
	cm.receivedHost = host
	cm.receivedMessage = message
	cm.numRuns++
	return cm.returnedMessage, cm.returnedError
}

func (cm *commsMock) GetHost(*id.ID) (*connect.Host, bool) {
	return &connect.Host{}, true
}
