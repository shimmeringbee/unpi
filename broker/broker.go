package broker

import (
	. "github.com/shimmeringbee/unpi/library"
	"io"
	"sync"
)

type Broker struct {
	reader io.Reader
	writer io.Writer

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

		sendingChannel: make(chan outgoingFrame, PermittedQueuedRequests),
		sendingEnd:     make(chan bool),

		receivingEnd: make(chan bool, 1),

		syncReceivingMutex: &sync.Mutex{},

		listenMutex:          &sync.Mutex{},
		awaitMessageSequence: new(uint64),
		listenRequests:       map[listenRequest]ResponseFunction{},

		messageLibrary: ml,
	}

	z.start()

	return z
}

func (b *Broker) start() {
	go b.handleSending()
	go b.handleReceiving()
}

func (b *Broker) Stop() {
	b.sendingEnd <- true
	b.receivingEnd <- true
}
