package packet

import (
	"github.com/DvdSpijker/GoBroker/types"
)

type (
	ConackVariableHeader struct {
		VariableHeaderBase      VariableHeaderBase
		ConnectAcknowledgeFlags byte
		ConnectReasonCode      ReasonCode
	}

	ConackPacket struct {
		FixedHeader    FixedHeader
		VariableHeader ConackVariableHeader
	}

	ConnectAcknowledgeFlag byte
)

const (
	SessionPresent ConnectAcknowledgeFlag = 0b10000000
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
