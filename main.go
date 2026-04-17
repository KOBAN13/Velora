package main

import (
	"Velora/server/Internal/server"
	"Velora/server/Internal/server/clients"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8080, "The port to listen on")
	hub  = server.NewHub()
)

func main() {
	flag.Parse()

	http.HandleFunc("/velora", handler)

	go hub.Run()

	var addr = fmt.Sprintf(":%d", *port)

	var err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Failed start server %v: ", err)
	}
}

func handler(writer http.ResponseWriter, reader *http.Request) {
	hub.Serve(clients.NewWebsocketConnection, writer, reader)
}
