package lobby

import (
	"Velora/server/Internal"
	"Velora/server/Internal/objects"
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"errors"
	"sync"
)

var (
	ErrUserIsNotAuthenticated = errors.New("user is not authenticated")
	ErrMaxPlayersExceeded     = errors.New("max players exceeded")
	ErrUserInRoom             = errors.New("user is in room")
	ErrRoomNotFound           = errors.New("room not found")
	ErrRoomIsFull             = errors.New("room is full")
	ErrRoomIsNotJoinable      = errors.New("room is not joinable")
	ErrUserIsNotRoom          = errors.New("user is not a room")
)

type LobbyManager struct {
	mutex              sync.Mutex
	rooms              *objects.SharedCollection[*Room]
	userRoom           map[uint64]uint64
	roomIdGenerator    *Internal.IdGenerator
	matchIdGenerator   *Internal.IdGenerator
	playersIdGenerator *Internal.IdGenerator
}

type Room struct {
	ID          uint64
	MaxPlayers  uint32
	Status      packets.RoomStatus
	Players     map[uint64]*RoomPlayer
	PlayerOrder []uint64
}

type RoomPlayer struct {
	UserID   uint64
	ClientID uint64
	Username string
	IsReady  bool
	IsOwner  bool
	Client   server.ClientInterface
}

func (lobby *LobbyManager) CreateRoom(client server.ClientInterface, maxPlayers uint32) error {
	if !client.IsAuthenticated() {
		return ErrUserIsNotAuthenticated
	}

	if maxPlayers == 0 {
		maxPlayers = 2
	}

	if maxPlayers > 5 {
		return ErrMaxPlayersExceeded
	}

	var roomId = lobby.roomIdGenerator.Next()

	var roomPlayer = &RoomPlayer{
		UserID:   client.GetUser().ID,
		ClientID: client.Id(),
		Username: client.GetUser().Username,
		IsReady:  false,
		IsOwner:  true,
		Client:   client,
	}

	var room = &Room{
		ID:          roomId,
		MaxPlayers:  maxPlayers,
		Status:      packets.RoomStatus_ROOM_STATUS_WAITING,
		Players:     make(map[uint64]*RoomPlayer),
		PlayerOrder: []uint64{roomPlayer.UserID},
	}

	room.Players[roomPlayer.UserID] = roomPlayer
	syncPlayerOwner(room)

	lobby.userRoom[roomPlayer.UserID] = roomId

	//TODO:
	//Отправлять снепшот

	return nil
}

func (lobby *LobbyManager) JoinRoom(client server.ClientInterface, roomId uint64) error {
	if !client.IsAuthenticated() {
		return ErrUserIsNotAuthenticated
	}

	if _, isRoomFind := lobby.rooms.Get(roomId); !isRoomFind {
		return ErrRoomNotFound
	}

	var room, _ = lobby.rooms.Get(roomId)

	if _, ok := room.Players[client.GetUser().ID]; ok {
		return ErrUserInRoom
	}

	if room.Status != packets.RoomStatus_ROOM_STATUS_WAITING {
		return ErrRoomIsNotJoinable
	}

	if uint32(len(room.Players)) == room.MaxPlayers {
		return ErrRoomIsFull
	}

	var roomPlayer = &RoomPlayer{
		UserID:   client.GetUser().ID,
		ClientID: client.Id(),
		Username: client.GetUser().Username,
		IsReady:  false,
		IsOwner:  false,
		Client:   client,
	}

	room.Players[roomPlayer.UserID] = roomPlayer
	room.PlayerOrder = append(room.PlayerOrder, roomPlayer.UserID)
	syncPlayerOwner(room)

	lobby.userRoom[roomPlayer.UserID] = roomId

	//TODO:
	//Отправлять снепшот

	return nil
}

func (lobby *LobbyManager) LeaveRoom(client server.ClientInterface) error {
	var roomId = lobby.userRoom[client.GetUser().ID]

	if _, isRoomFind := lobby.rooms.Get(roomId); !isRoomFind {
		return ErrRoomNotFound
	}

	var room, _ = lobby.rooms.Get(roomId)

	if _, ok := room.Players[client.GetUser().ID]; !ok {
		return ErrUserIsNotRoom
	}

	delete(room.Players, client.GetUser().ID)
	delete(lobby.userRoom, client.GetUser().ID)

	if len(room.Players) == 0 {
		lobby.rooms.Remove(roomId)
		return nil
	}

	syncPlayerOwner(room)

	//TODO:
	//Отправлять снепшот

	return nil
}

func (lobby *LobbyManager) SetReady(client server.ClientInterface, isReady bool) error {
	var roomId = lobby.userRoom[client.GetUser().ID]

	if _, isRoomFind := lobby.rooms.Get(roomId); !isRoomFind {
		return ErrRoomNotFound
	}

	var room, _ = lobby.rooms.Get(roomId)

	if _, ok := room.Players[client.GetUser().ID]; !ok {
		return ErrUserIsNotRoom
	}

	if room.Status != packets.RoomStatus_ROOM_STATUS_WAITING {
		return ErrRoomIsNotJoinable
	}

	room.Players[client.GetUser().ID].IsReady = isReady

	//TODO:
	//Отправлять снепшот
	//Проверить условие старта матча

	return nil
}

func (lobby *LobbyManager) RemoveClient(client server.ClientInterface) error {
	if !client.IsAuthenticated() {
		return nil
	}

	var userId = client.GetUser().ID
	var roomId, ok = lobby.userRoom[userId]

	if !ok {
		return nil
	}

	var room, isRoomFind = lobby.rooms.Get(roomId)

	if !isRoomFind {
		delete(lobby.userRoom, userId)
		return nil
	}

	delete(room.Players, userId)
	delete(lobby.userRoom, userId)

	if len(room.Players) == 0 {
		lobby.rooms.Remove(roomId)
		return nil
	}

	syncPlayerOwner(room)

	return nil
}

func syncPlayerOwner(room *Room) {
	var nextOwner uint64
	var ownerFound bool
	var playerOrder []uint64

	for _, userId := range room.PlayerOrder {
		if _, ok := room.Players[userId]; !ok {
			continue
		}

		if !ownerFound {
			nextOwner = userId
			ownerFound = true
		}

		playerOrder = append(playerOrder, userId)
	}

	room.PlayerOrder = playerOrder

	for userId, player := range room.Players {
		player.IsOwner = userId == nextOwner
	}
}
