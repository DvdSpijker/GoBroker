package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/DvdSpijker/GoBroker/packet"
	"github.com/DvdSpijker/GoBroker/protocol"
)

const connectTimeout = time.Second * 5

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

	var client *Client
	for {
		println("----------")

		// Use the client to control the keep-alive deadline for the connection
		// if the client exists (which is created upon connect)
		// Use a fixed amount of time allowed between the client opening a connection and
		// sending a connect message if there is not client.
		// The specification is unclear about what that amount of time should be,
		// it specifies 'reasonable'.
		if client != nil {
			// Reset keep-alive after receiving a control packet.
			client.setkeepAliveDeadline()
		} else {
			conn.SetReadDeadline(time.Now().Add(connectTimeout))
		}

		fixedHeader, bytes, err := readPacket(conn)
		if errors.Is(err, io.EOF) {
			fmt.Println("client closed connection:", client.ID)
			client.disconnect()
			return
		} else if errors.Is(err, os.ErrDeadlineExceeded) {
			if client != nil {
				fmt.Printf("no control packet received within keep-alive timeout from %s:\n",
					client.ID)
				client.disconnect()
			} else {
				fmt.Println("no connect received after client opened connection")
			}
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
			fmt.Println(connectPacket.String())
			client = connect(connectPacket.Payload.ClientId.String(), conn, &connectPacket)

			go client.writer()

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
			_ = n
			client.onPublish(&publishPacket)

		case packet.PUBACK:
			fmt.Println("puback")
			pubackPacket := packet.PubackPacket{}
			n, err := pubackPacket.Decode(bytes)
			if err != nil {
				fmt.Println("invalid subscribe packet:", err)
				panic(err)
			}
			_ = n
			client.puback(&pubackPacket)

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
