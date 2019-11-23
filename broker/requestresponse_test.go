package broker

import (
	"context"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
	"github.com/shimmeringbee/unpi/library"
	testunpi "github.com/shimmeringbee/unpi/testing"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRequestResponse(t *testing.T) {
	t.Run("sends asynchronous request and waits for the response", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct {
		}

		type Response struct {
			Value uint8
		}

		ml.Add(AREQ, SYS, 0x01, Request{})
		ml.Add(AREQ, SYS, 0x02, Response{})

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		defer b.Stop()

		expectedResponse := Response{Value: 42}
		data, _ := bytecodec.Marshall(expectedResponse)

		m.On(AREQ, SYS, 0x01).Return(Frame{
			MessageType: AREQ,
			Subsystem:   SYS,
			CommandID:   0x02,
			Payload:     data,
		})

		request := Request{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		actualResponse := Response{}
		err := b.RequestResponse(ctx, request, &actualResponse)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)

		m.AssertCalls(t)
	})

	t.Run("sends synchronous request and waits for the response", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct {
		}

		type Response struct {
			Value uint8
		}

		ml.Add(SREQ, SYS, 0x01, Request{})
		ml.Add(SRSP, SYS, 0x02, Response{})

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		defer b.Stop()

		expectedResponse := Response{Value: 42}
		data, _ := bytecodec.Marshall(expectedResponse)

		m.On(SREQ, SYS, 0x01).Return(Frame{
			MessageType: SRSP,
			Subsystem:   SYS,
			CommandID:   0x02,
			Payload:     data,
		})

		request := Request{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		actualResponse := Response{}
		err := b.RequestResponse(ctx, request, &actualResponse)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)

		m.AssertCalls(t)
	})

	t.Run("sends request with no response, context deadline takes effect", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct {
		}

		type Response struct {
		}

		ml.Add(SREQ, SYS, 0x01, Request{})
		ml.Add(SRSP, SYS, 0x01, Response{})

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		defer b.Stop()

		m.On(SREQ, SYS, 0x01)

		request := Request{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		actualResponse := Response{}
		err := b.RequestResponse(ctx, request, &actualResponse)

		assert.Error(t, err)

		m.AssertCalls(t)
	})
}
