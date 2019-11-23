package broker

import (
	"errors"
	"github.com/shimmeringbee/bytecodec"
	. "github.com/shimmeringbee/unpi"
)

func (b *Broker) Request(req interface{}) error {
	reqIdentity, reqFound := b.messageLibrary.GetByObject(req)

	if !reqFound {
		return errors.New("request message was not in message library")
	}

	if reqIdentity.MessageType == SREQ {
		return errors.New("synchronous messages cannot be sent one shot")
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

	return b.writeFrame(requestFrame)
}
