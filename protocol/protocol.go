package protocol

import "github.com/DvdSpijker/GoBroker/packet"

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
