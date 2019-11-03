package unpi

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFrame_Marshall(t *testing.T) {
	t.Run("marshall empty payload", func(t *testing.T) {
		frame := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{},
		}

		expected := []byte{0xfe, 0x00, 0x25, 0x37, 0x12}

		actual := frame.Marshall()

		assert.Equal(t, expected, actual)
	})

	t.Run("marshall nil payload", func(t *testing.T) {
		frame := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     nil,
		}

		expected := []byte{0xfe, 0x00, 0x25, 0x37, 0x12}

		actual := frame.Marshall()

		assert.Equal(t, expected, actual)
	})

	t.Run("marshall with payload", func(t *testing.T) {
		frame := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		expected := []byte{0xfe, 0x02, 0x25, 0x37, 0x55, 0xdd, 0x98}

		actual := frame.Marshall()

		assert.Equal(t, expected, actual)
	})
}

func TestUnmarshall(t *testing.T) {
	t.Run("unmarshall frame with no payload", func(t *testing.T) {
		expected := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{},
		}

		asBytes := expected.Marshall()
		actual, err := UnmarshallFrame(asBytes)

		assert.NoError(t, err)
		assert.Equal(t, &expected, actual)
	})

	t.Run("unmarshall frame with payload", func(t *testing.T) {
		expected := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		asBytes := expected.Marshall()
		actual, err := UnmarshallFrame(asBytes)

		assert.NoError(t, err)
		assert.Equal(t, &expected, actual)
	})

	t.Run("unmarshall frame with invalid checksum", func(t *testing.T) {
		frame := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		asBytes := frame.Marshall()
		asBytes[len(asBytes)-1] = ^asBytes[len(asBytes)-1]

		_, err := UnmarshallFrame(asBytes)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameChecksumFailed))
	})

	t.Run("unmarshall frame which is too short for no payload", func(t *testing.T) {
		asBytes := []byte{StartOfFrame, 0x00}

		_, err := UnmarshallFrame(asBytes)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameTooShort))
	})

	t.Run("unmarshall frame which is too short for with a payload", func(t *testing.T) {
		asBytes := []byte{StartOfFrame, 0x02, 0x00, 0x00, 0x00, 0x00}

		_, err := UnmarshallFrame(asBytes)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameTooShort))
	})

	t.Run("unmarshall frame which is missing its start of frame header", func(t *testing.T) {
		asBytes := []byte{0x00, 0x00, 0x00, 0x00, 0x00}

		_, err := UnmarshallFrame(asBytes)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameMissingStartOfFrame))
	})
}
