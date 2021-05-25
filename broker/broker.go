package broker

import (
	"github.com/shimmeringbee/unpi"
	. "github.com/shimmeringbee/unpi/library"
	"io"
	"sync"
)

type FrameReader func(r io.Reader) (unpi.Frame, error)
type FrameWriter func(w io.Writer, frame unpi.Frame) error

type Broker struct {
	reader io.Reader
	writer io.Writer

	FrameReader FrameReader
	FrameWriter FrameWriter

	sendingChannel chan outgoingFrame
	sendingEnd     chan bool

	syncReceivingMutex *sync.Mutex
	receivingEnd       chan bool

	listenMutex          *sync.Mutex
	awaitMessageSequence *uint64
	listenRequests       map[listenRequest]ResponseFunction

	messageLibrary *Library
}

const PermittedQueuedRequests int = 50

func NewBroker(reader io.Reader, writer io.Writer, ml *Library) *Broker {
	z := &Broker{
		reader: reader,
		writer: writer,

		FrameReader: unpi.Read,
		FrameWriter: unpi.Write,

		sendingChannel: make(chan outgoingFrame, PermittedQueuedRequests),
		sendingEnd:     make(chan bool),

		receivingEnd: make(chan bool, 1),

		syncReceivingMutex: &sync.Mutex{},

		listenMutex:          &sync.Mutex{},
		awaitMessageSequence: new(uint64),
		listenRequests:       map[listenRequest]ResponseFunction{},

		messageLibrary: ml,
	}

	return z
}

func (b *Broker) Start() {
	go b.handleSending()
	go b.handleReceiving()
}

func (b *Broker) Stop() {
	b.sendingEnd <- true
	b.receivingEnd <- true
}
