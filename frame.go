package unpi

import (
	"bytes"
	"errors"
)

// Constants extracted from:
// - http://processors.wiki.ti.com/index.php/NPI_Type_SubSystem
// - http://processors.wiki.ti.com/index.php/Unified_Network_Processor_Interface
// - http://dev.ti.com/tirex/content/simplelink_cc13x2_sdk_2_30_00_45/docs/zstack/html/zigbee/znp_interface.html

type MessageType byte

const (
	POLL MessageType = 0x00
	SREQ MessageType = 0x01
	AREQ MessageType = 0x02
	SRSP MessageType = 0x03
)

type Subsystem byte

const (
	RES0        Subsystem = 0x00
	SYS         Subsystem = 0x01
	MAC         Subsystem = 0x02
	NWK         Subsystem = 0x03
	AF          Subsystem = 0x04
	ZDO         Subsystem = 0x05
	SAPI        Subsystem = 0x06
	UTIL        Subsystem = 0x07
	DBG         Subsystem = 0x08
	APP         Subsystem = 0x09
	RCAF        Subsystem = 0x0a
	RCN         Subsystem = 0x0b
	RCN_CLIENT  Subsystem = 0x0c
	BOOT        Subsystem = 0x0d
	ZIPTEST     Subsystem = 0x0e
	APP_CNF     Subsystem = 0x0f
	DEBUG       Subsystem = 0x0f
	PERIPHERALS Subsystem = 0x10
	NFC         Subsystem = 0x11
	PB_NWK_MGR  Subsystem = 0x12
	PB_GW       Subsystem = 0x13
	PB_OTA_MGR  Subsystem = 0x14
	BLE_SPNP    Subsystem = 0x15
	BLE_HCI     Subsystem = 0x16
	SRV_CTR     Subsystem = 0x1f
)

type Frame struct {
	MessageType MessageType
	Subsystem   Subsystem
	CommandID   byte
	Payload     []byte
}

const StartOfFrame byte = 0xfe

// Marshall constructs a byte array representing the Frame it was called upon, ready for writing directly to the wire,
// this includes the Start Of Frame header and Checksum.
func (f *Frame) Marshall() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte(StartOfFrame)

	payloadLength := len(f.Payload)
	buffer.WriteByte(byte(payloadLength))

	typeSystem := byte(f.MessageType<<5) | byte(f.Subsystem)
	buffer.WriteByte(typeSystem)

	buffer.WriteByte(f.CommandID)
	buffer.Write(f.Payload)

	checksum := calculateChecksum(buffer.Bytes()[1:])
	buffer.WriteByte(checksum)

	return buffer.Bytes()
}

var FrameChecksumFailed = errors.New("frame failed checksum")
var FrameTooShort = errors.New("frame too short")
var FrameMissingStartOfFrame = errors.New("frame is missing start of frame")

const MinimumFrameSize int = 5

// Unmarshall converts a byte array into a Frame, providing it is valid. Byte array must include Start Of Frame and
// a checksum, as it would be on the wire.
// It may return an error if the provided byte array does not correctly represent a frame.
func UnmarshallFrame(data []byte) (Frame, error) {
	dataLength := len(data)

	if dataLength < MinimumFrameSize {
		return Frame{}, FrameTooShort
	}

	if data[0] != StartOfFrame {
		return Frame{}, FrameMissingStartOfFrame
	}

	payloadLength := int(data[1])

	if dataLength < MinimumFrameSize+payloadLength {
		return Frame{}, FrameTooShort
	}

	checksum := calculateChecksum(data[1 : dataLength-1])

	if checksum != data[dataLength-1] {
		return Frame{}, FrameChecksumFailed
	}

	messageType := MessageType(data[2] >> 5)
	subSystem := Subsystem(data[2] & 0x1f)

	frame := Frame{
		MessageType: messageType,
		Subsystem:   subSystem,
		CommandID:   data[3],
		Payload:     data[4 : dataLength-1],
	}

	return frame, nil
}

func calculateChecksum(data []byte) (checksum byte) {
	for _, b := range data {
		checksum = checksum ^ b
	}

	return
}
