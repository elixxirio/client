package collective

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// xxdkDeviceOffsetHeader is the header of the device offset.
const xxdkDeviceOffsetHeader = "XXDKTXLOGDVCOFFST"

// deviceOffset is the last index a certain instance ID has Read.
type deviceOffset map[InstanceID]int

func newDeviceOffset() deviceOffset {
	return make(deviceOffset, 0)
}

// serialize serializes the deviceOffset object.
func (d deviceOffset) serialize() ([]byte, error) {
	dataToSave := make(map[string]int, len(d))
	for k, v := range d {
		dataToSave[k.String()] = v
	}
	deviceOffsetMarshal, err := json.Marshal(dataToSave)
	if err != nil {
		return nil, errors.Errorf("failed to marshal device offsets: %+v", err)
	}

	deviceOffsetInfo := xxdkDeviceOffsetHeader +
		base64.URLEncoding.EncodeToString(deviceOffsetMarshal)

	return []byte(deviceOffsetInfo), err
}

func deserializeDeviceOffset(deviceOffsetSerial []byte) (deviceOffset, error) {
	// Extract the device offset
	splitter := strings.Split(string(deviceOffsetSerial), xxdkDeviceOffsetHeader)
	if len(splitter) != 2 {
		return nil, errors.Errorf("unexpected data is serialized device offset.")
	}

	// Decode device offset
	deviceOffsetJson, err := base64.URLEncoding.DecodeString(splitter[1])
	if err != nil {
		return nil, err
	}

	// Unmarshal offset
	dvcOffset := make(map[string]int)
	if err = json.Unmarshal(deviceOffsetJson, &dvcOffset); err != nil {
		return nil, err
	}

	dvcOff := make(deviceOffset, len(dvcOffset))
	for k, v := range dvcOffset {
		newK, err := NewInstanceIDFromString(k)
		if err != nil {
			return nil, err
		}
		dvcOff[newK] = v
	}

	return dvcOff, nil
}

// serializeInt is a utility function which serializes an integer into a byte
// slice.
//
// This is the inverse operation of deserializeInt.
func serializeInt(i int) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

// deserializeInt is a utility function which deserializes byte data into an
// integer.
//
// This is the inverse operation of serializeInt.
func deserializeInt(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
