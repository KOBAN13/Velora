package server

import (
	"Velora/server/Internal"
	"Velora/server/Internal/objects"
	"Velora/server/Internal/server/contracts"
	"Velora/server/Internal/server/db"
	"Velora/server/Internal/server/lobby"
	"Velora/server/pkg/packets"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ contracts.Hub = (*Hub)(nil)

type Hub struct {
	Generator *Internal.IdGenerator

	Clients *objects.SharedCollection[contracts.Client]

	Lobby *lobby.LobbyManager

	Broadcast chan *packets.Packet

	Register chan contracts.Client

	Unregister chan contracts.Client

	DbPool *pgxpool.Pool
}

func NewHub() *Hub {
	var idGenerator = &Internal.IdGenerator{}
	var clients = objects.NewSharedCollection[contracts.Client]()

	var connect, errConnect = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if errConnect != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", errConnect)
		os.Exit(1)
	}

	return &Hub{
		Generator:  idGenerator,
		Clients:    clients,
		Broadcast:  make(chan *packets.Packet),
		Register:   make(chan contracts.Client),
		Unregister: make(chan contracts.Client),
		DbPool:     connect,
		Lobby:      lobby.NewLobbyManager(),
	}
}

func (h *Hub) NewDbTx() *db.DbTx {
	return &db.DbTx{
		Ctx:            context.Background(),
		UserRepository: db.NewUserRepository(h.DbPool),
	}
}

func (h *Hub) Client(id uint64) (contracts.Client, bool) {
	return h.Clients.Get(id)
}

func (h *Hub) BroadcastFrom(senderID uint64, msg packets.Msg) {
	h.Broadcast <- &packets.Packet{SenderId: senderID, Msg: msg}
}

func (h *Hub) UnregisterClient(client contracts.Client) {
	h.Unregister <- client
}

func (h *Hub) Run() {
	log.Println("Initializing database connection")

	db.TestPostgresConnection()

	log.Println("Hub is running")

	for {
		select {
		case client := <-h.Register:
			log.Println("register client")
			var id = h.Clients.Add(client, h.Generator)
			client.Initialize(id)

		case client := <-h.Unregister:
			log.Println("unregister client")
			if err := h.Lobby.RemoveClient(client); err != nil {
				log.Printf("failed to remove client from lobby: %v", err)
			}
			h.Clients.Remove(client.Id())

		case packet := <-h.Broadcast:
			log.Println("broadcast packet")

			h.Clients.Foreach(func(clientInterface contracts.Client, id uint64) {
				if id != packet.SenderId {
					clientInterface.ProcessPacket(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(getNewClient func(hub contracts.Hub, writer http.ResponseWriter, request *http.Request) (contracts.Client, error), writer http.ResponseWriter, request *http.Request) {
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
