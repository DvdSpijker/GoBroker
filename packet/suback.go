package packet

import "github.com/DvdSpijker/GoBroker/types"

const (
  GrantedQoS0 ReasonCode = 0x00
  GrantedQoS1 ReasonCode = 0x01
  GrantedQoS2 ReasonCode = 0x02
  NotAuthorized ReasonCode = 0x87
  TopicFilterInvalid ReasonCode = 0x8F
  PacketIdentifierInUse ReasonCode = 0x91
  SharedSubscriptionsNotSupported ReasonCode = 0x9E
  SubscriptionIdentifierNotSupported ReasonCode = 0xA1
  WildCardSubscriptionsNotSUpported ReasonCode = 0xA2
)
type (
  ReasonCode byte

  SubackVariableHeader struct {
    PacketIdentifier types.UnsignedInt
    PropertyLength types.VariableByteInteger
    ReasonString types.UtfString
    UserProperty types.UtfStringPair
  }

  SubackPayload struct {
    ReasonCodes []ReasonCode
  }

  SubackPacket struct {
    FixedHeader FixedHeader
    VariableHeader SubackVariableHeader
    Payload SubackPayload
  }
)


func (packet *SubackPacket) Encode() ([]byte, error) {

  bytes := []byte{}
  packet.FixedHeader.PacketType = SUBACK
  packet.FixedHeader.Flags = 0

  packet.VariableHeader.PropertyLength.Value = 0
  b, err := packet.VariableHeader.Encode()
  if err != nil {
    return nil, err
  }

  bytes = append(bytes, b...)

  b, err = packet.Payload.Encode()
  if err != nil {
    return nil, err
  }

  bytes = append(bytes, b...)

  packet.FixedHeader.RemainingLength.Value = int32(len(bytes))

  b, err = packet.FixedHeader.Encode()
  if err != nil {
    return nil, err
  }

  return append(b, bytes...), nil
}

func (header *SubackVariableHeader) Encode() ([]byte, error) {
  bytes := []byte{}

  b, err := header.PacketIdentifier.Encode()
  if err != nil {
    return nil, err
  }

  bytes = append(bytes, b...)

  header.PropertyLength.Value = 0 // TODO: Allow properties to be set.
  b, err = header.PropertyLength.Encode()
  if err != nil {
    return nil, err
  }

  return append(bytes, b...), nil
}

func (payload *SubackPayload) Encode() ([]byte, error) {
  bytes := make([]byte, 0, len(payload.ReasonCodes))

  for _, reasonCode := range payload.ReasonCodes {
    bytes = append(bytes, byte(reasonCode))
  }

  return bytes, nil
}
