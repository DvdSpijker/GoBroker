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

      WillProperties struct {
        PropertiesLength types.VariableByteInteger
        DelayInterval types.UnsignedInt
        PayloadFormatIndicator PayloadFormatIndicator
        MessageExpiryInterval types.UnsignedInt
        ContentType types.UtfString
        ResponseTopic types.UtfString
        CorrelationData types.BinaryData
        UserProperty types.UtfStringPair
      }

      WillTopic types.UtfString
      WillPayload types.BinaryData
      UserName types.UtfString
      Password types.BinaryData
    }
	}
)

func (packet *ConnectPacket) String() string {
  return fmt.Sprintf(`CONNECT
    ClientId: %s
    Version: %d

    UserNameFlag: %t
    PasswordFlag: %t
    WillRetain: %t
    WillQos: %d
    WillFlag: %t
    CleanStart: %t

    KeepAlive: %d sec

    WillProperties
      Length: %d
      DelayInterval: %d
      PayloadFormatIndicator: %x
      MessageExpiryInterval: %d
      ContentType: %s
      ResponseTopic: %s
      CorrelationData: %v
      UserProperty: %s-%s

    WillTopic: %s
    WillPayload: %v
    UserName: %s
    Password: %v
    `, 
    &packet.Payload.ClientId,
    packet.VariableHeader.Version,
    packet.VariableHeader.UserNameFlag,
    packet.VariableHeader.PasswordFlag,
    packet.VariableHeader.WillRetain,
    packet.VariableHeader.WillQos,
    packet.VariableHeader.WillFlag,
    packet.VariableHeader.CleanStart,
    packet.VariableHeader.KeepAlive.Value,
    packet.Payload.WillProperties.PropertiesLength.Value,
    packet.Payload.WillProperties.DelayInterval.Value,
    packet.Payload.WillProperties.PayloadFormatIndicator,
    packet.Payload.WillProperties.MessageExpiryInterval.Value,
    &packet.Payload.WillProperties.ContentType,
    packet.Payload.WillProperties.ResponseTopic.String(),
    packet.Payload.WillProperties.CorrelationData.Data,
    &packet.Payload.WillProperties.UserProperty.Name,
    &packet.Payload.WillProperties.UserProperty.Value,
    &packet.Payload.WillTopic,
    packet.Payload.WillPayload,
    &packet.Payload.UserName,
    packet.Payload.Password.Data)
}

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

  input = input[n:]

  if packet.VariableHeader.WillFlag {
    n, err = packet.Payload.WillProperties.PropertiesLength.Decode(input)
    if err != nil {
      return 0, err
    }

    input = input[n:]
    remainingPropLength := packet.Payload.WillProperties.PropertiesLength.Value

    for len(input) > 0 && err == nil && n != 0  && remainingPropLength > 0 {
      propertyIdentifier := PropertyIdentifier(input[0])
      switch propertyIdentifier {
      case WillDelayIntervalProperty:
        packet.Payload.WillProperties.DelayInterval.Size = 4
        n, err = packet.Payload.WillProperties.DelayInterval.Decode(input[1:])

      case PayloadFormatIndicatorProperty:
        packet.Payload.WillProperties.PayloadFormatIndicator = PayloadFormatIndicator(input[1])
        n = 1

      case MessageExpiryIntervalProperty:
        packet.Payload.WillProperties.MessageExpiryInterval.Size = 4
        n, err = packet.Payload.WillProperties.MessageExpiryInterval.Decode(input[1:])

      case ContentTypeProperty:
        n, err = packet.Payload.WillProperties.ContentType.Decode(input[1:])

      case ResponseTopicProperty:
        n, err = packet.Payload.WillProperties.ResponseTopic.Decode(input[1:])

      case CorrelationDataProperty:
        n, err = packet.Payload.WillProperties.CorrelationData.Decode(input[1:])

      case UserPropertyProperty:
        panic("cannot process user property")
        // TODO: Enable this when Decode is implemeneted.
        // n, err = packet.Payload.WillProperties.UserProperty.Decode(input)

      default:
        fmt.Printf("unknown property identifier in CONNECT payload: %x\n", propertyIdentifier)
    }

      // +1 for the read property identifier
      input = input[n+1:]
      remainingPropLength -= int32(n+1)
    }

    n, err = packet.Payload.WillTopic.Decode(input)
    if err != nil {
      return 0, err
    }

    input = input[n:]

    n, err = packet.Payload.WillPayload.Decode(input)
    if err != nil {
      return 0, err
    }

    input = input[n:]
  }

  if packet.VariableHeader.UserNameFlag {
    n, err = packet.Payload.UserName.Decode(input)
    if err != nil {
      return 0, err
    }

    input = input[n:]

    if packet.VariableHeader.PasswordFlag {
      n, err = packet.Payload.Password.Decode(input)
      if err != nil {
        return 0, err
      }
    }
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
