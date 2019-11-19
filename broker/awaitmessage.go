package broker

import (
	"context"
	"errors"
	. "github.com/shimmeringbee/unpi"
	"log"
)

type awaitMessageRequest struct {
	MessageType MessageType
	SubSystem   Subsystem
	CommandID   byte
	Response    chan Frame
}

func (b *Broker) handleAwaitMessage(frame Frame) {
	b.awaitMessageMutex.Lock()
	defer b.awaitMessageMutex.Unlock()

	for req, _ := range b.awaitMessageRequests {
		if req.MessageType == frame.MessageType &&
			req.SubSystem == frame.Subsystem &&
			req.CommandID == frame.CommandID {

			select {
			case req.Response <- frame:
			default:
				log.Println("wait for matched, but no receivers in channel, probably timed out")
			}

			delete(b.awaitMessageRequests, req)
			close(req.Response)
		}
	}
}

func (b *Broker) addAwaitMessage(request awaitMessageRequest) {
	b.awaitMessageMutex.Lock()
	defer b.awaitMessageMutex.Unlock()

	b.awaitMessageRequests[request] = true
}

var AwaitMessageContextCancelled = errors.New("await message context cancelled")

func (b *Broker) AwaitMessage(ctx context.Context, messageType MessageType, subsystem Subsystem, commandID byte) (Frame, error) {
	wfr := awaitMessageRequest{
		MessageType: messageType,
		SubSystem:   subsystem,
		CommandID:   commandID,
		Response:    make(chan Frame),
	}

	b.addAwaitMessage(wfr)

	select {
	case frame := <-wfr.Response:
		return frame, nil
	case <-ctx.Done():
		return Frame{}, AwaitMessageContextCancelled
	}
}
