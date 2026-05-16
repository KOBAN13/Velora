package contracts

import (
	"Velora/server/Internal/server/db"
	"Velora/server/pkg/packets"
)

type LobbyService interface {
	CreateRoom(client ClientInterface, maxPlayers uint32) error
	JoinRoom(client ClientInterface, roomId uint64) error
	LeaveRoom(client ClientInterface) error
	SetReady(client ClientInterface, isReady bool) error
	StartGame(client ClientInterface) error
}

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

	GetUser() *db.User
	IsAuthenticated() bool

	Lobby() LobbyService

	DbTx() *db.DbTx
}

type Client interface {
	ClientInterface

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
	GetLobby() LobbyService
}
