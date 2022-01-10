package main

import (
	"flag"
	"log"

	pubhttp "github.com/datravis/websocket-example/internal/pkg/http"
	"github.com/datravis/websocket-example/internal/pkg/publish"
)

var addr = flag.String("address", "localhost:8081", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	p := publish.NewInMemoryPublisher()
	s := pubhttp.NewPubSubServer(*addr, p)

	log.Fatal(s.Serve())
}
