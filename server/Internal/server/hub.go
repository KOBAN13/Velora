package server

import (
	"Velora/server/Internal"
	"Velora/server/Internal/objects"
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
	Generator *Internal.IdGenerator

	Clients *objects.SharedCollection[ClientInterface]

	Broadcast chan *packets.Packet

	Register chan ClientInterface

	Unregister chan ClientInterface
}

func NewHub() *Hub {
	var idGenerator = &Internal.IdGenerator{}
	var clients = objects.NewSharedCollection[ClientInterface](idGenerator)

	return &Hub{
		Generator:  idGenerator,
		Clients:    clients,
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
			var id = h.Clients.Add(client, h.Generator)
			client.Initialize(id)

		case client := <-h.Unregister:
			log.Println("unregister client")
			h.Clients.Remove(client.Id())

		case packet := <-h.Broadcast:
			log.Println("broadcast packet")

			h.Clients.Foreach(func(clientInterface ClientInterface, id uint64) {
				if id != packet.SenderId {
					clientInterface.ProcessPacket(id, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(getNewClient func(hub *Hub, writer http.ResponseWriter, request *http.Request) (ClientInterface, error), writer http.ResponseWriter, request *http.Request) {
	log.Println("New client connected from ", request.RemoteAddr)

	var client, err = getNewClient(h, writer, request)

	if err != nil {
		log.Printf("Error getting new client: %v", err)
		return
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
