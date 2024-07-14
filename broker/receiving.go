package broker

import (
	"errors"
	"log"
	"syscall"
)

func (b *Broker) handleReceiving() {
	for {
		frame, err := b.FrameReader(b.reader)

		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}

			log.Printf("unpi read failed: %v\n", err)
			return
		} else {
			b.handleListeners(frame)
		}

		select {
		case <-b.receivingEnd:
			return
		default:
		}
	}
}
