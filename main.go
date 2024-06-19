package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
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
	var client *Client
	for {
		msg := make([]byte, 0)

		for {
			b := make([]byte, blockSize)
			n, err := conn.Read(b)
			if err != nil {
				panic(err)
			}

			b = b[:n]
			msg = append(msg, b...)

			// TODO: we need packet size info
			// to make this work properly
			if n < blockSize {
				break
			}
		}

		msg = msg[:len(msg)-2]

		fmt.Printf("%#v\n", string(msg))

		// Check message type and call that handler

		//
		// tmp
		switch string(msg[0:3]) {
		case "con":
			client = connect(string(msg[3:]), conn)
		case "pub":
			if client == nil {
				panic("pub before con")
			}
			parts := strings.SplitN(string(msg[3:]), "-", 2)
			if len(parts) == 1 {
				parts = append(parts, "empty")
			}
			publish(client, parts[0], parts[1])
		case "sub":
			if client == nil {
				panic("sub before con")
			}
			subscribe(client, string(msg[3:]))
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

func publish(client *Client, topic string, payload string) {
	fmt.Println(client.ID, "published to", topic)

	mutex.Lock()
	defer mutex.Unlock()

	for _, c := range Clients {
		for _, t := range c.Subscriptions {
			if topicMatches(t, topic) {
				fmt.Println(client.ID, "sends to", c.ID, "on topic", topic)
				_, err := c.Conn.Write([]byte(topic + ": " + payload))
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
