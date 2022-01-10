package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	pubhttp "github.com/datravis/websocket-example/internal/pkg/http"
)

var addr = flag.String("address", "localhost:8081", "http service address")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [topic]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	flag.Parse()
	log.SetFlags(0)
	topic := os.Args[1]

	client := pubhttp.NewPubSubClient(*addr, topic)
	log.Fatal(client.Run())
}
