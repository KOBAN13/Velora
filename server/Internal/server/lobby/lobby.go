package lobby

import (
	"Velora/server/Internal"
	"Velora/server/Internal/objects"
	"Velora/server/Internal/server/contracts"
	"Velora/server/Internal/server/match"
	"Velora/server/pkg/packets"
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

var (
	ErrUserIsNotAuthenticated = errors.New("user is not authenticated")
	ErrUserIsOwner            = errors.New("user is owner this room and action is not allowed")
	ErrMaxPlayersExceeded     = errors.New("max players exceeded")
	ErrUserInRoom             = errors.New("user is in room")
	ErrRoomNotFound           = errors.New("room not found")
	ErrRoomIsFull             = errors.New("room is full")
	ErrRoomIsNotJoinable      = errors.New("room is not joinable")
	ErrUserIsNotRoom          = errors.New("user is not a room")
	ErrOwnerIsNotStartGame    = errors.New("owner is not starting game")
	ErrUserInNotOwnerGame     = errors.New("user is not owner game")
)

func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		rooms:            objects.NewSharedCollection[*Room](),
		userRoom:         make(map[uint64]uint64),
		matchIdGenerator: &Internal.IdGenerator{},
	}
}

type LobbyManager struct {
	mutex            sync.Mutex
	rooms            *objects.SharedCollection[*Room]
	userRoom         map[uint64]uint64
	matchIdGenerator *Internal.IdGenerator

	roomListChanged func()
}

type Room struct {
	Name        string
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
	Client   contracts.ClientInterface
}

func (lobby *LobbyManager) GetRoomList() *objects.SharedCollection[*Room] {
	return lobby.rooms
}

func (lobby *LobbyManager) StartGame(client contracts.ClientInterface) error {
	var roomId, room, player, err = lobby.validateRoom(client)

	if err != nil {
		return err
	}

	if !player.IsOwner {
		return ErrUserInNotOwnerGame
	}

	if room.Status != packets.RoomStatus_ROOM_STATUS_WAITING || !allPLayersReady(room) {
		return ErrOwnerIsNotStartGame
	}

	var matchId = lobby.matchIdGenerator.Next()
	var mapSeed = rand.Uint64()
	var startsAtUnixMs = time.Now().UnixMilli()

	var slot = uint32(0)

	var matchService = client.GetMatches()

	var matchConfig = match.MatchConfig{
		RoomId:  roomId,
		MatchId: matchId,
		MapSeed: mapSeed,
		Players: make([]match.PlayerRef, 0, len(room.Players)),
	}

	for _, userId := range room.PlayerOrder {
		var roomPlayer, ok = room.Players[userId]

		if !ok {
			continue
		}

		matchConfig.Players = append(matchConfig.Players, match.PlayerRef{
			UserId:   roomPlayer.UserID,
			ClientId: roomPlayer.ClientID,
			Client:   roomPlayer.Client,
			Slot:     slot,
		})

		slot++
	}

	var rollbackMatchStart = func(startErr error) error {
		room.Status = packets.RoomStatus_ROOM_STATUS_WAITING

		var msg = lobby.buildSnapshot(room)
		lobby.broadcastToRoom(room, msg)

		return startErr
	}

	var createMatch, errCreateMatch = matchService.CreateMatch(matchConfig)

	if errCreateMatch != nil {
		return rollbackMatchStart(errCreateMatch)
	}

	var errInitializeSystems = createMatch.InitializeSystems(time.Now())

	if errInitializeSystems != nil {
		matchService.StopMatch(matchId)
		return rollbackMatchStart(errInitializeSystems)
	}

	room.Status = packets.RoomStatus_ROOM_STATUS_STARTED

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	lobby.notifyRoomListChanged()

	for _, playerRef := range matchConfig.Players {
		var roomPlayer, ok = room.Players[playerRef.UserId]

		if !ok || roomPlayer.Client == nil {
			continue
		}

		var matchStartedMsg = packets.NewMatchStarted(roomId, matchId, roomPlayer.UserID, playerRef.Slot, mapSeed, startsAtUnixMs)

		roomPlayer.Client.SocketSend(matchStartedMsg)
	}

	go createMatch.Run()

	return nil
}

func (lobby *LobbyManager) CreateRoom(client contracts.ClientInterface, roomName string, maxPlayers uint32) error {
	var roomPlayer, err = lobby.newRoomPlayer(client)
	if err != nil {
		return err
	}

	if maxPlayers == 0 {
		maxPlayers = 2
	}

	if maxPlayers > 5 {
		return ErrMaxPlayersExceeded
	}

	var room = &Room{
		Name:       roomName,
		MaxPlayers: maxPlayers,
		Status:     packets.RoomStatus_ROOM_STATUS_WAITING,
		Players:    make(map[uint64]*RoomPlayer),
	}

	var roomId = lobby.rooms.Add(room)
	room.ID = roomId

	lobby.addPlayerToRoom(roomId, room, roomPlayer)

	var msg = lobby.buildSnapshot(room)

	client.SocketSend(msg)

	lobby.notifyRoomListChanged()

	return nil
}

func (lobby *LobbyManager) JoinRoom(client contracts.ClientInterface, roomId uint64) error {
	var roomPlayer, err = lobby.newRoomPlayer(client)
	if err != nil {
		return err
	}

	var room, isRoomFind = lobby.rooms.Get(roomId)
	if !isRoomFind {
		return ErrRoomNotFound
	}

	if _, ok := room.Players[roomPlayer.UserID]; ok {
		return ErrUserInRoom
	}

	if room.Status != packets.RoomStatus_ROOM_STATUS_WAITING {
		return ErrRoomIsNotJoinable
	}

	if uint32(len(room.Players)) == room.MaxPlayers {
		return ErrRoomIsFull
	}

	lobby.addPlayerToRoom(roomId, room, roomPlayer)

	var playerRoom = newRoomPlayerMessage(roomPlayer)

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	var msgPlayerJoin = packets.NewPlayerJoinInRoom(playerRoom)
	lobby.broadcastToRoom(room, msgPlayerJoin)

	lobby.notifyRoomListChanged()

	return nil
}

