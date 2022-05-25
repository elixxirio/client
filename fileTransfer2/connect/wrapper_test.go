////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package connect

import (
	"bytes"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/connect"
	"gitlab.com/elixxir/client/e2e/receive"
	ft "gitlab.com/elixxir/client/fileTransfer2"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/elixxir/crypto/fastRNG"
	ftCrypto "gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/ekv"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
	"math"
	"sync"
	"testing"
	"time"
)

// Tests that Connection adheres to the connect.Connection interface.
var _ Connection = (connect.Connection)(nil)

// Smoke test of the entire file transfer system.
func Test_FileTransfer_Smoke(t *testing.T) {
	// jww.SetStdoutThreshold(jww.LevelDebug)
	// Set up cMix and E2E message handlers
	cMixHandler := newMockCmixHandler()
	e2eHandler := newMockConnectionHandler()
	rngGen := fastRNG.NewStreamGenerator(1000, 10, csprng.NewSystemRNG)
	ftParams := ft.DefaultParams()
	ftParams.MaxThroughput = math.MaxInt
	params := DefaultParams()

	type receiveCbValues struct {
		tid      *ftCrypto.TransferID
		fileName string
		fileType string
		sender   *id.ID
		size     uint32
		preview  []byte
	}

	// Set up the first client
	receiveCbChan1 := make(chan receiveCbValues, 10)
	receiveCB1 := func(tid *ftCrypto.TransferID, fileName, fileType string,
		sender *id.ID, size uint32, preview []byte) {
		receiveCbChan1 <- receiveCbValues{
			tid, fileName, fileType, sender, size, preview}
	}
	myID1 := id.NewIdFromString("myID1", id.User, t)
	kv1 := versioned.NewKV(ekv.MakeMemstore())
	endE2eChan1 := make(chan receive.Message, 3)
	conn1 := newMockConnection(myID1, e2eHandler)
	conn1.RegisterListener(catalog.EndFileTransfer, newMockListener(endE2eChan1))
	cmix1 := newMockCmix(myID1, cMixHandler)
	ftManager1, err := ft.NewManager(ftParams, myID1, cmix1, kv1, rngGen)
	if err != nil {
		t.Errorf("Failed to make new file transfer manager: %+v", err)
	}
	stop1, err := ftManager1.StartProcesses()
	if err != nil {
		t.Errorf("Failed to start processes for manager 1: %+v", err)
	}
	m1, err := NewWrapper(receiveCB1, params, ftManager1, conn1, cmix1)
	if err != nil {
		t.Errorf("Failed to create new file transfer manager 1: %+v", err)
	}

	// Set up the second client
	receiveCbChan2 := make(chan receiveCbValues, 10)
	receiveCB2 := func(tid *ftCrypto.TransferID, fileName, fileType string,
		sender *id.ID, size uint32, preview []byte) {
		receiveCbChan2 <- receiveCbValues{
			tid, fileName, fileType, sender, size, preview}
	}
	myID2 := id.NewIdFromString("myID2", id.User, t)
	kv2 := versioned.NewKV(ekv.MakeMemstore())
	endE2eChan2 := make(chan receive.Message, 3)
	conn2 := newMockConnection(myID2, e2eHandler)
	conn2.RegisterListener(catalog.EndFileTransfer, newMockListener(endE2eChan2))
	cmix2 := newMockCmix(myID1, cMixHandler)
	ftManager2, err := ft.NewManager(ftParams, myID2, cmix2, kv2, rngGen)
	if err != nil {
		t.Errorf("Failed to make new file transfer manager: %+v", err)
	}
	stop2, err := ftManager2.StartProcesses()
	if err != nil {
		t.Errorf("Failed to start processes for manager 2: %+v", err)
	}
	m2, err := NewWrapper(receiveCB2, params, ftManager2, conn2, cmix2)
	if err != nil {
		t.Errorf("Failed to create new file transfer manager 2: %+v", err)
	}

	// Wait group prevents the test from quiting before the file has completed
	// sending and receiving
	var wg sync.WaitGroup

	// Define details of file to send
	fileName, fileType := "myFile", "txt"
	fileData := []byte(loremIpsum)
	preview := []byte("Lorem ipsum dolor sit amet")
	retry := float32(2.0)

	// Create go func that waits for file transfer to be received to register
	// a progress callback that then checks that the file received is correct
	// when done
	wg.Add(1)
	var called bool
	timeReceived := make(chan time.Time)
	go func() {
		select {
		case r := <-receiveCbChan2:
			receiveProgressCB := func(completed bool, received, total uint16,
				rt ft.ReceivedTransfer, fpt ft.FilePartTracker, err error) {
				if completed && !called {
					timeReceived <- netTime.Now()
					receivedFile, err2 := m2.Receive(r.tid)
					if err2 != nil {
						t.Errorf("Failed to receive file: %+v", err2)
					}

					if !bytes.Equal(fileData, receivedFile) {
						t.Errorf("Received file does not match sent."+
							"\nsent:     %q\nreceived: %q",
							fileData, receivedFile)
					}
					wg.Done()
				}
			}
			err3 := m2.RegisterReceivedProgressCallback(
				r.tid, receiveProgressCB, 0)
			if err3 != nil {
				t.Errorf(
					"Failed to Rregister received progress callback: %+v", err3)
			}
		case <-time.After(2100 * time.Millisecond):
			t.Errorf("Timed out waiting to receive new file transfer.")
			wg.Done()
		}
	}()

	// Define sent progress callback
	wg.Add(1)
	sentProgressCb1 := func(completed bool, arrived, total uint16,
		st ft.SentTransfer, fpt ft.FilePartTracker, err error) {
		if completed {
			wg.Done()
		}
	}

	// Send file
	sendStart := netTime.Now()
	tid1, err := m1.Send(
		myID2, fileName, fileType, fileData, retry, preview, sentProgressCb1, 0)
	if err != nil {
		t.Errorf("Failed to send file: %+v", err)
	}

	go func() {
		select {
		case tr := <-timeReceived:
			fileSize := len(fileData)
			sendTime := tr.Sub(sendStart)
			fileSizeKb := float32(fileSize) * .001
			speed := fileSizeKb * float32(time.Second) / (float32(sendTime))
			t.Logf("Completed receiving file %q in %s (%.2f kb @ %.2f kb/s).",
				fileName, sendTime, fileSizeKb, speed)
		}
	}()

	// Wait for file to be sent and received
	wg.Wait()

	select {
	case <-endE2eChan2:
	case <-time.After(15 * time.Millisecond):
		t.Errorf("Timed out waiting for end file transfer message.")
	}

	err = m1.CloseSend(tid1)
	if err != nil {
		t.Errorf("Failed to close transfer: %+v", err)
	}

	err = stop1.Close()
	if err != nil {
		t.Errorf("Failed to close processes for manager 1: %+v", err)
	}

	err = stop2.Close()
	if err != nil {
		t.Errorf("Failed to close processes for manager 2: %+v", err)
	}
}

const loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed sit amet urna venenatis, rutrum magna maximus, tempor orci. Cras sit amet nulla id dolor blandit commodo. Suspendisse potenti. Praesent gravida porttitor metus vel aliquam. Maecenas rutrum velit at lobortis auctor. Mauris porta blandit tempor. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Morbi volutpat posuere maximus. Nunc in augue molestie ante mattis tempor.

Phasellus placerat elit eu fringilla pharetra. Vestibulum consectetur pulvinar nunc, vestibulum tincidunt felis rhoncus sit amet. Duis non dolor eleifend nibh luctus eleifend. Nunc urna odio, euismod sit amet feugiat ut, dapibus vel elit. Nulla est mauris, posuere eget enim cursus, vehicula viverra est. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Quisque mattis, nisi quis consectetur semper, neque enim rhoncus dolor, ut aliquam leo orci sed dolor. Integer ullamcorper pulvinar turpis, a sollicitudin nunc posuere et. Nullam orci nibh, facilisis ac massa eu, bibendum bibendum sapien. Sed tincidunt nunc mauris, nec ullamcorper enim lacinia nec. Nulla dapibus sapien ut odio bibendum, tempus ornare sapien lacinia.

Duis ac hendrerit augue. Nullam porttitor feugiat finibus. Nam enim urna, maximus et ligula eu, aliquet convallis turpis. Vestibulum luctus quam in dictum efficitur. Vestibulum ac pulvinar ipsum. Vivamus consectetur augue nec tellus mollis, at iaculis magna efficitur. Nunc dictum convallis sem, at vehicula nulla accumsan non. Nullam blandit orci vel turpis convallis, mollis porttitor felis accumsan. Sed non posuere leo. Proin ultricies varius nulla at ultricies. Phasellus et pharetra justo. Quisque eu orci odio. Pellentesque pharetra tempor tempor. Aliquam ac nulla lorem. Sed dignissim ligula sit amet nibh fermentum facilisis.

Donec facilisis rhoncus ante. Duis nec nisi et dolor congue semper vel id ligula. Mauris non eleifend libero, et sodales urna. Nullam pharetra gravida velit non mollis. Integer vel ultrices libero, at ultrices magna. Duis semper risus a leo vulputate consectetur. Cras sit amet convallis sapien. Sed blandit, felis et porttitor fringilla, urna tellus commodo metus, at pharetra nibh urna sed sem. Nam ex dui, posuere id mi et, egestas tincidunt est. Nullam elementum pulvinar diam in maximus. Maecenas vel augue vitae nunc consectetur vestibulum in aliquet lacus. Nullam nec lectus dapibus, dictum nisi nec, congue quam. Suspendisse mollis vel diam nec dapibus. Mauris neque justo, scelerisque et suscipit non, imperdiet eget leo. Vestibulum leo turpis, dapibus ac lorem a, mollis pulvinar quam.

Sed sed mauris a neque dignissim aliquet. Aliquam congue gravida velit in efficitur. Integer elementum feugiat est, ac lacinia libero bibendum sed. Sed vestibulum suscipit dignissim. Nunc scelerisque, turpis quis varius tristique, enim lacus vehicula lacus, id vestibulum velit erat eu odio. Donec tincidunt nunc sit amet sapien varius ornare. Phasellus semper venenatis ligula eget euismod. Mauris sodales massa tempor, cursus velit a, feugiat neque. Sed odio justo, rhoncus eu fermentum non, tristique a quam. In vehicula in tortor nec iaculis. Cras ligula sem, sollicitudin at nulla eget, placerat lacinia massa. Mauris tempus quam sit amet leo efficitur egestas. Proin iaculis, velit in blandit egestas, felis odio sollicitudin ipsum, eget interdum leo odio tempor nisi. Curabitur sed mauris id turpis tempor finibus ut mollis lectus. Curabitur neque libero, aliquam facilisis lobortis eget, posuere in augue. In sodales urna sit amet elit euismod rhoncus.`