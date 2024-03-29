////////////////////////////////////////////////////////////////////////////////
// Copyright © 2023 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// sqlite requires cgo, which is not available in WASM.
//go:build !js || !wasm

package storage

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/require"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/xx_network/primitives/id"
	"os"
	"testing"
	"time"
)

type dummyCallbacks struct{}

func (d dummyCallbacks) MessageReceived(uint64, ed25519.PublicKey, bool, bool) {}
func (d dummyCallbacks) MessageDeleted(message.ID)                             {}

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelTrace)
	os.Exit(m.Run())
}

// Test simple receive of a new message for a new conversation.
func TestImpl_Receive(t *testing.T) {
	m, err := newImpl("TestImpl_Receive", &dummyCallbacks{}, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	testString := "test"
	testBytes := []byte(testString)
	partnerPubKey := ed25519.PublicKey(testBytes)
	testRound := id.Round(10)

	// Can use ChannelMessageID for ease, doesn't matter here
	testMsgId := message.DeriveChannelMessageID(&id.ID{1}, uint64(testRound), testBytes)

	// Receive a test message
	uuid := m.Receive(testMsgId, testString, testBytes,
		partnerPubKey, partnerPubKey, 0, 0, time.Now(),
		rounds.Round{ID: testRound}, dm.TextType, dm.Received)
	if uuid == 0 {
		t.Fatalf("Expected non-zero message uuid")
	}
	jww.DEBUG.Printf("Received test message: %d", uuid)

	// First, we expect a conversation to be created
	testConvo := m.GetConversation(partnerPubKey)
	if testConvo == nil {
		t.Fatalf("Expected conversation to be created")
	}
	// Spot check a conversation attribute
	if testConvo.Nickname != testString {
		t.Fatalf("Expected conversation nickname %s, got %s",
			testString, testConvo.Nickname)
	}

	// Next, we expect the message to be created
	testMessage := &Message{Id: int64(uuid)}
	err = m.db.Take(testMessage).Error
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Spot check a message attribute
	if !bytes.Equal(testMessage.SenderPubKey, partnerPubKey) {
		t.Fatalf("Expected message attibutes to match, expected %v got %v",
			partnerPubKey, testMessage.SenderPubKey)
	}
}

// Test happy path. Insert some conversations and check they exist.
func TestImpl_GetConversations(t *testing.T) {
	m, err := newImpl("TestImpl_GetConversations", &dummyCallbacks{}, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	numTestConvo := 10

	// Insert a test convo
	for i := 0; i < numTestConvo; i++ {
		testBytes := []byte(fmt.Sprintf("%d", i))
		testPubKey := ed25519.PublicKey(testBytes)
		err = m.upsertConversation("test", testPubKey,
			uint32(i), uint8(i), &time.Time{})
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	results := m.GetConversations()
	if len(results) != numTestConvo {
		t.Fatalf("Expected %d convos, got %d", numTestConvo, len(results))
	}

	for i, convo := range results {
		if convo.Token != uint32(i) {
			t.Fatalf("Expected %d convo token, got %d", i, convo.Token)
		}
		if convo.CodesetVersion != uint8(i) {
			t.Fatalf("Expected %d convo codeset, got %d",
				i, convo.CodesetVersion)
		}
	}
}

// Test failed and successful deletes
func TestWasmModel_DeleteMessage(t *testing.T) {
	m, err := newImpl("TestWasmModel_DeleteMessage", &dummyCallbacks{}, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Insert test message
	testBadBytes := []byte("uwu")
	testString := "test"
	testBytes := []byte(testString)
	partnerPubKey := ed25519.PublicKey(testBytes)
	testRound := id.Round(10)

	// Can use ChannelMessageID for ease, doesn't matter here
	testMsgId := message.DeriveChannelMessageID(&id.ID{1}, uint64(testRound), testBytes)

	// Receive a test message
	uuid := m.Receive(testMsgId, testString, testBytes,
		partnerPubKey, partnerPubKey, 0, 0, time.Now(),
		rounds.Round{ID: testRound}, dm.TextType, dm.Received)
	require.Positive(t, uuid)
	require.NoError(t, err)

	// Non-matching pub key, should fail to delete
	require.False(t, m.DeleteMessage(testMsgId, testBadBytes))

	// Correct pub key, should have deleted
	require.True(t, m.DeleteMessage(testMsgId, testBytes))
}
