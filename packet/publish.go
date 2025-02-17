package packet

import (
	"encoding/hex"
	"fmt"

	"github.com/DvdSpijker/GoBroker/types"
)

type (
	PublishFixedHeader struct {
		CommonFixedHeader FixedHeader
		Dup               bool // Duplicate flag, set when client tries to re-deliver.
		Qos               types.QoS
		Retain            bool
	}
	PublishVariableHeader struct {
		TopicName              types.UtfString
		PacketIdentifier       types.UnsignedInt
		PropertyLength         types.VariableByteInteger
		PropertiesRaw          []byte // Temporarily used to store the properties until they are parsed.
		PayloadFormatIndicator PayloadFormatIndicator
		MessageExpirtyInterval types.UnsignedInt
		TopicAlias             types.UnsignedInt
		ResponseTopic          types.UtfString
		CorrelationData        types.BinaryData
		UserProperty           types.UtfStringPair
		SubscriptionIdentifier types.VariableByteInteger
		ContentType            types.UtfString
	}
	PublishPayload struct {
		Data []byte
	}
	PublishPacket struct {
		FixedHeader    PublishFixedHeader
		VariableHeader PublishVariableHeader
		Payload        PublishPayload
	}
)

func PublishPacketFlags(qos types.QoS, dup bool, retain bool) PacketFlag {
	var dupInt, qosInt, retainInt int
	if retain {
		retainInt = 1
	}
	if dup {
		dupInt = 1
	}
	qosInt = int(qos)
	return PacketFlag(dupInt<<4 | qosInt<<1 | retainInt)
}

func (packet *PublishPacket) Decode(input []byte) (int, error) {
	totalRead := 0

	n, err := packet.FixedHeader.CommonFixedHeader.Decode(input)
	if err != nil {
		fmt.Println("failed to decode fixed header")
		return 0, err
	}

	packet.FixedHeader.Dup = packet.FixedHeader.CommonFixedHeader.Flags&0b00001000 > 0
	packet.FixedHeader.Qos = types.QoS(packet.FixedHeader.CommonFixedHeader.Flags & 0b00000110)
	packet.FixedHeader.Retain = packet.FixedHeader.CommonFixedHeader.Flags&0b00000001 > 0

	input = input[n:]
	totalRead += n

	n, err = packet.VariableHeader.TopicName.Decode(input)
	if err != nil {
		fmt.Println("failed to decode topic name")
		return 0, err
	}

	input = input[n:]
	totalRead += n

	if packet.FixedHeader.Qos > 0 {
		packet.VariableHeader.PacketIdentifier.Size = 2
		n, err := packet.VariableHeader.PacketIdentifier.Decode(input)
		if err != nil {
			fmt.Println("failed to decode packet identifier")
			return 0, err
		}
		input = input[n:]
		totalRead += n
	}

	if len(input) > 1 {
		n, err = packet.VariableHeader.PropertyLength.Decode(input)
		if err != nil {
			fmt.Println("failed to decode property length")
			return 0, err
		}

		n = int(packet.VariableHeader.PropertyLength.Value)
		if n > 0 {
			input = input[n:]

			packet.VariableHeader.PropertiesRaw = input[:n]                  // TODO: Actually parse properties
			totalRead += n + int(packet.VariableHeader.PropertyLength.Value) // Pretend properties have been read
			input = input[n:]
		}
	}

	packet.Payload.Data = input
	totalRead += len(input)

	return totalRead, nil
}

func (packet *PublishPacket) Encode() ([]byte, error) {
	bytes := []byte{}

	b, err := packet.VariableHeader.TopicName.Encode()
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, b...)

	if packet.FixedHeader.Qos > 0 {
		b, err = packet.VariableHeader.PacketIdentifier.Encode()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, b...)
	}

	if packet.VariableHeader.PropertyLength.Value > 0 || len(packet.Payload.Data) > 0 {
		b, err = packet.VariableHeader.PropertyLength.Encode()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, b...)
		bytes = append(bytes, packet.VariableHeader.PropertiesRaw...)
	}

	if len(packet.Payload.Data) > 0 {
		bytes = append(bytes, packet.Payload.Data...)
	}

	packet.FixedHeader.CommonFixedHeader.RemainingLength.Value = int32(len(bytes))

	b, err = packet.FixedHeader.CommonFixedHeader.Encode()

	return append(b, bytes...), nil
}

func (packet *PublishPacket) String() string {
	return fmt.Sprintf(`PUBLISH
    Common header: %s
    Retain: %t
    QoS: %d
    Duplicate: %t
    Topic: %s
    PropertyLength: %d
    Properties: %s
    Payload (%d): %s`,
		packet.FixedHeader.CommonFixedHeader.String(),
		packet.FixedHeader.Retain,
		packet.FixedHeader.Qos,
		packet.FixedHeader.Dup,
		packet.VariableHeader.TopicName.String(),
		packet.VariableHeader.PropertyLength.Value,
		hex.EncodeToString(packet.VariableHeader.PropertiesRaw),
		len(packet.Payload.Data),
		packet.Payload.Data)
}
