package broker

import (
	"context"
	"errors"
	"fmt"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
	. "github.com/shimmeringbee/unpi/library"
	"io"
	"log"
	"sync"
)

type Broker struct {
	reader io.Reader
	writer io.Writer

	requestsChannel chan OutgoingFrame
	requestsEnd     chan bool

	syncReceivingMutex    *sync.Mutex
	syncReceivingChannel  chan Frame
	asyncReceivingChannel chan Frame
	receivingEnd          chan bool

	waitForRequestsMutex *sync.Mutex
	waitForRequests      map[WaitFrameRequest]bool

	messageLibrary *Library
}

const PermittedQueuedRequests int = 50

type OutgoingFrame struct {
	Frame        Frame
	ErrorChannel chan error
}

type WaitFrameRequest struct {
	MessageType MessageType
	SubSystem   Subsystem
	CommandID   byte
	Response    chan Frame
}

func NewBroker(reader io.Reader, writer io.Writer, ml *Library) *Broker {
	z := &Broker{
		reader: reader,
		writer: writer,

		requestsChannel: make(chan OutgoingFrame, PermittedQueuedRequests),
		requestsEnd:     make(chan bool),

		syncReceivingMutex:    &sync.Mutex{},
		syncReceivingChannel:  make(chan Frame),
		asyncReceivingChannel: make(chan Frame, PermittedQueuedRequests),
		receivingEnd:          make(chan bool, 1),

		waitForRequestsMutex: &sync.Mutex{},
		waitForRequests:      map[WaitFrameRequest]bool{},

		messageLibrary: ml,
	}

	z.start()

	return z
}

func (b *Broker) start() {
	go b.handleRequests()
	go b.handleReceiving()
}

func (b *Broker) handleRequests() {
	for {
		select {
		case outgoing := <-b.requestsChannel:
			outgoing.ErrorChannel <- Write(b.writer, outgoing.Frame)
		case <-b.requestsEnd:
			return
		}
	}
}

func (b *Broker) handleReceiving() {
	for {
		frame, err := Read(b.reader)

		if err != nil {
			log.Printf("unpi read failed: %v\n", err)

			if errors.Is(err, io.EOF) {
				return
			}
		} else {
			if frame.MessageType != SRSP {
				b.serviceWaitForRequests(frame)
				b.asyncReceivingChannel <- frame
			} else {
				select {
				case b.syncReceivingChannel <- frame:
				default:
					log.Println("received synchronous response, but no receivers in channel")
				}
			}
		}

		select {
		case <-b.receivingEnd:
			return
		default:
		}
	}
}

func (b *Broker) serviceWaitForRequests(frame Frame) {
	b.waitForRequestsMutex.Lock()
	defer b.waitForRequestsMutex.Unlock()

	for req, _ := range b.waitForRequests {
		if req.MessageType == frame.MessageType &&
			req.SubSystem == frame.Subsystem &&
			req.CommandID == frame.CommandID {

			select {
			case req.Response <- frame:
			default:
				log.Println("wait for matched, but no receivers in channel, probably timed out")
			}

			delete(b.waitForRequests, req)
			close(req.Response)
		}
	}
}

func (b *Broker) addWaitForRequest(request WaitFrameRequest) {
	b.waitForRequestsMutex.Lock()
	defer b.waitForRequestsMutex.Unlock()

	b.waitForRequests[request] = true
}

func (b *Broker) Stop() {
	b.requestsEnd <- true
	b.receivingEnd <- true
}

func (b *Broker) writeFrame(frame Frame) error {
	errCh := make(chan error)

	b.requestsChannel <- OutgoingFrame{
		Frame:        frame,
		ErrorChannel: errCh,
	}

	return <-errCh
}

var FrameNotAsynchronous = errors.New("frame not asynchronous")
var FrameNotSynchronous = errors.New("frame not synchronous")

func (b *Broker) AsyncRequest(frame Frame) error {
	if frame.MessageType != AREQ {
		return FrameNotAsynchronous
	}

	return b.writeFrame(frame)
}

var WaitForFrameContextCancelled = errors.New("wait for frame context cancelled")

func (b *Broker) WaitForFrame(ctx context.Context, messageType MessageType, subsystem Subsystem, commandID byte) (Frame, error) {
	wfr := WaitFrameRequest{
		MessageType: messageType,
		SubSystem:   subsystem,
		CommandID:   commandID,
		Response:    make(chan Frame),
	}

	b.addWaitForRequest(wfr)

	select {
	case frame := <-wfr.Response:
		return frame, nil
	case <-ctx.Done():
		return Frame{}, WaitForFrameContextCancelled
	}
}

var SyncRequestContextCancelled = errors.New("synchronous request context cancelled")

func (b *Broker) SyncRequest(ctx context.Context, frame Frame) (Frame, error) {
	if frame.MessageType != SREQ {
		return Frame{}, FrameNotSynchronous
	}

	b.syncReceivingMutex.Lock()
	defer b.syncReceivingMutex.Unlock()

	if err := b.writeFrame(frame); err != nil {
		return Frame{}, err
	}

	select {
	case frame := <-b.syncReceivingChannel:
		return frame, nil
	case <-ctx.Done():
		return Frame{}, SyncRequestContextCancelled
	}
}

func (b *Broker) Receive() (Frame, error) {
	return <-b.asyncReceivingChannel, nil
}

func (b *Broker) MessageRequestResponse(ctx context.Context, req interface{}, resp interface{}) error {
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

		responseFrame, err = b.WaitForFrame(ctx, respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID)
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
