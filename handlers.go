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
)

const sendQueueSize = 100

type (
	SharedSubscriptionKey struct {
		Topic string
		Group string
	}

	Subscription struct {
		clients      []*Client
		publishIndex int
		shared       bool
	}

	Client struct {
		ID             string
		Conn           net.Conn // If Conn is nil the client is offline
		Subscriptions  []string
		SendQueue      chan []byte
		KeepAlive      time.Duration
		LastWill       protocol.LastWill
		WillDelayTimer *time.Timer
		Ctx            context.Context
		Cancel         context.CancelFunc
	}

	clientSubscriptionMap map[string]Subscription
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

func deleteSubscription(topic string, client *Client) {
	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	sub := ClientSubscriptions[topic]

	if len(ClientSubscriptions[topic].clients) == 1 {
		delete(ClientSubscriptions, topic)
	} else {

		index := slices.Index(sub.clients, client)
		// Current publish index is about to be removed.
		if sub.shared && sub.publishIndex == index {
			incPublishIndex(&sub)
		}

		var clients []*Client
		if index+1 >= len(sub.clients)-1 {
			clients = make([]*Client, 0)
		} else {
			clients = slices.Delete(sub.clients, index, index+1)
		}
		ClientSubscriptions[topic] = Subscription{
			clients:      clients,
			publishIndex: sub.publishIndex,
			shared:       sub.shared,
		}
	}
}

func incPublishIndex(sharedSubscription *Subscription) Subscription {
	sharedSubscription.publishIndex++
	if sharedSubscription.publishIndex >= len(sharedSubscription.clients) {
		sharedSubscription.publishIndex = 0
	}
	return *sharedSubscription
}

func addSubscription(topic string, client *Client) {
	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	sub, ok := ClientSubscriptions[topic]
	if !ok {
		ClientSubscriptions[topic] = Subscription{
			clients: make([]*Client, 0, 10),
			shared:  isSharedSubscription(topic),
		}
		sub = ClientSubscriptions[topic]
	}

	ClientSubscriptions[topic] = Subscription{
		clients:      append(sub.clients, client),
		publishIndex: sub.publishIndex,
		shared:       sub.shared,
	}

	fmt.Println("added subscription for:", client.ID, "topic:", topic, "shared:", isSharedSubscription(topic))
	fmt.Println("total subscribers for topic", topic, ":", len(ClientSubscriptions[topic].clients))
}

func (retained retainedMessageMap) addRetainedMessage(topic string, p *packet.PublishPacket) {
	retainedMessagesMutex.Lock()
	defer retainedMessagesMutex.Unlock()

	if len(p.Payload.Data) == 0 {
		// MQTT-3.3.1-6: If the payload if empty the retained message for a topic is removed.
		retained[topic] = nil
		fmt.Println("removed retained message on topic", topic)
	} else {
		// MQTT-3.3.1-5: New retained message on a topic replaces old.
		fmt.Println("added retained message on topic", topic)
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

	var client *Client
	c, ok := Clients[id]
	if ok {
		if c.Conn != nil {
			// TODO: Send client a disconnect message instead of panicing
			panic("client already connected: " + id)
		}
		c.Conn = conn
		client = c

		// Cancel a delayed Last Will publish when the client
		// has reconnected.
		if client.WillDelayTimer != nil {
			client.WillDelayTimer.Stop()
		}
		fmt.Println("existing client reconnected", id)
	} else {

		ctx, cancel := context.WithCancel(context.Background())
		client = &Client{
			ID:     id,
			Conn:   conn,
			Ctx:    ctx,
			Cancel: cancel,
		}
		Clients[id] = client
		fmt.Println("new client connected", id)
	}

	client.SendQueue = make(chan []byte, 100)
	// 3.1.2-22: The server allows 1.5x the keep-alive period between control packets.
	client.KeepAlive = time.Second * time.Duration(
		math.Round(float64(p.VariableHeader.KeepAlive.Value)*float64(1.7)))

	client.LastWill = copyLastWill(p)

	return client
}

func (client *Client) setkeepAliveDeadline() {
	// 3.1.2.10: A Keep-Alive value of 0 has the effect of turning
	// of the Keep-Alive mechanism.
	if client.KeepAlive > 0 {
		client.Conn.SetReadDeadline(time.Now().Add(client.KeepAlive))
	}
}

func (client *Client) disconnect() {
	client.unsubscribeAll()

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if client.LastWill.WillFlag {
		lastWill := protocol.MakeLastWillPublishPacket(&client.LastWill)

		retainedMessages.addRetainedMessage(client.LastWill.Topic.String(), lastWill)

		if client.LastWill.Properties.DelayInterval > 0 {
			client.WillDelayTimer = time.AfterFunc(client.LastWill.Properties.DelayInterval, func() {
				fmt.Println("publishing delayed last will to", client.LastWill.Topic.String())
				client.publish(lastWill, client.LastWill.Topic.String())
			})
		} else {
			fmt.Println("publishing last will to", client.LastWill.Topic.String())
			client.publish(lastWill, client.LastWill.Topic.String())
		}
	}
	// TODO: Delete client at some point
	// delete(Clients, client.ID)

	client.Cancel()
	client.Conn = nil
}

func (client *Client) unsubscribeAll() {
	for _, topic := range client.Subscriptions {
		fmt.Println("removing client subscription to", topic)
		client.unsubscribeTopic(topic)
	}
}

func (client *Client) unsubscribeTopic(topic string) {
	deleteSubscription(topic, client)
}

func (client *Client) onPublish(p *packet.PublishPacket) {
	topic := p.VariableHeader.TopicName.String()
	fmt.Println(client.ID, "published", string(p.Payload.Data), "to", topic)

	// MQTT-3.3.1-8: If the retained flag is not set the message should not be stored.
	if p.FixedHeader.Retain {
		retainedMessages.addRetainedMessage(topic, p)
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

	client.publish(p, topic)
}

// TODO: This should actually be a server.publish method
func (client *Client) publish(p *packet.PublishPacket, topic string) {
	clientSubscriptionMutex.Lock()
	defer clientSubscriptionMutex.Unlock()

	pub := func(c *Client, bytes []byte) {
		fmt.Println(client.ID, "sends to", c.ID, "on topic", topic)
		_, err := c.Write(
			bytes,
		)
		if err != nil {
			fmt.Println("failed to send publish to", c.ID, err)
		}
	}

	// TODO: Make changes to received packet before forwarding.
	// - New packet id?
	// - Use subscriber's QoS instead of publisher's
	bytes, err := p.Encode()
	if err != nil {
		fmt.Println("failed to encode publish packet", err)
		return
	}

	// Loop over client subscriptions instead of clients because
	// it is more efficient when the larger part of the connected
	// clients have few subscriptions.
	for t, subscription := range ClientSubscriptions {
		if topicMatches(t, topic) {
			if subscription.shared {
				ClientSubscriptions[t] = incPublishIndex(&subscription) // Pre-increment to avoid out of bounds issues.
				fmt.Println("shared subscription:", topic, " publish index:", subscription.publishIndex)
				c := subscription.clients[subscription.publishIndex]
				go pub(c, bytes)
			} else {
				for _, c := range subscription.clients {
					go pub(c, bytes)
				}
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

	// New subscribers to a shared subscription do not received rainted messages.
	if !isSharedSubscription(topic) {
		retainedMessage := retainedMessages.getRetainedMessages(topic)
		if retainedMessage != nil {
			fmt.Printf("sending retained message on topic %s to %s\n", topic, client.ID)
			bytes, err := retainedMessage.Encode()
			if err != nil {
				fmt.Println("failed to encode publish message:", err)
			}
			client.Conn.Write(bytes)
		}
	}

	addSubscription(topic, client)
}

// TODO: not very efficient probably
func topicMatches(filter, name string) bool {
	if filter == name {
		return true
	}

	filterParts := strings.Split(filter, "/")
	if filterParts[0] == "$share" && len(filterParts) > 3 {
		filterParts = filterParts[2:]
	}
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

// writer takes packets that have to be sent by this
// client from a queue.
// It exits when the client context is cancelled.
//
// Writing to a client is done like this to avoid mulitple
// handlers accessing the connection and scrambling packets
// that way.
func (client *Client) writer() {
	for {
		select {
		case bytes, ok := <-client.SendQueue:
			if !ok {
				fmt.Println(client.ID, "send queue closed")
				return
			}
			if client.Conn == nil {
				break
			}
			n, err := client.Conn.Write(bytes)
			if err != nil {
				fmt.Println(client.ID, "write error", err)
			}
			if n != len(bytes) {
				fmt.Printf("%s wrote %d of %d bytes\n", client.ID, n, len(bytes))
			}
		case <-client.Ctx.Done():
			fmt.Println("exitting writer:", client.ID)
			return
		}
	}
}

func copyLastWill(p *packet.ConnectPacket) protocol.LastWill {
	lastWill := protocol.LastWill{}
	lastWill.WillFlag = p.VariableHeader.WillFlag
	if lastWill.WillFlag {
		lastWill.Properties.DelayInterval = time.Second *
			time.Duration(p.Payload.WillProperties.DelayInterval.Value)
		lastWill.Properties.CorrelationData = p.Payload.WillProperties.CorrelationData
		lastWill.Properties.ContentType = p.Payload.WillProperties.ContentType
		lastWill.Properties.MessageExpiryInterval = time.Second *
			time.Duration(p.Payload.WillProperties.MessageExpiryInterval.Value)
		lastWill.Properties.ReponseTopic = p.Payload.WillProperties.ResponseTopic
		lastWill.Qos = p.VariableHeader.WillQos
		lastWill.Payload = p.Payload.WillPayload
		lastWill.Topic = p.Payload.WillTopic
		lastWill.Retain = p.VariableHeader.WillRetain
	}
	return lastWill
}

func isSharedSubscription(topic string) bool {
	parts := strings.Split(topic, "/")

	if len(parts) < 3 {
		return false
	} else if parts[0] != "$share" {
		return false
	}

	return true
}
