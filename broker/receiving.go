package broker

import (
	"errors"
	. "github.com/shimmeringbee/unpi"
	"io"
	"log"
)

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
				b.handleAwaitMessage(frame)
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

func (b *Broker) Receive() (Frame, error) {
	return <-b.asyncReceivingChannel, nil
}
