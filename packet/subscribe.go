package packet

import (
	"fmt"

	"github.com/DvdSpijker/GoBroker/types"
)

type (
	TopicFilterPair struct {
		TopicFilter         types.UtfString
		SubscriptionOptions byte // TODO: Make options separate fields and parse
	}

	SubscribePacket struct {
		FixedHeader FixedHeader

		VariableHeader struct {
			PacketIdentifier       types.UnsignedInt
			PropertyLength         types.VariableByteInteger
			SubscriptionIdentifier types.VariableByteInteger
			UserProperty           types.UtfStringPair
		}

		Payload struct {
			Filters []TopicFilterPair
		}
	}
)

func (packet *SubscribePacket) Decode(input []byte) (int, error) {
	n, err := packet.FixedHeader.Decode(input)
	if err != nil {
		return 0, err
	}

	input = input[n:]

	packet.VariableHeader.PacketIdentifier.Size = 2
	n, err = packet.VariableHeader.PacketIdentifier.Decode(input)
	if err != nil {
		fmt.Println("failed to parse packet identifier")
		return 0, err
	}

	input = input[n:]

	n, err = packet.VariableHeader.PropertyLength.Decode(input)
	// TODO: Parse properties
	fmt.Printf("property length: %d\n", packet.VariableHeader.PropertyLength.Value)
	if packet.VariableHeader.PropertyLength.Value > 0 {
		input = input[n+int(packet.VariableHeader.PropertyLength.Value):]
	}

	fmt.Printf("input: %x\n", input)
	tpfs, n, err := parseSubscribePayload(input)
	if err != nil {
		fmt.Println("failed to parse subscribe payload")
		return 0, err
	}

	packet.Payload.Filters = tpfs

	fmt.Printf("subscribe packet topic filter: %v\n", tpfs)

	return 0, nil
}

func parseSubscribePayload(input []byte) ([]TopicFilterPair, int, error) {
	tpfs := make([]TopicFilterPair, 0, 1)
	totalRead := 0
	for len(input) > 0 {
		tpf := TopicFilterPair{}
		n, err := tpf.TopicFilter.Decode(input)
		if err != nil {
			fmt.Println("failed to parse topic filter")
			return nil, 0, err
		}

		input = input[n:]
		tpf.SubscriptionOptions = input[0]

		input = input[1:]

		tpfs = append(tpfs, tpf)
		totalRead += n + 1
	}

	return tpfs, totalRead, nil
}
