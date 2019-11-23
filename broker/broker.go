package broker

import (
	"context"
	"errors"
	"fmt"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
	. "github.com/shimmeringbee/unpi/library"
	"io"
	"sync"
)

type Broker struct {
	reader io.Reader
	writer io.Writer

	sendingChannel chan outgoingFrame
	sendingEnd     chan bool

	syncReceivingMutex    *sync.Mutex
	syncReceivingChannel  chan Frame
	asyncReceivingChannel chan Frame
	receivingEnd          chan bool

	awaitMessageMutex    *sync.Mutex
	awaitMessageRequests map[awaitMessageRequest]bool

	messageLibrary *Library
}

const PermittedQueuedRequests int = 50

func NewBroker(reader io.Reader, writer io.Writer, ml *Library) *Broker {
	z := &Broker{
		reader: reader,
		writer: writer,

		sendingChannel: make(chan outgoingFrame, PermittedQueuedRequests),
		sendingEnd:     make(chan bool),

		syncReceivingMutex:    &sync.Mutex{},
		syncReceivingChannel:  make(chan Frame),
		asyncReceivingChannel: make(chan Frame, PermittedQueuedRequests),
		receivingEnd:          make(chan bool, 1),

		awaitMessageMutex:    &sync.Mutex{},
		awaitMessageRequests: map[awaitMessageRequest]bool{},

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

var FrameNotAsynchronous = errors.New("frame not asynchronous")
var FrameNotSynchronous = errors.New("frame not synchronous")

func (b *Broker) AsyncRequest(frame Frame) error {
	if frame.MessageType != AREQ {
		return FrameNotAsynchronous
	}

	return b.writeFrame(frame)
}

var SyncRequestContextCancelled = errors.New("synchronous request context cancelled")

func (b *Broker) SyncRequest(ctx context.Context, frame Frame) (Frame, error) {
	if frame.MessageType != SREQ {
		return Frame{}, FrameNotSynchronous
	}

	b.syncReceivingMutex.Lock()
	defer b.syncReceivingMutex.Unlock()

	type FrameOrError struct {
		Frame Frame
		Error error
	}

	ch := make(chan FrameOrError)
	defer close(ch)

	go func() {
		select {
		case frame := <-b.syncReceivingChannel:
			ch <- FrameOrError{Frame: frame}
		case <-ctx.Done():
			ch <- FrameOrError{Error: SyncRequestContextCancelled}
		default:
		}
	}()

	if err := b.writeFrame(frame); err != nil {
		return Frame{}, err
	}

	select {
	case foe := <-ch:
		return foe.Frame, foe.Error
	}
}

func (b *Broker) RequestResponse(ctx context.Context, req interface{}, resp interface{}) error {
	reqIdentity, reqFound := b.messageLibrary.GetByObject(req)
	respIdentity, respFound := b.messageLibrary.GetByObject(resp)

	if !reqFound || !respFound {
		return errors.New("message has not been recognised")
	}

	requestPayload, err := bytecodec.Marshall(req)

	if err != nil {
		return err
	}

	requestFrame := Frame{
		MessageType: reqIdentity.MessageType,
		Subsystem:   reqIdentity.Subsystem,
		CommandID:   reqIdentity.CommandID,
		Payload:     requestPayload,
	}

	responseFrame := Frame{}

	if reqIdentity.MessageType == SREQ {
		responseFrame, err = b.SyncRequest(ctx, requestFrame)
	} else {
		if err := b.AsyncRequest(requestFrame); err != nil {
			return err
		}

		responseFrame, err = b.AwaitMessage(ctx, respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID)
	}

	if err != nil {
		return fmt.Errorf("bad sync request: %+v", err)
	}

	err = bytecodec.Unmarshall(responseFrame.Payload, resp)

	if err != nil {
		return nil
	}

	return nil
}
