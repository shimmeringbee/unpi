// Package unpi provides marshalling and unmarshalling support for Texas Instruments
// Unified Network Processor Interface.
//
// More information available at:
// http://processors.wiki.ti.com/index.php/Unified_Network_Processor_Interface

package unpi // import "github.com/shimmeringbee/unipi"

import "io"

type UNPI struct {
	device io.ReadWriter
}

func New(device io.ReadWriter) *UNPI {
	return nil
}

func (u *UNPI) Read() (*Frame, error) {
	return nil, nil
}

func (u *UNPI) Write(frame *Frame) error {
	return nil
}
