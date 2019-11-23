package broker

import (
	"context"
	"errors"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
)

var RequestResponseContextCancelled = errors.New("request response context cancelled")

func (b *Broker) RequestResponse(ctx context.Context, req interface{}, resp interface{}) error {
	reqIdentity, reqFound := b.messageLibrary.GetByObject(req)
	respIdentity, respFound := b.messageLibrary.GetByObject(resp)

	if !reqFound {
		return errors.New("request message was not in message library")
	}

	if !respFound {
		return errors.New("response message was not in message library")
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

	cancelAwait := b.awaitMessage(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
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
		return RequestResponseContextCancelled
	}

	err = bytecodec.Unmarshall(f.Payload, resp)

	if err != nil {
		return err
	}

	return nil
}
