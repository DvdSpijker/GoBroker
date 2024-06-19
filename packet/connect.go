package packet

import (
  "fmt"
  "github.com/DvdSpijker/GoBroker/types"
)

type (
  ConnectVariableHeader struct {
    VariableHeaderBase VariableHeaderBase
  }

  ConnectPayload struct {
    ClientId types.UtfString
  }

  ConnectPacket struct {
    FixedHeader FixedHeader
    VariableHeader ConnectVariableHeader
    Payload ConnectPayload
  }
)


func (header *ConnectVariableHeader) Decode(input []byte) (int, error) {
  return 16, nil
}

func (payload *ConnectPayload) Decode(input []byte) (int, error) {
  n, err := payload.ClientId.Decode(input)
  if err != nil {
    return 0, err
  }

  return n, nil
}

func (payload *ConnectPayload) String() string {
  return payload.ClientId.String()
}

func (packet *ConnectPacket) Decode(input []byte) (int, error) {
  n, err := packet.FixedHeader.Decode(input)
  if err != nil {
    return 0, err
  }
  fixedHeaderSize := n
  fmt.Println("fixed header", packet.FixedHeader)

  input = input[n:]
  n, err = packet.VariableHeader.Decode(input)
  if err != nil {
    return 0, err
  }

  input = input[n:]
  n, err = packet.Payload.Decode(input)
  if err != nil {
    return 0, err
  }
  fmt.Println("payload", packet.Payload)

  return fixedHeaderSize + int(packet.FixedHeader.RemainingLength.Value), nil
}
