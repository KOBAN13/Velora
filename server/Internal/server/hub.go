package server

import (
	"Velora/server/pkg/packets"
	"log"
	"net/http"
)

type ClientInterface interface {
	Initialize(id uint64)
	Id() uint64
	ProcessPacket(id uint64, msg packets.Msg)

	SocketSend(message packets.Msg)
	SocketSendAs(message packets.Msg, id uint64)
	PassToPear(message packets.Msg, id uint64)
	Broadcast(message packets.Msg)

	WritePump()
	ReadPump()

	Close(reason string)
}

type Hub struct {
	Generator IdGenerator

	Clients map[uint64]ClientInterface

	Broadcast chan *packets.Packet

	Register chan ClientInterface

	Unregister chan ClientInterface
}

func NewHub() *Hub {
	return &Hub{
		Generator:  IdGenerator{},
		Clients:    make(map[uint64]ClientInterface),
		Broadcast:  make(chan *packets.Packet),
		Register:   make(chan ClientInterface),
		Unregister: make(chan ClientInterface),
	}
}

func (h *Hub) Run() {
	log.Println("Hub is running")

	for {
		select {
		case client := <-h.Register:
			log.Println("register client")
			var id = h.Generator.Next()
			client.Initialize(id)
			h.Clients[id] = client

		case client := <-h.Unregister:
			log.Println("unregister client")
			delete(h.Clients, client.Id())

		case packet := <-h.Broadcast:
			log.Println("broadcast packet")

			for id, client := range h.Clients {
				if id != packet.SenderId {
					client.ProcessPacket(id, packet.Msg)
				}
			}
		}
	}
}

func (h *Hub) Serve(getNewClient func(writer *http.ResponseWriter, request *http.Request) (ClientInterface, error), writer *http.ResponseWriter, request *http.Request) {
	log.Println("New client connected from ", request.RemoteAddr)

	var client, err = getNewClient(writer, request)

	if err != nil {
		log.Printf("Error getting new client: %v", err)
		return
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