func (lobby *LobbyManager) KickPlayerInRoom(client contracts.ClientInterface, idPlayerKick uint64) error {
	var roomId, room, player, err = lobby.validateRoom(client)

	if err != nil {
		return err
	}

	if !player.IsOwner {
		return ErrUserInNotOwnerGame
	}

	if player.UserID == idPlayerKick {
		return ErrUserIsOwner
	}

	if !lobby.removePlayerFromRoom(roomId, room, idPlayerKick) {
		return ErrUserIsNotRoom
	}

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	var msgPlayerLeave = packets.NewPlayerKickInRoomResponse(idPlayerKick)
	lobby.broadcastToRoom(room, msgPlayerLeave)

	lobby.notifyRoomListChanged()

	return nil
}

func (lobby *LobbyManager) LeaveRoom(client contracts.ClientInterface) error {
	var roomId, room, player, err = lobby.validateRoom(client)

	if err != nil {
		return err
	}

	if !lobby.removePlayerFromRoom(roomId, room, player.UserID) {
		return nil
	}

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	var playerRoom = newRoomPlayerMessage(player)

	var msgPlayerLeave = packets.NewPlayerRemoveInRoom(playerRoom)
	lobby.broadcastToRoom(room, msgPlayerLeave)

	lobby.notifyRoomListChanged()

	return nil
}

func (lobby *LobbyManager) SetReady(client contracts.ClientInterface, isReady bool) error {
	var _, room, player, err = lobby.validateRoom(client)

	if err != nil {
		return err
	}

	if room.Status != packets.RoomStatus_ROOM_STATUS_WAITING {
		return ErrRoomIsNotJoinable
	}

	player.IsReady = isReady

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	var message = packets.NewReadyResponse(isReady, player.UserID)

	lobby.broadcastToRoom(room, message)

	return nil
}

func (lobby *LobbyManager) RemoveClient(client contracts.ClientInterface) error {
	if !client.IsAuthenticated() {
		return nil
	}

	var userID = client.GetUser().ID
	var roomId, ok = lobby.userRoom[userID]

	if !ok {
		return nil
	}

	var room, isRoomFound = lobby.rooms.Get(roomId)

	if !isRoomFound {
		delete(lobby.userRoom, userID)
		return nil
	}

	var player, isPlayerInRoom = room.Players[userID]
	if !isPlayerInRoom {
		delete(lobby.userRoom, userID)
		return nil
	}

	if !lobby.removePlayerFromRoom(roomId, room, userID) {
		return nil
	}

	var msg = lobby.buildSnapshot(room)
	lobby.broadcastToRoom(room, msg)

	var playerRoom = newRoomPlayerMessage(player)
	var msgPlayerLeave = packets.NewPlayerRemoveInRoom(playerRoom)
	lobby.broadcastToRoom(room, msgPlayerLeave)

	lobby.notifyRoomListChanged()

	return nil
}

func (lobby *LobbyManager) SetRoomListChangedHandler(handler func()) {
	if handler != nil {
		lobby.roomListChanged = handler
	}
}

func (lobby *LobbyManager) notifyRoomListChanged() {
	lobby.roomListChanged()
}

func (lobby *LobbyManager) validateRoom(client contracts.ClientInterface) (uint64, *Room, *RoomPlayer, error) {
	if !client.IsAuthenticated() {
		return 0, nil, nil, ErrUserIsNotAuthenticated
	}

	var userID = client.GetUser().ID

	var roomId, ok = lobby.userRoom[userID]

	if !ok {
		return 0, nil, nil, ErrUserIsNotRoom
	}

	var room, isRoomFind = lobby.rooms.Get(roomId)

	if !isRoomFind {
		delete(lobby.userRoom, userID)
		return 0, nil, nil, ErrRoomNotFound
	}

	var player, isPlayerInRoom = room.Players[userID]

	if !isPlayerInRoom {
		delete(lobby.userRoom, userID)
		return 0, nil, nil, ErrUserIsNotRoom
	}

	return roomId, room, player, nil
}

func (lobby *LobbyManager) newRoomPlayer(client contracts.ClientInterface) (*RoomPlayer, error) {
	if !client.IsAuthenticated() {
		return nil, ErrUserIsNotAuthenticated
	}

	var user = client.GetUser()

	if _, ok := lobby.userRoom[user.ID]; ok {
		return nil, ErrUserInRoom
	}

	return &RoomPlayer{
		UserID:   user.ID,
		ClientID: client.Id(),
		Username: user.Username,
		Client:   client,
	}, nil
}

func (lobby *LobbyManager) addPlayerToRoom(roomId uint64, room *Room, player *RoomPlayer) {
	room.Players[player.UserID] = player
	room.PlayerOrder = append(room.PlayerOrder, player.UserID)
	lobby.userRoom[player.UserID] = roomId
	syncPlayerOwner(room)
}

func (lobby *LobbyManager) removePlayerFromRoom(roomId uint64, room *Room, userID uint64) bool {
	delete(room.Players, userID)
	delete(lobby.userRoom, userID)

	if len(room.Players) == 0 {
		lobby.rooms.Remove(roomId)
		return false
	}

	syncPlayerOwner(room)

	return true
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

func allPLayersReady(room *Room) bool {
	for _, roomPLayer := range room.Players {
		if !roomPLayer.IsReady {
			return false
		}
	}

	return true
}
