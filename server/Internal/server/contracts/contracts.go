package contracts

import (
	"Velora/server/Internal/server/db"
	"Velora/server/Internal/server/match"
	"Velora/server/pkg/packets"
)

type MatchService interface {
	CreateMatch(config match.MatchConfig) (*match.Match, error)
	HandleInput(client match.Client, input *packets.PlayerInputMessage) error
	RemoveClient(client match.Client)
	StopMatch(matchId uint64)
}

type LobbyService interface {
	RoomListSnapshot() packets.Msg
	PlayersListSnapshot(id uint64) packets.Msg

	CreateRoom(client ClientInterface, roomName string, maxPlayers uint32) error
	JoinRoom(client ClientInterface, roomId uint64) error
	LeaveRoom(client ClientInterface) error
	SetReady(client ClientInterface, isReady bool) error
	StartGame(client ClientInterface) error
	KickPlayerInRoom(client ClientInterface, idPlayerKick uint64) error
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
	GetMatches() MatchService

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
	GetMatches() MatchService
}
