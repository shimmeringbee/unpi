// Package unpi provides marshalling and unmarshalling support for Texas Instruments
// Unified Network Processor Interface.
//
// More information available at:
// http://processors.wiki.ti.com/index.php/Unified_Network_Processor_Interface
package unpi // import "github.com/shimmeringbee/unipi"

import (
	"fmt"
	"io"
)

type UNPI struct {
	device io.ReadWriter
}

func New(device io.ReadWriter) *UNPI {
	u := &UNPI{device: device}

	return u
}

func (u *UNPI) Read() (*Frame, error) {
	data := []byte{StartOfFrame, 0x00, 0x00, 0x00, 0x00}

	if err := u.seekStartOfFrame(); err != nil {
		return nil, err
	}

	_, err := io.ReadFull(u.device, data[1:])
	if err != nil {
		return nil, err
	}

	payloadLength := data[1]

	data = append(data, make([]byte, payloadLength)...)

	c, err := io.ReadFull(u.device, data[5:])
	if err != nil && c != int(payloadLength) {
		return nil, err
	}

	return UnmarshallFrame(data)
}

func (u *UNPI) seekStartOfFrame() error {
	b := []byte{0x00}

	for b[0] != StartOfFrame {
		_, err := u.device.Read(b)

		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UNPI) Write(frame *Frame) error {
	data := frame.Marshall()

	dataSize := len(data)
	dataWritten, err := u.device.Write(data)

	if err != nil {
		return err
	}

	if dataWritten != len(data) {
		return fmt.Errorf("writer did not accept whole frame, sent %d, written %d", dataSize, dataWritten)
	}

	return nil
}
