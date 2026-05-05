package main

import (
	"Velora/server/Internal/server"
	"Velora/server/Internal/server/clients"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

var (
	port = flag.Int("port", 8080, "The port to listen on")
)

func main() {
	flag.Parse()

	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("Error loading config.env file")
	}

	var hub = server.NewHub()

	http.HandleFunc("/velora", func(writer http.ResponseWriter, reader *http.Request) {
		hub.Serve(clients.NewWebsocketConnection, writer, reader)
	})

	go hub.Run()

	var addr = fmt.Sprintf(":%d", *port)

	var err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Failed start server %v: ", err)
	}
}
