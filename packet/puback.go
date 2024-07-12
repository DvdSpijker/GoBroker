package packet

import "github.com/DvdSpijker/GoBroker/types"

type (
  PubackVariableHeader struct {
    PacketIdentifer types.UnsignedInt
    ReasonCode ReasonCode
    PropertyLength types.VariableByteInteger
  }
  PubackPacket struct {
    FixedHeader FixedHeader
    VariableHeader PubackVariableHeader
  }
)

func (packet *PubackPacket) Encode() ([]byte, error) {
  bytes := []byte{}

  b, err := packet.VariableHeader.PacketIdentifer.Encode()
  if err != nil {
    return nil, err
  }

  bytes = append(bytes, b...)

  if packet.VariableHeader.ReasonCode != Success &&
    packet.VariableHeader.PropertyLength.Value > 0 {
    bytes = append(bytes, byte(packet.VariableHeader.ReasonCode))

    b, err = packet.VariableHeader.PropertyLength.Encode()
    if err != nil {
      return nil, err
    }

    bytes = append(bytes, b...)
  }

  packet.FixedHeader.PacketType = PUBACK
  packet.FixedHeader.Flags = 0
  packet.FixedHeader.RemainingLength.Value = int32(len(bytes))

  b, err = packet.FixedHeader.Encode()
  if err != nil {
    return nil, err
  }

  return append(b, bytes...), nil
}
