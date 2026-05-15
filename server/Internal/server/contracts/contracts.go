package contracts

import (
	"Velora/server/Internal/server/db"
	"Velora/server/pkg/packets"
)

type ClientStateHandler interface {
	Name() string
	SetClient(client ClientInterface)
	HandleMessage(id uint64, msg packets.Msg)
	OnEnter()
	OnLeave()
}

type ClientInterface interface {
	SetState(newState ClientStateHandler)

	Id() uint64

	SetUser(user *db.User)
	SocketSend(message packets.Msg)
	SocketSendAs(message packets.Msg, id uint64)
	Broadcast(message packets.Msg)

	DbTx() *db.DbTx
}

type LobbyClient interface {
	Id() uint64

	GetUser() *db.User
	IsAuthenticated() bool

	SocketSend(message packets.Msg)
}

type Client interface {
	ClientInterface
	LobbyClient

	Initialize(id uint64)
	ProcessPacket(id uint64, msg packets.Msg)

	WritePump()
	ReadPump()

	Close(reason string)
}

type Hub interface {
	NewDbTx() *db.DbTx
	Client(id uint64) (Client, bool)
	BroadcastFrom(senderID uint64, msg packets.Msg)
	UnregisterClient(client Client)
	RemoveClient(client Client)
}
