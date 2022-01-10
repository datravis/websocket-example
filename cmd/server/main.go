package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type InMemoryPublisher struct {
	mux           sync.RWMutex
	subscriptions map[string][]chan interface{}
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subscriptions: make(map[string][]chan interface{}),
	}
}

func (p *InMemoryPublisher) Subscribe(topic string, channel chan interface{}) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.subscriptions[topic] = append(p.subscriptions[topic], channel)
}

func (p *InMemoryPublisher) Publish(topic string, message interface{}) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	for _, c := range p.subscriptions[topic] {
		c <- message
	}
}

var addr = flag.String("addr", "localhost:8081", "http service address")

var upgrader = websocket.Upgrader{}

func (p *InMemoryPublisher) subscribe(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	topic := queryParams.Get("topic")
	if topic == "" {
		log.Println("ERROR: Topic is empty")
		http.Error(w, "Please include the topic query parameter", http.StatusBadRequest)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: Unable to upgrade connection: %v", err)
		return
	}
	defer c.Close()

	channel := make(chan interface{})
	p.Subscribe(topic, channel)

	for message := range channel {
		err = c.WriteJSON(message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (p *InMemoryPublisher) publish(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("ERROR: Content-Type is %v", contentType)
		http.Error(w, "Please confirm your Content-Type. Only applicat/json is currently supported", http.StatusBadRequest)
	}

	queryParams := r.URL.Query()
	topic := queryParams.Get("topic")
	if topic == "" {
		log.Println("ERROR: Topic is empty")
		http.Error(w, "Please include the topic query parameter", http.StatusBadRequest)
		return
	}

	if r.ContentLength == 0 {
		log.Println("ERROR: ContentLength is 0 length")
		http.Error(w, "Please include a request body", http.StatusBadRequest)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Error reading body: %v", err)
		http.Error(w, "Unable to read message body", http.StatusInternalServerError)
	}

	var message interface{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		log.Printf("ERROR: Error unmarshalling message: %v", err)
		http.Error(w, "Please ensure your message body is properly formatted JSON", http.StatusBadRequest)
	}

	// publish here
	p.Publish(topic, message)
}

func main() {
	p := NewInMemoryPublisher()

	flag.Parse()
	log.SetFlags(0)

	r := mux.NewRouter()
	r.HandleFunc("/subscribe", p.subscribe)
	r.HandleFunc("/publish", p.publish)
	http.Handle("/", r)

	log.Println(("Listening on localhost:8081"))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
