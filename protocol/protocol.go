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
