package testing

import (
	. "github.com/shimmeringbee/unpi"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func TestMockAdapter_ReceivingFrames(t *testing.T) {
	t.Run("written frames have been stored in the mock", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, expectedFrame)
		assert.NoError(t, err)

		err = Write(m, expectedFrame)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		assert.Equal(t, 2, len(m.ReceivedFrames))
		assert.Equal(t, expectedFrame, m.ReceivedFrames[0])
		assert.Equal(t, expectedFrame, m.ReceivedFrames[1])
	})
}

func TestMockAdapter(t *testing.T) {
	t.Run("single mocked response return their correct Frame", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(SREQ, ZDO, 0xf0).Return(expectedFrame)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("mocked response with no return frame does nothing but record", func(t *testing.T) {
		m := NewMockAdapter()
		m.On(SREQ, ZDO, 0xf0)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		m.Stop()

		// Force stop to ensure the Read results in an EOF
		time.Sleep(20 * time.Millisecond)

		_, err = Read(m)
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("single mocked response errors if not called", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(SREQ, ZDO, 0xf0).Return(expectedFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.True(t, internalT.Failed())
	})

	t.Run("single mocked response return their correct frame if repeated, though assert fails due to only one call allowed", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(SREQ, ZDO, 0xf0).Return(expectedFrame)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		err = Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err = Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.True(t, internalT.Failed())
	})

	t.Run("single mocked response return their correct frame if repeated, multiple calls allowed", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(SREQ, ZDO, 0xf0).Return(expectedFrame).Times(2)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		err = Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err = Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("single mocked response return their correct frame if repeated, unlimited calls allowed", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(SREQ, ZDO, 0xf0).Return(expectedFrame).UnlimitedTimes()

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		err = Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err = Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("single mocked response with multiple returns will return the correct Frame if repeated", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrameOne := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		expectedFrameTwo := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x01}}

		m.On(SREQ, ZDO, 0xf0).Return(expectedFrameOne, expectedFrameTwo).Times(2)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrameOne, actualFrame)

		err = Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err = Read(m)
		assert.Equal(t, expectedFrameTwo, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("single mocked response return their correct frame if wild carded", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SRSP, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x00}}
		m.On(AnyType, AnySubsystem, AnyCommand).Return(expectedFrame)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		actualFrame, err := Read(m)
		assert.Equal(t, expectedFrame, actualFrame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)

		assert.False(t, internalT.Failed())
	})

	t.Run("multiple writes to the mock are captures and ordering is correct", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		c := m.On(SREQ, ZDO, 0xf0).Times(2)

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		err = Write(m, frame)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 2, len(c.CapturedCalls))

		assert.Equal(t, frame, c.CapturedCalls[0].Frame)
		assert.Equal(t, frame, c.CapturedCalls[1].Frame)

		assert.True(t, c.CapturedCalls[0].Before(c.CapturedCalls[1]))

		internalT := new(testing.T)
		c.CapturedCalls[0].AssertBefore(internalT, c.CapturedCalls[1])
		assert.False(t, internalT.Failed())

		assert.True(t, c.CapturedCalls[1].After(c.CapturedCalls[0]))

		internalT = new(testing.T)
		c.CapturedCalls[1].AssertAfter(internalT, c.CapturedCalls[0])
		assert.False(t, internalT.Failed())
	})

	t.Run("injecting an outgoing frame works", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		expectedFrame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}
		m.InjectOutgoing(expectedFrame)

		actualFrame, err := Read(m)

		assert.NoError(t, err)
		assert.Equal(t, expectedFrame, actualFrame)
	})

	t.Run("mock records unexpected calls and asserts failure", func(t *testing.T) {
		m := NewMockAdapter()
		defer m.Stop()

		frame := Frame{MessageType: SREQ, Subsystem: ZDO, CommandID: 0xf0, Payload: []byte{0x02, 0x11}}

		err := Write(m, frame)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		assert.Equal(t, 1, len(m.UnexpectedCalls))
		assert.Equal(t, frame, m.UnexpectedCalls[0].Frame)

		internalT := new(testing.T)
		m.AssertCalls(internalT)
		assert.True(t, internalT.Failed())
	})
}
