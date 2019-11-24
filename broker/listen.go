package broker

import (
	. "github.com/shimmeringbee/unpi"
	"sync/atomic"
)

type ResponseFunction func(Frame)

type listenRequest struct {
	MessageType MessageType
	SubSystem   Subsystem
	CommandID   byte
	Sequence    uint64
}

func (b *Broker) handleListeners(frame Frame) {
	b.listenMutex.Lock()
	defer b.listenMutex.Unlock()

	for req, _ := range b.listenRequests {
		if req.MessageType == frame.MessageType &&
			req.SubSystem == frame.Subsystem &&
			req.CommandID == frame.CommandID {

			go b.listenRequests[req](frame)
		}
	}
}

func (b *Broker) addListen(request listenRequest, function ResponseFunction) {
	b.listenMutex.Lock()
	defer b.listenMutex.Unlock()

	b.listenRequests[request] = function
}

func (b *Broker) listen(messageType MessageType, subsystem Subsystem, commandID byte, function ResponseFunction) func() {
	awaitRequest := listenRequest{
		MessageType: messageType,
		SubSystem:   subsystem,
		CommandID:   commandID,
		Sequence:    atomic.AddUint64(b.awaitMessageSequence, 1),
	}

	b.addListen(awaitRequest, function)

	return func() {
		b.listenMutex.Lock()
		defer b.listenMutex.Unlock()
		delete(b.listenRequests, awaitRequest)
	}
}
