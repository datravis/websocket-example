package http

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/datravis/websocket-example/internal/pkg/publish"
)

// PubSubServer implements an HTTP server backed by the InMemoryPublisher in the publish package
//
// PubSubServer enables an HTTP client to subscribe to messages to a given topic over a web socket.
// A /publish endpoint enables a client to publish arbitrary JSON messages to a provided topic.
type PubSubServer struct {
	address      string
	publisher    *publish.InMemoryPublisher
	upgrader     websocket.Upgrader
	contentTypes map[string]bool
}

// NewPubSubServer creates a PubSubServer configured with the provided address and InMemoryPublisher
func NewPubSubServer(address string, publisher *publish.InMemoryPublisher) *PubSubServer {
	return &PubSubServer{
		address:      address,
		publisher:    publisher,
		upgrader:     websocket.Upgrader{},
		contentTypes: map[string]bool{"application/json": true},
	}
}

// Subscribe handles requests to subscribe to a topic. Connections are upgraded to the websocket protocol.
// When messages are published to the subscribed topic, they are returned to the client as a JSON string.
func (s *PubSubServer) Subscribe(w http.ResponseWriter, r *http.Request) {
	// ensure a topic was provided
	queryParams := r.URL.Query()
	topic := queryParams.Get("topic")
	if topic == "" {
		log.Println("ERROR: topic is empty")
		http.Error(w, "Please include the topic query parameter", http.StatusBadRequest)
		return
	}

	// upgrade the connection to a websocket
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: unable to upgrade connection: %v", err)
		return
	}
	defer c.Close()

	// subscribe to our topic, and ensure we unsubscribe if the client disconnects
	consumer := s.publisher.Subscribe(topic)
	defer s.publisher.UnsubscribeAndClose(consumer)

	// As messages are received, return them to the client
	for message := range consumer.Channel {
		err = c.WriteJSON(message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// Publish handles requests to publish messages to a topic. A simple HTTP POST is supported.
// The Client is expected to set the correct Content-Type, topic query param, and to provide valid JSON in the
// request body.
func (s *PubSubServer) Publish(w http.ResponseWriter, r *http.Request) {
	// ensure the proper content type is set
	contentType := r.Header.Get("Content-Type")
	if !s.contentTypes[contentType] {
		log.Printf("ERROR: Content-Type is %v", contentType)
		http.Error(w, "Please confirm your Content-Type is supported", http.StatusBadRequest)
	}

	// ensure the topic query param was provided
	queryParams := r.URL.Query()
	topic := queryParams.Get("topic")
	if topic == "" {
		log.Println("ERROR: topic is empty")
		http.Error(w, "Please include the topic query parameter", http.StatusBadRequest)
		return
	}

	// bail if the there's no message body
	if r.ContentLength == 0 {
		log.Println("ERROR: Content-Length is 0 length")
		http.Error(w, "Please include a request body", http.StatusBadRequest)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: error reading body: %v", err)
		http.Error(w, "Unable to read message body", http.StatusInternalServerError)
	}

	// Since no specific JSON schema is specified in the requirements, unmarshall to an empty interface just to ensure
	// the data is valid JSON.
	var message interface{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		log.Printf("ERROR: error unmarshalling message: %v", err)
		http.Error(w, "Please ensure your message body is properly formatted JSON", http.StatusBadRequest)
	}

	// publish the message to all clients subscribed to this topic
	s.publisher.Publish(topic, message)
}

// Serve configures the HTTP routing and starts the server
func (s *PubSubServer) Serve() error {
	r := mux.NewRouter()
	r.HandleFunc("/subscribe", s.Subscribe).Methods("GET")
	r.HandleFunc("/publish", s.Publish).Methods("POST")
	http.Handle("/", r)

	log.Printf("Listening on %v", s.address)
	log.Println("HTTP Endpoints:")
	log.Println("GET /subscribe")
	log.Println("POST /publish")

	return http.ListenAndServe(s.address, nil)
}
