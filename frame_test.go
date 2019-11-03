package unpi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFrame_Marshall(t *testing.T) {
	t.Run("test empty payload", func(t *testing.T) {
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

	t.Run("test nil payload", func(t *testing.T) {
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

	t.Run("test with payload", func(t *testing.T) {
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
