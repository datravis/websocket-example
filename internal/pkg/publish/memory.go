package publish

import (
	"sync"

	"github.com/google/uuid"
)

// InMemoryPublisher implements a simple in memory Pub/Sub capability, across multiple topics.
//
// The caller can use the Subscribe/Unsubscribe methods to create/delete a client.
// Messages can be published with the Publish method.
type InMemoryPublisher struct {
	mux sync.RWMutex
	// nested maps enable us to quickly check for the existence of a given topic & consumer
	subscriptions map[string]map[string]*PublisherClient
}

// PublisherClient is a container for the metadata for this subscription, along with the channel messages
// will be published to.
type PublisherClient struct {
	Topic      string
	ConsumerId string
	Channel    chan interface{}
}

// NewInMemoryPublisher creates an InMemoryPublisher with an empty set of subscriptions
func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subscriptions: make(map[string]map[string]*PublisherClient),
	}
}

// Subscribe creates a unique client and subscribes it to the provided topic.
// The client is then returned to the caller.
func (p *InMemoryPublisher) Subscribe(topic string) *PublisherClient {
	p.mux.Lock()
	defer p.mux.Unlock()

	// initialize the map of consumers for this topic, if it's not already initialized
	if p.subscriptions[topic] == nil {
		p.subscriptions[topic] = make(map[string]*PublisherClient)
	}

	consumerId := uuid.New().String()
	p.subscriptions[topic][consumerId] = &PublisherClient{
		Topic:      topic,
		ConsumerId: consumerId,
		Channel:    make(chan interface{}),
	}

	return p.subscriptions[topic][consumerId]
}

// UnsubscribeAndClose unsubscribes a client and closes the client's channel.
func (p *InMemoryPublisher) UnsubscribeAndClose(client *PublisherClient) {
	p.mux.Lock()
	defer p.mux.Unlock()

	close(client.Channel)
	delete(p.subscriptions[client.Topic], client.ConsumerId)
}

// Publish publishes a message to all consumers subscribed to a topic.
func (p *InMemoryPublisher) Publish(topic string, message interface{}) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	for _, c := range p.subscriptions[topic] {
		c.Channel <- message
	}
}
