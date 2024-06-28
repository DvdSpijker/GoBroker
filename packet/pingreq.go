package packet

type (
	PingReqPacket struct {
		FixedHeader FixedHeader
	}
)

func (packet PingReqPacket) Decode(input []byte) (int, error) {
	n, err := packet.FixedHeader.Decode(input)
	if err != nil {
		return 0, err
	}
	return n, nil
}
