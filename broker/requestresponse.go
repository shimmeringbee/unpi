package broker

import (
	"context"
	"errors"
	"fmt"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
	"log"
	"reflect"
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

	requestPayload, err := bytecodec.Marshal(req)

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

	oneShot := false

	cancelAwait := b.listen(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
		if !oneShot {
			oneShot = true
			select {
			case ch <- f:
			default:
			}
		}
	})

	defer func() {
		cancelAwait()
		close(ch)
	}()

	if err := b.writeFrame(requestFrame); err != nil {
		return err
	}

	var f Frame

	select {
	case f = <-ch:
	case <-ctx.Done():
		return ContextCancelled
	}

	err = bytecodec.Unmarshal(f.Payload, resp)

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

	cancelAwait := b.listen(respIdentity.MessageType, respIdentity.Subsystem, respIdentity.CommandID, func(f Frame) {
		select {
		case ch <- f:
		default:
		}
	})

	defer func() {
		cancelAwait()
		close(ch)
	}()

	var f Frame

	select {
	case f = <-ch:
	case <-ctx.Done():
		return ContextCancelled
	}

	err := bytecodec.Unmarshal(f.Payload, resp)

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

	cancelAwait := b.listen(msgIdentity.MessageType, msgIdentity.Subsystem, msgIdentity.CommandID, func(f Frame) {
		copiedMessage, err := copyInterface(message)

		if err != nil {
			log.Printf("could not copy message for a callback: %+v", err)
		} else {
			err := bytecodec.Unmarshal(f.Payload, copiedMessage)

			if err != nil {
				log.Printf("failed to unmarshal: %+v", err)
			} else {
				callback(copiedMessage)
			}
		}
	})

	return nil, cancelAwait
}

func copyInterface(source interface{}) (interface{}, error) {
	v := reflect.ValueOf(source)

	switch v.Kind() {
	case reflect.Struct:
		sourceValue := reflect.New(v.Type()).Elem()
		return sourceValue.Interface(), nil
	case reflect.Ptr:
		e := v.Elem()
		sourceValue := reflect.New(e.Type())
		return sourceValue.Interface(), nil
	default:
		return nil, fmt.Errorf("unable to copy interface: %+v", v.Kind())
	}
}
