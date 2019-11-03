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
	return nil, nil
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
