package broker

import . "github.com/shimmeringbee/unpi"

type outgoingFrame struct {
	Frame        Frame
	ErrorChannel chan error
}

func (b *Broker) handleSending() {
	for {
		select {
		case outgoing := <-b.sendingChannel:
			outgoing.ErrorChannel <- b.FrameWriter(b.writer, outgoing.Frame)
		case <-b.sendingEnd:
			return
		}
	}
}

func (b *Broker) writeFrame(frame Frame) error {
	errCh := make(chan error)

	b.sendingChannel <- outgoingFrame{
		Frame:        frame,
		ErrorChannel: errCh,
	}

	return <-errCh
}
