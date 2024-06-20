package packet

import "fmt"

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
	fmt.Println("fixed header", packet.FixedHeader)
	return n, nil
}
