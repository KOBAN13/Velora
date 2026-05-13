package server

import (
	"Velora/server/Internal"
	"Velora/server/Internal/objects"
	"Velora/server/Internal/server/db"
	"Velora/server/pkg/packets"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DbTx struct {
	Ctx            context.Context
	UserRepository *db.UserRepository
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:            context.Background(),
		UserRepository: db.NewUserRepository(h.dbPool),
	}
}

type ClientStateHandler interface {
	Name() string

	SetClientInterface(client ClientInterface)

	HandleMessage(id uint64, msg packets.Msg)

	OnEnter()
	OnLeave()
}

type ClientInterface interface {
	SetState(newState ClientStateHandler)

	Initialize(id uint64)
	Id() uint64
	ProcessPacket(id uint64, msg packets.Msg)

	SetUser(user *db.User)
	GetUser() *db.User
	IsAuthenticated() bool

	SocketSend(message packets.Msg)
	SocketSendAs(message packets.Msg, id uint64)
	PassToPear(message packets.Msg, id uint64)
	Broadcast(message packets.Msg)

	WritePump()
	ReadPump()

	DbTx() *DbTx

	Close(reason string)
}

type Hub struct {
	Generator *Internal.IdGenerator

	Clients *objects.SharedCollection[ClientInterface]

	Broadcast chan *packets.Packet

	Register chan ClientInterface

	Unregister chan ClientInterface

	dbPool *pgxpool.Pool
}

func NewHub() *Hub {
	var idGenerator = &Internal.IdGenerator{}
	var clients = objects.NewSharedCollection[ClientInterface](idGenerator)

	var connect, errConnect = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if errConnect != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", errConnect)
		os.Exit(1)
	}

	return &Hub{
		Generator:  idGenerator,
		Clients:    clients,
		Broadcast:  make(chan *packets.Packet),
		Register:   make(chan ClientInterface),
		Unregister: make(chan ClientInterface),
		dbPool:     connect,
	}
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
			h.Clients.Remove(client.Id())

		case packet := <-h.Broadcast:
			log.Println("broadcast packet")

			h.Clients.Foreach(func(clientInterface ClientInterface, id uint64) {
				if id != packet.SenderId {
					clientInterface.ProcessPacket(packet.SenderId, packet.Msg)
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
