package broker

import (
	"context"
	"errors"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
	"log"
)

var ContextCancelled = errors.New("context cancelled")
var RequestMessageNotInLibrary = errors.New("request message was not in message library")
var ResponseMessageNotInLibrary = errors.New("response message was not in message library")

func (b *Broker) RequestResponse(ctx context.Context, req interface{}, resp interface{}) error {
	reqIdentity, reqFound := b.messageLibrary.GetByObject(req)
	respIdentity, respFound := b.messageLibrary.GetByObject(resp)

	if !reqFound {
		return RequestMessageNotInLibrary
	}

	if !respFound {
		return ResponseMessageNotInLibrary
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

	if reqIdentity.MessageType == SREQ {
		b.syncReceivingMutex.Lock()
		defer b.syncReceivingMutex.Unlock()
	}

	ch := make(chan Frame, 1)
	defer close(ch)

	cancelAwait := b.listen(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
		ch <- f
	})
	defer cancelAwait()

	if err := b.writeFrame(requestFrame); err != nil {
		return err
	}

	var f Frame

	select {
	case f = <-ch:
	case <-ctx.Done():
		return ContextCancelled
	}

	err = bytecodec.Unmarshall(f.Payload, resp)

	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Await(ctx context.Context, resp interface{}) error {
	respIdentity, respFound := b.messageLibrary.GetByObject(resp)

	if !respFound {
		return ResponseMessageNotInLibrary
	}

	ch := make(chan Frame, 1)
	defer close(ch)

	cancelAwait := b.listen(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
		ch <- f
	})
	defer cancelAwait()

	var f Frame

	select {
	case f = <-ch:
	case <-ctx.Done():
		return ContextCancelled
	}

	err := bytecodec.Unmarshall(f.Payload, resp)

	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Subscribe(message interface{}, callback func(v interface{})) (error, func()) {
	msgIdentity, msgFound := b.messageLibrary.GetByObject(message)

	if !msgFound {
		return ResponseMessageNotInLibrary, func() {}
	}

	ch := make(chan Frame)
	done := make(chan bool)

	go func() {
		for {
			select {
			case f := <- ch:
				err := bytecodec.Unmarshall(f.Payload, message)

				if err != nil {
					log.Printf("failed to unmarshal: %+v", err)
				} else {
					callback(message)
				}
			case <- done:
				return
			}
		}
	}()

	cancelAwait := b.listen(msgIdentity.MessageType, msgIdentity.Subsystem, msgIdentity.CommandID, func(f Frame) {
		ch <- f
	})

	return nil, func() {
		cancelAwait()
		done <- true
		close(done)
		close(ch)
	}
}
