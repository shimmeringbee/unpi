package broker

import (
	. "github.com/shimmeringbee/unpi"
	"sync/atomic"
)

type ResponseFunction func(Frame)

type awaitMessageRequest struct {
	MessageType MessageType
	SubSystem   Subsystem
	CommandID   byte
	Sequence    uint64
}

func (b *Broker) handleAwaitMessage(frame Frame) {
	b.awaitMessageMutex.Lock()
	defer b.awaitMessageMutex.Unlock()

	for req, _ := range b.awaitMessageRequests {
		if req.MessageType == frame.MessageType &&
			req.SubSystem == frame.Subsystem &&
			req.CommandID == frame.CommandID {

			go b.awaitMessageRequests[req](frame)
			delete(b.awaitMessageRequests, req)
		}
	}
}

func (b *Broker) addAwaitMessage(request awaitMessageRequest, function ResponseFunction) {
	b.awaitMessageMutex.Lock()
	defer b.awaitMessageMutex.Unlock()

	b.awaitMessageRequests[request] = function
}

func (b *Broker) internalAwaitMessage(messageType MessageType, subsystem Subsystem, commandID byte, function ResponseFunction) func() {
	awaitRequest := awaitMessageRequest{
		MessageType: messageType,
		SubSystem:   subsystem,
		CommandID:   commandID,
		Sequence:    atomic.AddUint64(b.awaitMessageSequence, 1),
	}

	b.addAwaitMessage(awaitRequest, function)

	return func() {
		b.awaitMessageMutex.Lock()
		defer b.awaitMessageMutex.Unlock()
		delete(b.awaitMessageRequests, awaitRequest)
	}
}
