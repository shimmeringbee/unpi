package broker

import (
	"bytes"
	"github.com/shimmeringbee/unpi"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBroker_Receive(t *testing.T) {
	t.Run("receive frames from unpi", func(t *testing.T) {
		reader := bytes.Buffer{}
		writer := bytes.Buffer{}

		expectedFrameOne := unpi.Frame{
			MessageType: 0,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		expectedFrameTwo := unpi.Frame{
			MessageType: 0,
			Subsystem:   unpi.SYS,
			CommandID:   2,
			Payload:     []byte{},
		}

		reader.Write(expectedFrameOne.Marshall())
		reader.Write(expectedFrameTwo.Marshall())

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		frame, err := z.Receive()
		assert.NoError(t, err)
		assert.Equal(t, expectedFrameOne, frame)

		frame, err = z.Receive()
		assert.NoError(t, err)
		assert.Equal(t, expectedFrameTwo, frame)
	})
}
