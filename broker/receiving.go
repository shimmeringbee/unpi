package broker

import (
	"errors"
	. "github.com/shimmeringbee/unpi"
	"io"
	"log"
	"syscall"
)

func (b *Broker) handleReceiving() {
	for {
		frame, err := Read(b.reader)

		if err != nil {
			switch e := err.(type) {
			case syscall.Errno:
				if e == syscall.EINTR {
					continue
				}
			}

			log.Printf("unpi read failed: %v\n", err)

			if errors.Is(err, io.EOF) {
				return
			}
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
