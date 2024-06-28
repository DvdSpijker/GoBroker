package packet

import (
	"fmt"

	"github.com/DvdSpijker/GoBroker/types"
)

const (
	UnspecifiedBytes PayloadFormatIndicator = 0x00
	UtfCharacterData PayloadFormatIndicator = 0x01
)

type (
	PayloadFormatIndicator byte

	PublishPacket struct {
		FixedHeader struct {
			CommonFixedHeader FixedHeader
			Dup               bool // Duplicate flag, set when client tries to re-deliver.
			Qos               types.QoS
			Retain            bool
		}

		VariableHeader struct {
			TopicName              types.UtfString
			PacketIdentifier       types.UnsignedInt // TODO: Decode packet ID if QoS > 0
			PayloadFormatIndicator PayloadFormatIndicator
			MessageExpirtyInterval types.UnsignedInt
			TopicAlias             types.UnsignedInt
			ResponseTopic          types.UtfString
			CorrelationData        types.BinaryData
			UserProperty           types.UtfStringPair
			SubscriptionIdentifier types.VariableByteInteger
			ContentType            types.UtfString
		}

		Payload struct {
			Data []byte
		}
	}
)

func (packet *PublishPacket) Decode(input []byte) (int, error) {
	totalRead := 0

	n, err := packet.FixedHeader.CommonFixedHeader.Decode(input)
	if err != nil {
		fmt.Println("failed to decode fixed header")
		return 0, err
	}

	packet.FixedHeader.Dup = packet.FixedHeader.CommonFixedHeader.Flags&0b00001000 == 1
	packet.FixedHeader.Qos = types.QoS(packet.FixedHeader.CommonFixedHeader.Flags & 0b00000110)
	packet.FixedHeader.Retain = packet.FixedHeader.CommonFixedHeader.Flags&0b00000001 == 1

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
		// TODO: Decode packet identifier
	}

	propertyLength := types.VariableByteInteger{}
	n, err = propertyLength.Decode(input)
	if err != nil {
		fmt.Println("failed to decode property length")
		return 0, err
	}

	input = input[n:]
	totalRead += n + int(propertyLength.Value) // Pretend properties have been read
	// TODO: Parse properties

	packet.Payload.Data = input
	totalRead += len(input)

	return totalRead, nil
}
