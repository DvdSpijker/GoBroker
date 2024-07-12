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

func (packet *PubackPacket) Decode(input []byte) (int, error) {
  totalRead := 0

  n, err := packet.FixedHeader.Decode(input)
  if err != nil {
    return 0, err
  }

  totalRead += n
  input = input[n:]

  if len(input) > 0 {
    packet.VariableHeader.PacketIdentifer.Size = 2
  n, err = packet.VariableHeader.PacketIdentifer.Decode(input)
  if err != nil {
    return 0, err
  }
  totalRead += n
  input = input[n:]
  }


  if len(input) > 0 {
    packet.VariableHeader.ReasonCode = ReasonCode(input[0])
    totalRead += 1
    input = input[1:]
  }

  if len(input) > 0 {
    n, err := packet.VariableHeader.PropertyLength.Decode(input)
    if err != nil {
      return 0, err
    }
    // TODO: Read the properties.
    totalRead += n + int(packet.VariableHeader.PropertyLength.Value)
    input = input[n:]
  }

  return totalRead, nil
}
