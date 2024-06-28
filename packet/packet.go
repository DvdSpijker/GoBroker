package packet

import (
	"fmt"

	"github.com/DvdSpijker/GoBroker/codec"
	"github.com/DvdSpijker/GoBroker/types"
)

type (
	PacketType         byte
	PacketFlag         byte
	PayloadExpectation string

	FixedHeader struct {
		PacketType      PacketType
		Flags           PacketFlag
		RemainingLength types.VariableByteInteger
	}

	VariableHeaderBase struct {
		PacketIdentifier types.UnsignedInt
		Properties       struct {
			Length    types.VariableByteInteger
			Identifer types.VariableByteInteger
			Property  any
		}
	}

	ControlPacket struct {
		FixedHeader    FixedHeader
		VariableHeader codec.Codec
		Payload        codec.Codec
	}

	PacketDefinition struct {
		PacketType         PacketType
		CanHaveProperties  bool
		PayloadExpectation PayloadExpectation
	}
)

const (
	Reserved PacketType = iota << 4
	CONNECT
	CONNACK
	PUBLISH
	PUBACK  // Publish Acknowledge Qos 1 Not implemented
	PUBREC  // Publish Received Qos 2 Not implemented
	PUBREL  // Publish Release Qos 2 Not implemented
	PUBCOMP // Publish Complete Qos 2 Not implemented
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH
)

const (
	CONNECTFLAGS     PacketFlag = 0b0000
	CONNACKFLAGS     PacketFlag = 0b0000
	PUBLISHFLAGS     PacketFlag = 0b0000
	PUBACKFLAGS      PacketFlag = 0b0000
	PUBRECFLAGS      PacketFlag = 0b0000
	PUBRELFLAGS      PacketFlag = 0b0010
	PUBCOMPFLAGS     PacketFlag = 0b0000
	SUBSCRIBEFLAGS   PacketFlag = 0b0010
	SUBACKFLAGS      PacketFlag = 0b0000
	UNSUBSCRIBEFLAGS PacketFlag = 0b0010
	UNSUBACKFLAGS    PacketFlag = 0b0000
	PINGREQFLAGS     PacketFlag = 0b0000
	PINGRESPFLAGS    PacketFlag = 0b0000
	DISCONNECTFLAGS  PacketFlag = 0b0000
	AUTHFLAGS        PacketFlag = 0b0000
)

const (
	NONE     PayloadExpectation = "None"
	OPTIONAL PayloadExpectation = "Optional"
	REQUIRED PayloadExpectation = "Required"
)

func (fixedHeader *FixedHeader) Encode() ([]byte, error) {
  encoded := make([]byte, 1)
	encoded[0] = byte(fixedHeader.PacketType)
	encoded[0] |= byte(fixedHeader.Flags)

  if fixedHeader.RemainingLength.Value > 0 {
    b, err := fixedHeader.RemainingLength.Encode()
    if err != nil {
      return nil, err
    }
    return append(encoded, b...), nil
  }

  return encoded, nil
}

func (fixedHeader *FixedHeader) Decode(input []byte) (int, error) {
	if len(input) < 1 {
		return 0, codec.DecodeErr(fixedHeader, "input length < 1")
	}
	fixedHeader.PacketType = PacketType(input[0] & 0xF0)
	fixedHeader.Flags = PacketFlag(input[0] & 0x0F)
	n, err := fixedHeader.RemainingLength.Decode(input[1:])
	if err != nil {
		return 0, err
	}
	return n + 1, nil
}

func (fixedHeader *FixedHeader) String() string {
	return fmt.Sprintf("packet type: %d | flags: %x | rem. length: %d",
		fixedHeader.PacketType,
		fixedHeader.Flags,
		fixedHeader.RemainingLength.Value)
}

func (packet *ControlPacket) Encode() ([]byte, error) {
	return nil, nil
}

func (packet *ControlPacket) Decode(input []byte) (int, error) {
	n, err := packet.FixedHeader.Decode(input)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (vhb VariableHeaderBase) Encode() ([]byte, error) {
	// vhb.

	return nil, nil
}
