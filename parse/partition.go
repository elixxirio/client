////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package parse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"gitlab.com/elixxir/client/globals"
	"gitlab.com/elixxir/primitives/format"
	"math"
	"sync"
	"gitlab.com/elixxir/crypto/e2e"
)

// TODO is there a better way to generate unique message IDs locally?
func IDCounter() func() []byte {
	// 32 bits to put a smaller upper bound on the varint size on the wire
	// It should be possible to go back down to just a byte for the message ID
	// field once the message format includes a timestamp, since you won't send
	// more than 256 individual messages to one other user in a second. If we
	// can use that timestamp information to help identify different multi-part
	// messages, the currentID becomes a lot more superfluous.
	currentID := uint32(0)
	var mux sync.Mutex
	return func() []byte {
		// this will use up to 5 bytes for the message ID
		result := make([]byte, binary.MaxVarintLen32)
		mux.Lock()
		n := binary.PutUvarint(result, uint64(currentID))
		currentID++
		mux.Unlock()
		return result[:n]
	}
}

const MessageTooLongError = "Partition(): Message is too long to partition"

// length in bytes of index and max index together
// change this if you change the index type
const IndexLength = 2

// The max index is 0 for one partition, 1 for two partitions, 2 for three
// partitions, and so on. This assumes that each partition can be completely
// filled.
func GetMaxIndex(body []byte, id []byte) int32 {
	bodyLen := len(body)
	if bodyLen > 0 {
		bodyLen--
	}
	maxIndex := bodyLen / (format.MP_PAYLOAD_LEN - e2e.MinPaddingLen - len(id) - IndexLength)
	return int32(maxIndex)
}

func Partition(body []byte, id []byte) ([][]byte, error) {
	// index and quantity of the partitioned message are a fixed length of 8
	// bits because sending more than that through the system is really slow and
	// making them variable length makes the required length of the body part
	// of the partitions different per partition depending on what the length
	// of the index is for the input message
	// the bottom line is that there's a dependency cycle to calculate the right
	// number of partitions if you do them that way and i'm having none of that

	// a zero here means that the message has one partition
	maxIndex := GetMaxIndex(body, id)
	if maxIndex > math.MaxUint8 {
		return nil, errors.New(MessageTooLongError)
	}

	partitions := make([][]byte, maxIndex+1)
	var lastPartitionLength int
	partitionReadIdx := 0
	for i := range partitions {
		maxPartitionLength := format.MP_PAYLOAD_LEN-e2e.MinPaddingLen
		partitions[i], lastPartitionLength = makePartition(maxPartitionLength,
			body[partitionReadIdx:], id, byte(i), byte(maxIndex))
		partitionReadIdx += lastPartitionLength
	}

	var file []byte
	for i := range partitions {
		file = append(file, []byte(fmt.Sprintf("%q\n", partitions[i]))...)
	}

	return partitions, nil
}

// can you believe that golang doesn't provide a min function in the std lib?
// neither can i
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// makePartition makes a new partition of a multi-part message and prepends the
// id, index, and length that are needed to rebuild it on the receiving client.
// It returns the new partition and the length of the body that it consumed
// when making the new partition.
func makePartition(maxLength int, body []byte, id []byte, i byte,
	maxIndex byte) ([]byte, int) {

	partition := make([]byte, 0, maxLength)

	// Append the front matter
	partition = append(partition, id...)
	partition = append(partition, i, maxIndex)
	lengthBeforeBodyAppend := len(partition)

	// Find the biggest part of the body that can fit into the message length
	bodyWriteLength := min(maxLength-len(partition), len(body))

	// Append body
	partition = append(partition, body[:bodyWriteLength]...)

	// Return new partition and number of bytes from the body that are in it
	return partition, len(partition) - lengthBeforeBodyAppend
}

// Assemble assumes that messages are correctly ordered by their index
// It also assumes that messages have had all of their front matter and
// padding stripped.
func Assemble(partitions [][]byte) ([]byte, error) {
	// this will allocate a bit more capacity than needed but not so much that
	// it breaks the bank
	result := make([]byte, 0, int(format.MP_PAYLOAD_LEN)*len(partitions))

	for i := range partitions {
		result = append(result, partitions[i]...)
	}
	return result, nil
}

type MultiPartMessage struct {
	ID       []byte
	Index    byte
	MaxIndex byte
	Body     []byte
}

func ValidatePartition(partition []byte) (message *MultiPartMessage,
	err error) {
	globals.Log.DEBUG.Printf("%v\n", partition)
	// ID is first, and it's variable length
	msbMask := byte(0x80)
	indexInformationStart := 0
	for i := 0; i < len(partition); i++ {
		if msbMask&partition[i] == 0 {
			// this is the last byte in the ID. stop the loop
			indexInformationStart = i + 1
			globals.Log.DEBUG.Println("Index information start:", indexInformationStart)
			break
		}
	}
	// validate: make sure that there's a payload beyond the front matter
	if indexInformationStart+IndexLength >= len(partition) {
		return nil, errors.New("There was nothing after the partition info")
		// make sure that the ID is within the length we expect
	} else if indexInformationStart > binary.MaxVarintLen32 {
		return nil, errors.New("ID was longer than expected")
		// make sure that the index is less than or equal to the maximum
	} else if partition[indexInformationStart] > partition[indexInformationStart+1] {
		return nil, errors.New(fmt.Sprintf(
			"Index %v was more than max index %v",
			partition[indexInformationStart],
			partition[indexInformationStart+1]))
		// make sure that we found a boundary between the index and ID
	} else if indexInformationStart == 0 {
		return nil, errors.New("Couldn't find end of ID")
	}

	result := &MultiPartMessage{
		ID:       partition[:indexInformationStart],
		Index:    partition[indexInformationStart],
		MaxIndex: partition[indexInformationStart+1],
		Body:     partition[indexInformationStart+2:],
	}

	globals.Log.DEBUG.Printf("Result of partition validation: %v, %v, %v, %v\n", result.ID,
		result.Index, result.MaxIndex, string(result.Body))
	return result, nil
}
