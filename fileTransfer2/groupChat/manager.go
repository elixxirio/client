////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package groupChat

import (
	"github.com/pkg/errors"
	ft "gitlab.com/elixxir/client/fileTransfer2"
	"gitlab.com/elixxir/client/groupChat"
	"gitlab.com/elixxir/client/stoppable"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/elixxir/crypto/fastRNG"
	ftCrypto "gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/crypto/group"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

// Error messages.
const (
	// NewManager
	errNewFtManager = "cannot create new group chat file transfer manager: %+v"

	// Manager.StartProcesses
	errAddNewService = "failed to add service to receive new group file transfers: %+v"
)

const (
	// Tag used when sending/receiving new group chat file transfers message
	newFileTransferTag = "NewGroupFileTransfer"

	// Tag used when sending/receiving end group chat file transfers message
	endFileTransferTag = "EndGroupFileTransfer"
)

// Manager handles the sending and receiving of file transfers for group chats.
type Manager struct {
	// Callback that is called every time a new file transfer is received
	receiveCB ft.ReceiveCallback

	// File transfer Manager
	ft ft.FileTransfer

	// Group chat Manager
	gc GroupChat

	myID *id.ID
	cmix ft.Cmix
}

// GroupChat interface matches a subset of the groupChat.GroupChat methods used
// by the Manager for easier testing.
type GroupChat interface {
	Send(groupID *id.ID, tag string, message []byte) (
		id.Round, time.Time, group.MessageID, error)
	AddService(tag string, p groupChat.Processor) error
}

// NewManager generates a new file transfer Manager for group chat.
func NewManager(receiveCB ft.ReceiveCallback, params ft.Params, myID *id.ID,
	gc GroupChat, cmix ft.Cmix, kv *versioned.KV, rng *fastRNG.StreamGenerator) (
	*Manager, error) {

	ftManager, err := ft.NewManager(params, myID, cmix, kv, rng)
	if err != nil {
		return nil, errors.Errorf(errNewFtManager, err)
	}

	return &Manager{
		receiveCB: receiveCB,
		ft:        ftManager,
		gc:        gc,
		myID:      myID,
		cmix:      cmix,
	}, nil
}

func (m *Manager) StartProcesses() (stoppable.Stoppable, error) {
	err := m.gc.AddService(newFileTransferTag, &processor{m})
	if err != nil {
		return nil, errors.Errorf(errAddNewService, err)
	}

	return m.ft.StartProcesses()
}

func (m *Manager) MaxFileNameLen() int {
	return m.ft.MaxFileNameLen()
}

func (m *Manager) MaxFileTypeLen() int {
	return m.ft.MaxFileTypeLen()
}

func (m *Manager) MaxFileSize() int {
	return m.ft.MaxFileSize()
}

func (m *Manager) MaxPreviewSize() int {
	return m.ft.MaxPreviewSize()
}

func (m *Manager) Send(fileName, fileType string, fileData []byte,
	recipient *id.ID, retry float32, preview []byte,
	progressCB ft.SentProgressCallback, period time.Duration) (*ftCrypto.TransferID, error) {
	sendNew := func(info *ft.TransferInfo) error {
		return sendNewFileTransferMessage(recipient, info, m.gc)
	}

	return m.ft.Send(fileName, fileType, fileData, recipient, retry, preview,
		progressCB, period, sendNew)
}

func (m *Manager) RegisterSentProgressCallback(tid *ftCrypto.TransferID,
	progressCB ft.SentProgressCallback, period time.Duration) error {
	return m.ft.RegisterSentProgressCallback(tid, progressCB, period)
}

func (m *Manager) CloseSend(tid *ftCrypto.TransferID) error {
	return m.ft.CloseSend(tid)
}

func (m *Manager) HandleIncomingTransfer(fileName string,
	key *ftCrypto.TransferKey, transferMAC []byte, numParts uint16, size uint32,
	retry float32, progressCB ft.ReceivedProgressCallback,
	period time.Duration) (*ftCrypto.TransferID, error) {
	return m.ft.HandleIncomingTransfer(
		fileName, key, transferMAC, numParts, size, retry, progressCB, period)
}

func (m *Manager) RegisterReceivedProgressCallback(tid *ftCrypto.TransferID,
	progressCB ft.ReceivedProgressCallback, period time.Duration) error {
	return m.ft.RegisterReceivedProgressCallback(tid, progressCB, period)
}

func (m *Manager) Receive(tid *ftCrypto.TransferID) ([]byte, error) {
	return m.ft.Receive(tid)
}
