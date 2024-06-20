package packet

type (
	PingRespPacket struct {
		FixedHeader FixedHeader
	}
)

func (packet PingRespPacket) Encode() (bin []byte, err error) {
	packet.FixedHeader.PacketType = PINGRESP

	fixedHdrBin, err := packet.FixedHeader.Encode()
	if err != nil {
		return nil, err
	}
	bin = append(bin, fixedHdrBin...)

	return bin, nil
}
