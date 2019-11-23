package testing

import (
	"bytes"
	. "github.com/shimmeringbee/unpi"
	"io"
	"math"
	"sync/atomic"
	"testing"
)

type MockAdapter struct {
	sequencer *int64

	ReceivedFrames []Frame

	incomingReader io.Reader
	incomingWriter io.WriteCloser
	incomingEnd    chan bool

	outgoingBuffer *bytes.Buffer
	outgoingFrames chan Frame
	outgoingEnd    chan bool

	Calls []*Call
}

type CallRecord struct {
	when  int64
	Frame Frame
}

func (cr CallRecord) Before(ocr CallRecord) bool {
	return cr.when < ocr.when
}

func (cr CallRecord) AssertBefore(t *testing.T, ocr CallRecord) {
	if !cr.Before(ocr) {
		t.Logf("assertion failed, call happened after")
		t.Fail()
	}
}

func (cr CallRecord) After(ocr CallRecord) bool {
	return cr.when > ocr.when
}

func (cr CallRecord) AssertAfter(t *testing.T, ocr CallRecord) {
	if !cr.After(ocr) {
		t.Logf("assertion failed, call happened before")
		t.Fail()
	}
}

type Call struct {
	mT MessageType
	s  Subsystem
	c  byte

	CapturedCalls []CallRecord
	returnFrames  []Frame

	expectedCalls int
	actualCalls   int
}

func (c *Call) Return(frames ...Frame) *Call {
	c.returnFrames = frames
	return c
}

func (c *Call) Times(times int) *Call {
	c.expectedCalls = times
	return c
}

func (c *Call) UnlimitedTimes() *Call {
	c.expectedCalls = math.MinInt64
	return c
}

func (c *Call) Frames() []Frame {
	return []Frame{}
}

func NewMockAdapter() *MockAdapter {
	m := &MockAdapter{
		sequencer:      new(int64),
		ReceivedFrames: []Frame{},

		incomingEnd: make(chan bool, 1),

		Calls: []*Call{},

		outgoingFrames: make(chan Frame, 50),
		outgoingEnd:    make(chan bool, 1),
	}

	*m.sequencer = 0
	m.incomingReader, m.incomingWriter = io.Pipe()

	m.start()

	return m
}

func (m *MockAdapter) Read(p []byte) (n int, err error) {
	if m.outgoingBuffer == nil {
		select {
		case f := <-m.outgoingFrames:
			data := f.Marshall()
			m.outgoingBuffer = bytes.NewBuffer(data)
		case <-m.outgoingEnd:
			return 0, io.EOF
		}
	}

	actualRead, err := m.outgoingBuffer.Read(p)

	if m.outgoingBuffer.Len() == 0 {
		m.outgoingBuffer = nil
	}

	return actualRead, err
}

func (m *MockAdapter) Write(p []byte) (n int, err error) {
	return m.incomingWriter.Write(p)
}

const (
	AnyType      MessageType = 0xff
	AnySubsystem Subsystem   = 0xff
	AnyCommand   uint8       = 0xff
)

func (m *MockAdapter) On(mT MessageType, s Subsystem, c uint8) *Call {
	call := &Call{
		mT:            mT,
		s:             s,
		c:             c,
		CapturedCalls: []CallRecord{},
		returnFrames:  []Frame{},
		expectedCalls: 1,
		actualCalls:   0,
	}

	m.Calls = append(m.Calls, call)

	return call
}

func (m *MockAdapter) AssertCalls(t *testing.T) {
	for _, call := range m.Calls {
		if call.expectedCalls != call.actualCalls && call.expectedCalls != math.MinInt64 {
			t.Logf("call count mismatch (mT: %v s: %v c: %v): expected(%d) != actual(%d)", call.mT, call.s, call.c, call.expectedCalls, call.actualCalls)
			t.Fail()
		}
	}
}

func (m *MockAdapter) handleIncoming() {
	for {
		frame, err := Read(m.incomingReader)
		if err != nil {
			return
		}

		m.ReceivedFrames = append(m.ReceivedFrames, frame)
		m.matchCalls(frame)

		select {
		case <-m.incomingEnd:
			return
		default:
		}
	}
}

func (m *MockAdapter) matchCalls(frame Frame) {
	for _, call := range m.Calls {
		if (call.mT == frame.MessageType || call.mT == AnyType) &&
			(call.s == frame.Subsystem || call.s == AnySubsystem) &&
			(call.c == frame.CommandID || call.c == AnyCommand) {

			if len(call.returnFrames) > 0 {
				i := call.actualCalls % len(call.returnFrames)
				m.outgoingFrames <- call.returnFrames[i]
			}

			call.CapturedCalls = append(call.CapturedCalls, CallRecord{
				when:  atomic.AddInt64(m.sequencer, 1),
				Frame: frame,
			})

			call.actualCalls += 1
		}
	}
}

func (m *MockAdapter) start() {
	go m.handleIncoming()
}

func (m *MockAdapter) Stop() {
	_ = m.incomingWriter.Close()
	m.incomingEnd <- true
	m.outgoingEnd <- true
}
