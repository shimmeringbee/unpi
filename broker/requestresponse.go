package broker

import (
	"context"
	"errors"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
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

	cancelAwait := b.internalAwaitMessage(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
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

	cancelAwait := b.internalAwaitMessage(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
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
