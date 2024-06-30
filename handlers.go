package main

import (
  "net"
  "sync"
  "fmt"
  "slices"
  "strings"
  "context"

	"github.com/DvdSpijker/GoBroker/packet"
)

const sendQueueSize = 100

type Client struct {
	ID            string
	Conn          net.Conn
	Subscriptions []string
  SendQueue chan []byte
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
    // TODO: Send client a disconnect message instead of panicing
		panic("client already connected: " + id)
	}

	fmt.Println(id, "connected")
	client := &Client{ID: id, Conn: conn}
	Clients[id] = client
  client.SendQueue = make(chan []byte, 100)
	return client
}

func (client *Client) disconnect() {
  client.unsubscribeAll()

	mutex.Lock()
	defer mutex.Unlock()
  delete(Clients, client.ID)
}

func (client *Client) unsubscribeAll() {
  for _, topic := range client.Subscriptions {
    fmt.Println("removing client subscription to", topic)
    client.unsubscribeTopic(topic)
  }
}

func (client *Client) unsubscribeTopic(topic string) {
    mutex.Lock()
    defer mutex.Unlock()

    index := slices.Index(ClientSubscriptions[topic], client)
    ClientSubscriptions[topic] = slices.Delete(ClientSubscriptions[topic], index, index+1)
}

func (client *Client) publish(p *packet.PublishPacket, packetBytes []byte) {
	topic := p.VariableHeader.TopicName.String()
	fmt.Println(client.ID, "published", string(p.Payload.Data), "to", topic)

	mutex.Lock()
	defer mutex.Unlock()

  // Loop over client subscriptions instead of clients because
  // it is more efficient when the largers part of the connected
  // clients have few subscriptions.cclient
  for t, clients := range ClientSubscriptions {
    for _, c := range clients {
      if topicMatches(t, topic) {
        fmt.Println(client.ID, "sends to", c.ID, "on topic", topic)
        _, err := c.Write(
          packetBytes,
        ) // Forward the packet as is for now instead of encoding again.
        if err != nil {
          panic(err)
        }
      }
    }
  }
}

func (client *Client) subscribe(topic string) {
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

// Implements io.Writer
func (client *Client) Write(p []byte) (n int, err error) {
  client.SendQueue <- p
  return len(p), nil
}

func (client *Client) writer(ctx context.Context) {
  for {
    select {
    case bytes, ok := <-client.SendQueue:
    if !ok {
        fmt.Println(client.ID, "send queue closed")
        return
      }
      n, err := client.Conn.Write(bytes)
      if err != nil {
        fmt.Println(client.ID, "write error", err)
      }
      if n != len(bytes) {
        fmt.Printf("%s wrote %d of %d bytes\n", client.ID, n, len(bytes))
      }
    case <-ctx.Done():
      fmt.Println(client.ID, "exitting writer")
      return
  }
  }
}
