package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8081", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func subscribe(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: Unable to upgrade connection: %v", err)
		return
	}
	defer c.Close()

	//do subscribing
	for {
		// read from channel here
	}
}

func publish(w http.ResponseWriter, r *http.Request) {
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
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/subscribe", subscribe)
	http.HandleFunc("/publish", publish)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
