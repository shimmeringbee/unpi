package broker

import (
	"bytes"
	"context"
	"errors"
	"github.com/shimmeringbee/bytecodec"
	"github.com/shimmeringbee/unpi"
	"github.com/shimmeringbee/unpi/library"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	t.Run("async outgoing request writes bytes", func(t *testing.T) {
		writer := bytes.Buffer{}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		err := z.AsyncRequest(f)
		assert.NoError(t, err)

		expectedFrame := f.Marshall()
		actualFrame := writer.Bytes()

		assert.Equal(t, expectedFrame, actualFrame)
	})

	t.Run("async outgoing request with non async request errors", func(t *testing.T) {
		writer := bytes.Buffer{}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		err := z.AsyncRequest(f)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameNotAsynchronous))
	})

	t.Run("async outgoing request passes error back to caller", func(t *testing.T) {
		expectedError := errors.New("error")

		writer := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return 0, expectedError
			},
		}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		actualError := z.AsyncRequest(f)
		assert.Error(t, actualError)
		assert.Equal(t, expectedError, actualError)
	})

	t.Run("requesting a sync send with a non sync frame errors", func(t *testing.T) {
		reader := bytes.Buffer{}
		writer := bytes.Buffer{}

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		_, err := z.SyncRequest(context.Background(), f)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameNotSynchronous))
	})

	t.Run("sync requests are sent to unpi and reply is read", func(t *testing.T) {
		responseFrame := unpi.Frame{
			MessageType: unpi.SRSP,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{},
		}
		responseBytes := responseFrame.Marshall()

		beenWrittenBuffer := bytes.Buffer{}
		r, w := io.Pipe()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				beenWrittenBuffer.Write(p)
				go func() { w.Write(responseBytes) }()
				return len(p), nil
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := NewBroker(&device, &device, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		actualResponseFrame, err := z.SyncRequest(context.Background(), f)
		assert.NoError(t, err)

		expectedFrame := f.Marshall()
		actualFrame := beenWrittenBuffer.Bytes()

		assert.Equal(t, expectedFrame, actualFrame)
		assert.Equal(t, responseFrame, actualResponseFrame)
	})

	t.Run("sync outgoing request passes error during write back to caller", func(t *testing.T) {
		expectedError := errors.New("error")

		reader := bytes.Buffer{}
		writer := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return 0, expectedError
			},
		}

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		_, actualError := z.SyncRequest(context.Background(), f)
		assert.Error(t, actualError)
		assert.Equal(t, expectedError, actualError)
	})

	t.Run("sync outgoing context cancellation causes function to error", func(t *testing.T) {
		reader := EmptyReader{End: make(chan bool)}
		writer := bytes.Buffer{}

		z := NewBroker(&reader, &writer, nil)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		defer cancel()

		_, actualError := z.SyncRequest(ctx, f)
		assert.Error(t, actualError)
		assert.Equal(t, SyncRequestContextCancelled, actualError)
	})
}

func TestMessageRequestResponse(t *testing.T) {
	t.Run("verifies send with receive on synchronous messages are handled", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct {
			Value uint16
		}

		type Response struct {
			Value uint8
		}

		ml.Add(unpi.SREQ, unpi.SYS, 0x01, Request{})
		ml.Add(unpi.SRSP, unpi.SYS, 0x02, Response{})

		sentMessage := Request{
			Value: 0x2040,
		}

		expectedMessage := Response{
			Value: 0x24,
		}

		payloadBytes, err := bytecodec.Marshall(expectedMessage)
		assert.NoError(t, err)

		frame := unpi.Frame{
			MessageType: unpi.SRSP,
			Subsystem:   unpi.SYS,
			CommandID:   0x02,
			Payload:     payloadBytes,
		}

		responseBytes := frame.Marshall()
		writtenBytes := bytes.Buffer{}

		r, w := io.Pipe()
		defer w.Close()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				go func() { w.Write(responseBytes) }()
				return writtenBytes.Write(p)
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := NewBroker(&device, &device, ml)
		defer z.Stop()

		actualReceivedMessage := Response{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err = z.RequestResponse(ctx, sentMessage, &actualReceivedMessage)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, actualReceivedMessage)

		frame, _ = unpi.UnmarshallFrame(writtenBytes.Bytes())

		assert.Equal(t, unpi.SREQ, frame.MessageType)
		assert.Equal(t, unpi.SYS, frame.Subsystem)
		assert.Equal(t, byte(0x01), frame.CommandID)

		actualSentMessage := Request{}
		err = bytecodec.Unmarshall(frame.Payload, &actualSentMessage)
		assert.NoError(t, err)

		assert.Equal(t, sentMessage, actualSentMessage)
	})

	t.Run("verifies send with receive on asynchronous messages are handled", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct {
			Value uint16
		}

		type Response struct {
			Value uint8
		}

		ml.Add(unpi.AREQ, unpi.SYS, 0x01, Request{})
		ml.Add(unpi.AREQ, unpi.SYS, 0x02, Response{})

		sentMessage := Request{
			Value: 0x2040,
		}

		expectedMessage := Response{
			Value: 0x24,
		}

		payloadBytes, err := bytecodec.Marshall(expectedMessage)
		assert.NoError(t, err)

		frame := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.SYS,
			CommandID:   0x02,
			Payload:     payloadBytes,
		}

		responseBytes := frame.Marshall()
		writtenBytes := bytes.Buffer{}

		r, w := io.Pipe()
		defer w.Close()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				go func() { w.Write(responseBytes) }()
				return writtenBytes.Write(p)
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := NewBroker(&device, &device, ml)
		defer z.Stop()

		actualReceivedMessage := Response{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err = z.RequestResponse(ctx, sentMessage, &actualReceivedMessage)

		assert.NoError(t, err)
		assert.Equal(t, expectedMessage, actualReceivedMessage)

		frame, _ = unpi.UnmarshallFrame(writtenBytes.Bytes())

		assert.Equal(t, unpi.AREQ, frame.MessageType)
		assert.Equal(t, unpi.SYS, frame.Subsystem)
		assert.Equal(t, byte(0x01), frame.CommandID)

		actualSentMessage := Request{}
		err = bytecodec.Unmarshall(frame.Payload, &actualSentMessage)
		assert.NoError(t, err)

		assert.Equal(t, sentMessage, actualSentMessage)
	})
}

type EmptyReader struct {
	End chan bool
}

func (e *EmptyReader) Done() {
	e.End <- true
}

func (e *EmptyReader) Read(p []byte) (n int, err error) {
	<-e.End

	return 0, io.EOF
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
