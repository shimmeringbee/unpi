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
	SREQ             = 0x01
	AREQ             = 0x02
	SRSP             = 0x03
)

type Subsystem byte

const (
	RES0        Subsystem = 0x00
	SYS                   = 0x01
	MAC                   = 0x02
	NWK                   = 0x03
	AF                    = 0x04
	ZDO                   = 0x05
	SAPI                  = 0x06
	UTIL                  = 0x07
	DBG                   = 0x08
	APP                   = 0x09
	RCAF                  = 0x0a
	RCN                   = 0x0b
	RCN_CLIENT            = 0x0c
	BOOT                  = 0x0d
	ZIPTEST               = 0x0e
	DEBUG                 = 0x0f
	PERIPHERALS           = 0x10
	NFC                   = 0x11
	PB_NWK_MGR            = 0x12
	PB_GW                 = 0x13
	PB_OTA_MGR            = 0x14
	BLE_SPNP              = 0x15
	BLE_HCI               = 0x16
	SRV_CTR               = 0x1f
)

type Frame struct {
	MessageType MessageType
	Subsystem   Subsystem
	CommandID   byte
	Payload     []byte
}

const StartOfFrame byte = 0xfe

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

func UnmarshallFrame(data []byte) (*Frame, error) {
	dataLength := len(data)

	if dataLength < MinimumFrameSize {
		return nil, FrameTooShort
	}

	if data[0] != StartOfFrame {
		return nil, FrameMissingStartOfFrame
	}

	payloadLength := int(data[1])

	if dataLength < MinimumFrameSize+payloadLength {
		return nil, FrameTooShort
	}

	checksum := calculateChecksum(data[1 : dataLength-1])

	if checksum != data[dataLength-1] {
		return nil, FrameChecksumFailed
	}

	messageType := MessageType(data[2] >> 5)
	subSystem := Subsystem(data[2] & 0x1f)

	frame := Frame{
		MessageType: messageType,
		Subsystem:   subSystem,
		CommandID:   data[3],
		Payload:     data[4 : dataLength-1],
	}

	return &frame, nil
}

func calculateChecksum(data []byte) (checksum byte) {
	for _, b := range data {
		checksum = checksum ^ b
	}

	return
}
