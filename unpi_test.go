package unpi

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUNPI_Read(t *testing.T) {
	t.Run("test valid frame is decoded", func(t *testing.T) {
		expected := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		data := expected.Marshall()
		device := bytes.NewBuffer(data)

		actual, err := Read(device)

		assert.NoError(t, err)
		assert.Equal(t, &expected, actual)
	})

	t.Run("test valid frame is decoded prefixed by junk", func(t *testing.T) {
		expected := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		data := []byte{0x01, 0x02, 0x03}
		data = append(data, expected.Marshall()...)
		device := bytes.NewBuffer(data)

		actual, err := Read(device)

		assert.NoError(t, err)
		assert.Equal(t, &expected, actual)
	})

	t.Run("test frame with invalid checksum raises error", func(t *testing.T) {
		expected := Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{0x55, 0xdd},
		}

		data := expected.Marshall()
		data[len(data)-1] = ^data[len(data)-1]

		device := bytes.NewBuffer(data)

		_, err := Read(device)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameChecksumFailed))
	})

	t.Run("test errors raised by reader are raised", func(t *testing.T) {
		originalError := errors.New("original")

		device := ControllableReaderWriter{
			Reader: func(p []byte) (n int, err error) {
				return 0, originalError
			},
			Writer: nil,
		}

		_, err := Read(&device)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, originalError))
	})
}

func TestUNPI_Write(t *testing.T) {
	t.Run("test frames are written to device", func(t *testing.T) {
		frame := &Frame{
			MessageType: SREQ,
			Subsystem:   ZDO,
			CommandID:   0x37,
			Payload:     []byte{},
		}

		expected := frame.Marshall()

		device := bytes.Buffer{}

		_ = Write(&device, frame)

		assert.Equal(t, expected, device.Bytes())
	})

	t.Run("test errors raised by writer are raised", func(t *testing.T) {
		frame := &Frame{}

		originalError := errors.New("original")

		device := ControllableReaderWriter{
			Reader: nil,
			Writer: func(p []byte) (n int, err error) {
				return 0, originalError
			},
		}

		err := Write(&device, frame)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, originalError))
	})

	t.Run("test errors raised by writer failing to write whole frame", func(t *testing.T) {
		frame := &Frame{}

		device := ControllableReaderWriter{
			Reader: nil,
			Writer: func(p []byte) (n int, err error) {
				return 2, nil
			},
		}

		err := Write(&device, frame)

		assert.Error(t, err)
	})
}

type ControllableReaderWriter struct {
	Reader func(p []byte) (n int, err error)
	Writer func(p []byte) (n int, err error)
}

func (c *ControllableReaderWriter) Read(p []byte) (n int, err error) {
	return c.Reader(p)
}

func (c *ControllableReaderWriter) Write(p []byte) (n int, err error) {
	return c.Writer(p)
}
