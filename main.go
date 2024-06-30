package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/DvdSpijker/GoBroker/packet"
	"github.com/DvdSpijker/GoBroker/protocol"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("new connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
  defer println("----------")

  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

	var client *Client
	for {
		println("----------")

		fixedHeader, bytes, err := readPacket(conn)
		if errors.Is(err, io.EOF) {
			fmt.Println("client closed connection:", client.ID)
      client.disconnect()
			return
		}
		if err != nil {
			fmt.Println("packet read error", err)
			panic("whoop whoop")
		}

		switch fixedHeader.PacketType {

		case packet.CONNECT:
      fmt.Println("connect")
			connectPacket := packet.ConnectPacket{}
			n, err := connectPacket.Decode(bytes)
			if err != nil {
				fmt.Println("invalid connect packet:", err)
				panic(err)
			}
			_ = n
			client = connect(connectPacket.Payload.ClientId.String(), conn)

      go client.writer(ctx)

			conackPacket := packet.ConackPacket{}
			conackPacket.VariableHeader.ConnectReasonCode = packet.Success
			bin, err := conackPacket.Encode()
			if err != nil {
				fmt.Println("failed to encode conack packet:", err)
				panic(err)
			}
			n, err = client.Write(bin)
			if err != nil {
				fmt.Println("failed to send conack packet:", err)
				panic(err)
			}
      fmt.Println("conack")
			_ = n

		case packet.DISCONNECT:
      println("client disconnecting:", client.ID)

		case packet.PUBLISH:
			if client == nil {
				panic("pub before con")
			}
			publishPacket := packet.PublishPacket{}
			n, err := publishPacket.Decode(bytes)
			if err != nil {
				fmt.Println("invalid publish packet:", err)
				panic(err)
			}
			bytes = bytes[:n]
			client.publish(&publishPacket, bytes)

		case packet.SUBSCRIBE:
			if client == nil {
				panic("sub before con")
			}
      fmt.Println("subscribe")
      subscribePacket := packet.SubscribePacket{}
      n, err := subscribePacket.Decode(bytes)
			if err != nil {
				fmt.Println("invalid subscribe packet:", err)
				panic(err)
			}
      _ = n
      // TODO: Subscribe to all topics in Filters
			client.subscribe(subscribePacket.Payload.Filters[0].TopicFilter.String())

      subackPacket := protocol.MakeSuback(&subscribePacket)
      bin, err := subackPacket.Encode()

      n, err = client.Write(bin)
      if err != nil || n != len(bin) {
        panic("failed to write suback")
      }
      fmt.Println("suback")

		case packet.PINGREQ:
			println("pingreq")

			pingRespPacket := packet.PingRespPacket{}
			bin, err := pingRespPacket.Encode()
			if err != nil {
				fmt.Println("failed to encode conack packet:", err)
				panic(err)
			}
			bin = append(bin, 0x00) // the rest of the message is 0 bytes
			n, err := client.Write(bin)
			if err != nil {
				fmt.Println("failed to send conack packet:", err)
				panic(err)
			}
			_ = n
			println("pingresp")
		default:
			panic("unknown")
		}

	}
}

func readPacket(conn net.Conn) (packet.FixedHeader, []byte, error) {
  const fixedHeaderMaxLength = 5
	headerBytes := make([]byte, fixedHeaderMaxLength)
	n, err := conn.Read(headerBytes)
	if err != nil {
		return packet.FixedHeader{}, nil, err
	}
	headerBytesRead := n

	fixedHeader := packet.FixedHeader{}
	n, err = fixedHeader.Decode(headerBytes)
	if err != nil {
		return packet.FixedHeader{}, nil, err
	}

  println("read header bytes:", n)
	if headerBytesRead < fixedHeaderMaxLength {
		return fixedHeader, nil, nil
	}

  // Part of the bytes that were read might not be part of the fixed header,
  // depending on n.
	packetBytes := make([]byte, int(fixedHeader.RemainingLength.Value)-(fixedHeaderMaxLength-n))
	println("bytes left to read:", len(packetBytes))

	n, err = conn.Read(packetBytes)
	if err != nil {
		return packet.FixedHeader{}, nil, err
	}
	if n != len(packetBytes) {
		return packet.FixedHeader{}, nil, fmt.Errorf(
			"read %d bytes instead of %d",
			n,
			len(packetBytes),
		)
	}

	return fixedHeader, append(headerBytes, packetBytes...), nil
}
