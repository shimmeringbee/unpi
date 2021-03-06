package broker

import (
	. "github.com/shimmeringbee/unpi"
	"github.com/shimmeringbee/unpi/library"
	testunpi "github.com/shimmeringbee/unpi/testing"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBroker_listen(t *testing.T) {
	t.Run("listen calls multiple listeners that match", func(t *testing.T) {
		ml := library.NewLibrary()
		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		b.Start()
		defer b.Stop()

		awaitOneMatch := false
		awaitTwoMatch := false

		b.listen(SREQ, SYS, 0x02, func(frame Frame) {
			awaitOneMatch = true
		})

		b.listen(SREQ, SYS, 0x02, func(frame Frame) {
			awaitTwoMatch = true
		})

		m.InjectOutgoing(Frame{
			MessageType: SREQ,
			Subsystem:   SYS,
			CommandID:   0x02,
			Payload:     nil,
		})

		time.Sleep(10 * time.Millisecond)

		assert.True(t, awaitOneMatch)
		assert.True(t, awaitTwoMatch)
	})

	t.Run("listen ignores unrelated frames", func(t *testing.T) {
		ml := library.NewLibrary()
		m := testunpi.NewMockAdapter()
		defer m.Stop()
		b := NewBroker(m, m, ml)
		b.Start()
		defer b.Stop()

		awaitOneMatch := false

		b.listen(SREQ, SYS, 0x02, func(frame Frame) {
			awaitOneMatch = true
		})

		m.InjectOutgoing(Frame{
			MessageType: SREQ,
			Subsystem:   SYS,
			CommandID:   0x03,
			Payload:     nil,
		})

		assert.False(t, awaitOneMatch)
	})
}
