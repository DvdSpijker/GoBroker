package packet

import (
	"errors"
	"fmt"

	"github.com/DvdSpijker/GoBroker/types"
)

type (


	ConnectPacket struct {
		FixedHeader    FixedHeader
    VariableHeader struct {
      ProtocolName types.UtfString
      Version byte

      UserNameFlag bool
      PasswordFlag bool
      WillRetain bool
      WillQos types.QoS
      WillFlag bool
      CleanStart bool

      KeepAlive types.UnsignedInt // In seconds
      PropertyLength         types.VariableByteInteger
    }
    Payload struct {
      ClientId types.UtfString
    }
	}
)

func (packet *ConnectPacket) Decode(input []byte) (int, error) {
  totalRead := 0

	n, err := packet.FixedHeader.Decode(input)
	if err != nil {
		return 0, err
	}

	totalRead += n
	input = input[n:]

	n, err = packet.verifyProtocolName(input)
	if err != nil {
		return 0, err
	}

	totalRead += n
	input = input[n:]

  packet.VariableHeader.Version = input[0]

  totalRead += 1
  input = input[1:]

  connectFlags := input[0]
  packet.VariableHeader.UserNameFlag = connectFlags & 0b10000000 > 0
  packet.VariableHeader.PasswordFlag = connectFlags & 0b01000000 > 0
  packet.VariableHeader.WillRetain = connectFlags & 0b00100000 > 0
  packet.VariableHeader.WillQos = types.QoS((connectFlags & 0b00011000) >> 3)
  packet.VariableHeader.WillFlag = connectFlags & 0b00000100 > 0
  packet.VariableHeader.CleanStart = connectFlags & 0b00000010 > 0

  totalRead += 1
  input = input[1:]

  packet.VariableHeader.KeepAlive.Size = 2
  n, err = packet.VariableHeader.KeepAlive.Decode(input)
	if err != nil {
		return 0, err
	}

	totalRead += n
	input = input[n:]

  n, err = packet.VariableHeader.PropertyLength.Decode(input)
	if err != nil {
		return 0, err
	}

	totalRead += n
	input = input[n+int(packet.VariableHeader.PropertyLength.Value):]

	n, err = packet.Payload.ClientId.Decode(input)
	if err != nil {
		return 0, err
	}

	totalRead += n

	return totalRead, nil
}

func (packet *ConnectPacket) verifyProtocolName(input []byte) (int, error) {
  if len(input) < 6 {
    return 0, errors.New("need at least 6 bytes to verify 'MQTT' in connect packet")
  }

  n, err := packet.VariableHeader.ProtocolName.Decode(input)
  if err != nil {
    return 0, err
  }

  if packet.VariableHeader.ProtocolName.Str != "MQTT" {
    return 0, fmt.Errorf("found %s but expected 'MQTT'", packet.VariableHeader.ProtocolName.Str)
  }

  return n, nil
}
