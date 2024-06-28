package packet

import (
	"fmt"

	"github.com/DvdSpijker/GoBroker/types"
)

type (
	ConnectVariableHeader struct {
		VariableHeaderBase VariableHeaderBase
	}

	ConnectPayload struct {
		ClientId types.UtfString
	}

	ConnectPacket struct {
		FixedHeader    FixedHeader
		VariableHeader ConnectVariableHeader
		Payload        ConnectPayload
	}
)

func (header *ConnectVariableHeader) Decode(input []byte) (int, error) {
	if input[2] != 'M' {
		return 0, fmt.Errorf("found %c should be M", input[2])
	}
	if input[3] != 'Q' {
		return 0, fmt.Errorf("found %c should be Q", input[3])
	}
	if input[4] != 'T' {
		return 0, fmt.Errorf("found %c should be T", input[4])
	}
	if input[5] != 'T' {
		return 0, fmt.Errorf("found %c should be T", input[5])
	}

	vbi := types.VariableByteInteger{}
	n, err := vbi.Decode(input[10:])
	if err != nil {
		return n, err
	}

	return 10 + n + int(vbi.Value), nil
}

func (payload *ConnectPayload) Decode(input []byte) (int, error) {
	n, err := payload.ClientId.Decode(input)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (payload *ConnectPayload) String() string {
	return payload.ClientId.String()
}

func (packet *ConnectPacket) Decode(input []byte) (int, error) {
	n, err := packet.FixedHeader.Decode(input)
	if err != nil {
		return 0, err
	}
	fixedHeaderSize := n

	input = input[n:]
	n, err = packet.VariableHeader.Decode(input)
	if err != nil {
		return 0, err
	}

	input = input[n:]
	n, err = packet.Payload.Decode(input)
	if err != nil {
		return 0, err
	}

	return fixedHeaderSize + int(packet.FixedHeader.RemainingLength.Value), nil
}
