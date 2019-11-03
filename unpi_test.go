package unpi

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

		u := New(&device)
		_ = u.Write(frame)

		assert.Equal(t, expected, device.Bytes())
	})

	t.Run("test errors raised by writer are raised", func(t *testing.T) {
		frame := &Frame{}

		originalError := errors.New("original")

		rw := ControllableReaderWriter{
			Reader: nil,
			Writer: func(p []byte) (n int, err error) {
				return 0, originalError
			},
		}

		u := New(&rw)

		err := u.Write(frame)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, originalError))
	})

	t.Run("test errors raised by writer failing to write whole frame", func(t *testing.T) {
		frame := &Frame{}

		rw := ControllableReaderWriter{
			Reader: nil,
			Writer: func(p []byte) (n int, err error) {
				return 2, nil
			},
		}

		u := New(&rw)

		err := u.Write(frame)
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
