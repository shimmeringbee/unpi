package broker

import (
	. "github.com/shimmeringbee/unpi"
	"github.com/shimmeringbee/unpi/library"
	testunpi "github.com/shimmeringbee/unpi/testing"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	t.Run("sends asynchronous request", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct{}

		ml.Add(AREQ, SYS, 0x01, Request{})

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		b.Start()
		defer b.Stop()

		m.On(AREQ, SYS, 0x01)

		request := Request{}
		err := b.Request(request)

		time.Sleep(10 * time.Millisecond)

		assert.NoError(t, err)

		m.AssertCalls(t)
	})

	t.Run("synchronous request return an error", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct{}

		ml.Add(SREQ, SYS, 0x01, Request{})

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		b.Start()
		defer b.Stop()

		request := Request{}
		err := b.Request(request)

		assert.Error(t, err)

		m.AssertCalls(t)
	})

	t.Run("unrecognised messages return an error", func(t *testing.T) {
		ml := library.NewLibrary()

		type Request struct{}

		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		b.Start()
		defer b.Stop()

		request := Request{}
		err := b.Request(request)

		assert.Error(t, err)

		m.AssertCalls(t)
	})
}
