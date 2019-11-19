package broker

import (
	"bytes"
	"context"
	"errors"
	"github.com/shimmeringbee/unpi"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

func TestBroker_AwaitMessage(t *testing.T) {
	t.Run("await message responds to multiple listeners when frame matches", func(t *testing.T) {
		r, w := io.Pipe()
		defer w.Close()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return len(p), nil
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := NewBroker(&device, &device, nil)
		defer z.Stop()

		expectedFrame := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.SYS,
			CommandID:   0x20,
			Payload:     []byte{},
		}

		wg := &sync.WaitGroup{}

		for i := 0; i < 2; i++ {
			wg.Add(1)

			go func() {
				ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()
				actualFrame, err := z.AwaitMessage(ctxWithTimeout, unpi.AREQ, unpi.SYS, 0x20)

				assert.NoError(t, err)
				assert.Equal(t, expectedFrame, actualFrame)

				wg.Done()
			}()
		}

		time.Sleep(10 * time.Millisecond)
		data := expectedFrame.Marshall()
		w.Write(data)

		wg.Wait()
	})

	t.Run("await message respects context timeout", func(t *testing.T) {
		reader := EmptyReader{End: make(chan bool)}
		defer reader.Done()

		writer := bytes.Buffer{}

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()
		_, err := z.AwaitMessage(ctxWithTimeout, unpi.AREQ, unpi.SYS, 0x20)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, AwaitMessageContextCancelled))
	})

	t.Run("await message ignores unrelated frames", func(t *testing.T) {
		r, w := io.Pipe()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return len(p), nil
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := NewBroker(&device, &device, nil)
		defer z.Stop()

		expectedFrame := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.SYS,
			CommandID:   0x21,
			Payload:     nil,
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			data := expectedFrame.Marshall()
			w.Write(data)
		}()

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, err := z.AwaitMessage(ctxWithTimeout, unpi.AREQ, unpi.SYS, 0x20)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, AwaitMessageContextCancelled))
	})
}
