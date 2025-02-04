package protocol

import (
	"math/rand"
	"time"

	"github.com/DvdSpijker/GoBroker/packet"
	"github.com/DvdSpijker/GoBroker/types"
)

type (
	LastWill struct {
		WillFlag   bool
		Qos        types.QoS
		Retain     bool
		Topic      types.UtfString
		Payload    types.BinaryData
		Properties struct {
			DelayInterval         time.Duration
			MessageExpiryInterval time.Duration
			ContentType           types.UtfString
			ReponseTopic          types.UtfString
			CorrelationData       types.BinaryData
			// TODO: All properties
		}
	}
)

func NewPacketIdentifier() types.UnsignedInt {
	id := types.UnsignedInt{
		Value: uint32(rand.Intn(65535)),
		Size:  2,
	}
	return id
}

func MakeSuback(subscribePacket *packet.SubscribePacket) *packet.SubackPacket {
	subackPacket := packet.SubackPacket{
		VariableHeader: packet.SubackVariableHeader{
			PacketIdentifier: subscribePacket.VariableHeader.PacketIdentifier,
		},
		Payload: packet.SubackPayload{
			ReasonCodes: []packet.ReasonCode{
				packet.GrantedQoS0,
			},
		},
	}

	return &subackPacket
}

func MakePuback(publishPacket *packet.PublishPacket) *packet.PubackPacket {
	pubackPacket := packet.PubackPacket{
		VariableHeader: packet.PubackVariableHeader{
			PacketIdentifer: publishPacket.VariableHeader.PacketIdentifier,
		},
	}

	pubackPacket.VariableHeader.PropertyLength.Value = 0

	return &pubackPacket
}

func MakeLastWillPublishPacket(lastWill *LastWill) *packet.PublishPacket {
	pub := packet.PublishPacket{
		FixedHeader: packet.PublishFixedHeader{
			CommonFixedHeader: packet.FixedHeader{
				PacketType: packet.PUBLISH,
				Flags:      packet.PublishPacketFlags(lastWill.Qos, false, lastWill.Retain),
			},
			Dup:    false,
			Qos:    lastWill.Qos,
			Retain: lastWill.Retain,
		},
		VariableHeader: packet.PublishVariableHeader{
			TopicName:        lastWill.Topic,
			PacketIdentifier: NewPacketIdentifier(),
			PropertyLength:   types.VariableByteInteger{Value: 0},
			PropertiesRaw:    []byte{},
			// ContentType: lastWill.Properties.ContentType,
			// ResponseTopic: lastWill.Properties.ReponseTopic,
			// CorrelationData: lastWill.Properties.CorrelationData,
		},
		Payload: packet.PublishPayload(lastWill.Payload),
	}

	return &pub
}
