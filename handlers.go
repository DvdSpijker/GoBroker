package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DvdSpijker/GoBroker/packet"
	"github.com/DvdSpijker/GoBroker/protocol"
	"github.com/DvdSpijker/GoBroker/types"
)

const sendQueueSize = 100

type (
	LastWill struct {
		Qos     types.QoS
		Retain  bool
		Topic   string
		Payload types.BinaryData
	}
	Client struct {
		ID            string
		Conn          net.Conn
		Subscriptions []string
		SendQueue     chan []byte
		KeepAlive     time.Duration
	}

	clientSubscriptionMap map[string][]*Client
	retainedMessageMap    map[string]*packet.PublishPacket
)

var (
	clientsMutex = sync.Mutex{}
	Clients      = make(map[string]*Client)

	clientSubscriptionMutex = sync.Mutex{}
	ClientSubscriptions     = make(clientSubscriptionMap)

	retainedMessagesMutex = sync.Mutex{}
	retainedMessages      = make(retainedMessageMap)
)

func (subs clientSubscriptionMap) deleteSubscription(topic string, client *Client) {
	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	index := slices.Index(subs[topic], client)
	subs[topic] = slices.Delete(subs[topic], index, index+1)
}

func (subs clientSubscriptionMap) addSubscription(topic string, client *Client) {
	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	if subs[topic] == nil {
		subs[topic] = make([]*Client, 0, 1)
	}

	subs[topic] = append(subs[topic], client)
}

// TODO: Only pass packet to this function once encode works.
func (retained retainedMessageMap) addRetainedMessage(topic string, p *packet.PublishPacket) {
	retainedMessagesMutex.Lock()
	defer retainedMessagesMutex.Unlock()

	if len(p.Payload.Data) == 0 {
		// MQTT-3.3.1-6: If the payload if empty the retained message for a topic is removed.
		retained[topic] = nil
		fmt.Println("removed retained message on topic", topic)
	} else {
		// MQTT-3.3.1-5: New retained message on a topic replaces old.
		retained[topic] = p
	}
}

func (retained retainedMessageMap) getRetainedMessages(topic string) *packet.PublishPacket {
	retainedMessagesMutex.Lock()
	defer retainedMessagesMutex.Unlock()

	// Direct topic match
	message, ok := retained[topic]
	if ok {
		return message
	}

	for t, message := range retained {
		if topicMatches(topic, t) {
			return message
		}
	}

	return nil
}

func connect(id string, conn net.Conn, p *packet.ConnectPacket) *Client {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	_, ok := Clients[id]
	if ok {
		// TODO: Send client a disconnect message instead of panicing
		panic("client already connected: " + id)
	}

	fmt.Println(id, "connected")
	client := &Client{ID: id, Conn: conn}
	Clients[id] = client
	client.SendQueue = make(chan []byte, 100)
	// 3.1.2-22: The server allows 1.5x the keep-alive period between control packets.
	client.KeepAlive = time.Second * time.Duration(
		math.Round(float64(p.VariableHeader.KeepAlive.Value)*float64(1.5)))

	return client
}

func (client *Client) setkeepAliveDeadline() {
	// 3.1.2.10: A keep-alive value of 0 has the effect of turning
	// of the Keep-Alive mechanism.
	if client.KeepAlive > 0 {
		client.Conn.SetReadDeadline(time.Now().Add(client.KeepAlive))
	}
}

func (client *Client) disconnect() {
	client.unsubscribeAll()

	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	delete(Clients, client.ID)
}

func (client *Client) unsubscribeAll() {
	for _, topic := range client.Subscriptions {
		fmt.Println("removing client subscription to", topic)
		client.unsubscribeTopic(topic)
	}
}

func (client *Client) unsubscribeTopic(topic string) {
	ClientSubscriptions.deleteSubscription(topic, client)
}

func (client *Client) onPublish(p *packet.PublishPacket) {
	topic := p.VariableHeader.TopicName.String()
	fmt.Println(client.ID, "published", string(p.Payload.Data), "to", topic)

	// MQTT-3.3.1-8: If the retained flag is not set the message should not be stored.
	if p.FixedHeader.Retain {
		retainedMessages.addRetainedMessage(topic, p)
		fmt.Println(client.ID, "message retained:", retainedMessages)
	}

	if p.FixedHeader.Qos > 0 {
		fmt.Printf("puback to %s on %s\n", client.ID, topic)
		pubackPacket := protocol.MakePuback(p)
		bytes, err := pubackPacket.Encode()
		if err != nil {
			fmt.Println("failed to encode puback packet:", err)
		}
		go func(client *Client, bytes []byte) {
			client.Write(bytes)
		}(client, bytes)
	}

	// TODO: Make changes to received packet before forwarding.
	// - New packet id?
	bytes, err := p.Encode()
	if err != nil {
		fmt.Println("failed to encode publish packet", err)
		return
	}

	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	// Loop over client subscriptions instead of clients because
	// it is more efficient when the largers part of the connected
	// clients have few subscriptions.
	for t, subscription := range ClientSubscriptions {
		if topicMatches(t, topic) {
			for _, c := range subscription {
				go func(c *Client) { // Use a goroutine here to avoid blocking by a single client.
					fmt.Println(client.ID, "sends to", c.ID, "on topic", topic)
					_, err := c.Write(
						bytes,
					)
					if err != nil {
						fmt.Println("failed to send publish to", c.ID, err)
					}
				}(c)
			}
		}
	}
}

func (client *Client) puback(p *packet.PubackPacket) {
	fmt.Printf("puback from %s on packet %d\n",
		client.ID,
		p.VariableHeader.PacketIdentifer.Value)
	return
}

func (client *Client) subscribe(topic string) {
	fmt.Println(client.ID, "subbed to", topic)

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	client.Subscriptions = append(client.Subscriptions, topic)

	retainedMessage := retainedMessages.getRetainedMessages(topic)
	if retainedMessage != nil {
		fmt.Printf("sending retained message on topic %s to %s\n", topic, client.ID)
		// TODO: Encode packet, clear necessary flags then encode and send.
		bytes, err := retainedMessage.Encode()
		if err != nil {
			fmt.Println("failed to encode publish message:", err)
		}
		client.Conn.Write(bytes)
	}

	ClientSubscriptions.addSubscription(topic, client)
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
// Write puts the packet bytes in a queue to be handled by
// the client's writer routine.
func (client *Client) Write(p []byte) (n int, err error) {
	client.SendQueue <- p
	return len(p), nil
}

// writer takes packets that have to be sent to this
// client from a queue.
// It exits when the context is cancelled.
//
// Writing to a client is done like this to avoid mulitple
// handlers accessing the connection and scrambling packet
// that way.
// This way of writing also avoids the need for a lock on the connection.
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
