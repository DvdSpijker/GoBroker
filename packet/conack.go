package packet

import (
	"github.com/DvdSpijker/GoBroker/types"
)

type (
	ConackVariableHeader struct {
		VariableHeaderBase      VariableHeaderBase
		ConnectAcknowledgeFlags byte
		ConnectReasonCode       ConnectReasonCode
	}

	ConackPacket struct {
		FixedHeader    FixedHeader
		VariableHeader ConackVariableHeader
	}

	ConnectReasonCode      byte
	ConnectAcknowledgeFlag byte
)

const (
	SessionPresent ConnectAcknowledgeFlag = 0b10000000
)

const (
	Success                     ConnectReasonCode = 0x00
	UnspecifiedError            ConnectReasonCode = 0x80
	MalformedPacket             ConnectReasonCode = 0x81
	ProtocolError               ConnectReasonCode = 0x82
	ImplementationSpecificError ConnectReasonCode = 0x83
	UnsupportedProtocolVersion  ConnectReasonCode = 0x84
	ClientIdentifierNotValid    ConnectReasonCode = 0x85
	BadUserNameOrPassword       ConnectReasonCode = 0x86
	NotAuthenterized            ConnectReasonCode = 0x87
	ServerUnavailable           ConnectReasonCode = 0x88
	ServerBusy                  ConnectReasonCode = 0x89
	Banned                      ConnectReasonCode = 0x8A
	BadAuthenticationMethod     ConnectReasonCode = 0x8C
	TopicNameInvalid            ConnectReasonCode = 0x90
	PacketTooLarge              ConnectReasonCode = 0x95 // (That's what she said)
	QuotaExceeded               ConnectReasonCode = 0x97
	PayloadFormatInvalid        ConnectReasonCode = 0x99
	RetainNotSupported          ConnectReasonCode = 0x9A
	QosNotSupported             ConnectReasonCode = 0x9B
	UseAnotherServer            ConnectReasonCode = 0x9C
	ServerMoved                 ConnectReasonCode = 0x9D
	ConnectionRateExceeded      ConnectReasonCode = 0x9F
)

func (packet ConackPacket) Encode() (bin []byte, err error) {
	packet.FixedHeader.PacketType = CONNACK

	variabledHdrBin, err := packet.VariableHeader.Encode()
	if err != nil {
		return nil, err
	}

	fixedHdrBin, err := packet.FixedHeader.Encode()
	if err != nil {
		return nil, err
	}

	remainingLen := types.VariableByteInteger{
		Value: int32(len(variabledHdrBin)),
	}
	remainingLenBin, err := remainingLen.Encode()
	if err != nil {
		return nil, err
	}

	bin = append(bin, fixedHdrBin...)
	bin = append(bin, remainingLenBin...)
	bin = append(bin, variabledHdrBin...)

	return bin, nil
}

func (hdr ConackVariableHeader) Encode() (bin []byte, err error) {
	// bin, err := hdr.VariableHeaderBase.Encode()
	// if err != nil {
	// 	return nil, err
	// }
	bin = append(bin, hdr.ConnectAcknowledgeFlags)
	bin = append(bin, byte(hdr.ConnectReasonCode))
	// Property length of zero,
	// because there are no properties
	bin = append(bin, 0)
	return bin, nil
}
