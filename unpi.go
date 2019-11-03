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

// Read reads from the ReadWriter provided to the UNPI struct until it receives
// a whole UNPI frame. It returns a pointer to the frame, or an error if one
// was encountered. Error may either be an issue with the structure of the
// frame or an error raised by the ReadWriter.
func Read(r io.Reader) (*Frame, error) {
	data := []byte{StartOfFrame, 0x00, 0x00, 0x00, 0x00}

	if err := seekStartOfFrame(r); err != nil {
		return nil, err
	}

	_, err := io.ReadFull(r, data[1:])
	if err != nil {
		return nil, err
	}

	payloadLength := data[1]

	data = append(data, make([]byte, payloadLength)...)

	c, err := io.ReadFull(r, data[5:])
	if err != nil && c != int(payloadLength) {
		return nil, err
	}

	return UnmarshallFrame(data)
}

func seekStartOfFrame(r io.Reader) error {
	b := []byte{0x00}

	for b[0] != StartOfFrame {
		_, err := r.Read(b)

		if err != nil {
			return err
		}
	}

	return nil
}

// Write marshalls and writes a UNPI frame to the ReadWriter provided to the UNPI struct.
// It will return an error if one was encountered while writing to the ReadWriter
func Write(w io.Writer, frame *Frame) error {
	data := frame.Marshall()

	dataSize := len(data)
	dataWritten, err := w.Write(data)

	if err != nil {
		return err
	}

	if dataWritten != len(data) {
		return fmt.Errorf("writer did not accept whole frame, sent %d, written %d", dataSize, dataWritten)
	}

	return nil
}
