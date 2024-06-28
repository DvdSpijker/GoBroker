package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"sync"

	"github.com/DvdSpijker/GoBroker/packet"
)

const blockSize = 10

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

		fixedHeader, bytes, err := readPacket(conn)
		if errors.Is(err, io.EOF) {
			fmt.Println("client closed connection:", client.ID)
      disconnect(client)
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

			conackPacket := packet.ConackPacket{}
			conackPacket.VariableHeader.ConnectReasonCode = packet.Success
			bin, err := conackPacket.Encode()
			if err != nil {
				fmt.Println("failed to encode conack packet:", err)
				panic(err)
			}
			n, err = conn.Write(bin)
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
			publish(client, &publishPacket, bytes)

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
			subscribe(client, subscribePacket.Payload.Filters[0].TopicFilter.String())

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
      bin, err := subackPacket.Encode()

      n, err = conn.Write(bin)
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
			n, err := conn.Write(bin)
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

type Client struct {
	ID            string
	Conn          net.Conn
	Subscriptions []string
}

var (
	Clients = make(map[string]*Client)
  ClientSubscriptions = make(map[string][]*Client)
	mutex   = sync.Mutex{}
)

func connect(id string, conn net.Conn) *Client {
	mutex.Lock()
	defer mutex.Unlock()

	_, ok := Clients[id]
	if ok {
		panic("client already connected: " + id)
	}

	fmt.Println(id, "connected")
	client := &Client{ID: id, Conn: conn}
	Clients[id] = client
	return client
}

func disconnect(client *Client) {
  unsubscribeAll(client)

	mutex.Lock()
	defer mutex.Unlock()
  delete(Clients, client.ID)
}

func unsubscribeAll(client *Client) {
  for _, topic := range client.Subscriptions {
    fmt.Println("removing client subscription to", topic)
    unsubscribeTopic(client, topic)
  }
}

func unsubscribeTopic(client *Client, topic string) {
    mutex.Lock()
    defer mutex.Unlock()

    index := slices.Index(ClientSubscriptions[topic], client)
    ClientSubscriptions[topic] = slices.Delete(ClientSubscriptions[topic], index, index+1)
}

func publish(client *Client, p *packet.PublishPacket, packetBytes []byte) {
	topic := p.VariableHeader.TopicName.String()
	fmt.Println(client.ID, "published", string(p.Payload.Data), "to", topic)

	mutex.Lock()
	defer mutex.Unlock()

  for t, clients := range ClientSubscriptions {
    for _, c := range clients {
      if topicMatches(t, topic) {
        fmt.Println(client.ID, "sends to", c.ID, "on topic", topic)
        _, err := c.Conn.Write(
          packetBytes,
        ) // Forward the packet as is for now instead of encoding again.
        if err != nil {
          panic(err)
        }
      }
    }
  }
}

func subscribe(client *Client, topic string) {
	fmt.Println(client.ID, "subbed to", topic)

	mutex.Lock()
	defer mutex.Unlock()

	client.Subscriptions = append(client.Subscriptions, topic)

  if ClientSubscriptions[topic] == nil {
    ClientSubscriptions[topic] = make([]*Client, 0, 1)
  }

  ClientSubscriptions[topic] = append(ClientSubscriptions[topic], client)
}

// TODO: not very efficient probably
func topicMatches(filter, name string) bool {
	if filter == name {
		return true
	}

	filterParts := strings.Split(filter, "/")
	nameParts := strings.Split(name, "/")

	for i := range filterParts {
		if filterParts[i] == "+" && len(nameParts) > i {
			nameParts[i] = "+"
		}

		if filterParts[i] == "#" && len(nameParts) > i {
			nameParts[i] = "#"
			nameParts = nameParts[:i+1]
		}
	}

	if len(nameParts) != len(filterParts) {
		return false
	}

	for i := range filterParts {
		if nameParts[i] != filterParts[i] {
			return false
		}
	}

	return true
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
