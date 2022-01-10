package http

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type PubSubClient struct {
	address string
	topic   string
}

func NewPubSubClient(address, topic string) *PubSubClient {
	return &PubSubClient{
		address,
		topic,
	}
}

func (client *PubSubClient) Run() error {
	// define our connection URL
	u := url.URL{Scheme: "ws", Host: client.address, Path: "/subscribe"}
	q := u.Query()
	q.Set("topic", client.topic)
	u.RawQuery = q.Encode()

	// connect to the server
	log.Printf("Connecting to %s", u.String())
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == websocket.ErrBadHandshake {
		return fmt.Errorf("handshake failed with status %d", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("unable to dial server: %v", err)
	}
	defer c.Close()

	go func() {
		for {
			// TODO: there's a race condition on shutdown, where the connection is sometimes used here after
			// it's closed
			_, message, err := c.ReadMessage()
			if err != nil {

				log.Println("Error: error reading from socket: ", err)
				return
			}
			log.Printf("Received message [topic=%s]: %s", client.topic, message)
		}
	}()

	// wait for a sigint
	wait()

	// Send a close message to properly close our connection
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("error closing connection: %v", err)
	}
	return nil
}

// a helper that waits for a sigint from the OS
func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	// wait for sigint
	<-interrupt
	log.Println("Received interrupt, shutting down")
}
