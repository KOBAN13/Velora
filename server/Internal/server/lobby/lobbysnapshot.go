package lobby

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
)

func (lobby *LobbyManager) buildSnapshot(room *Room) packets.Msg {
	var players = collectRoomPlayers(room)

	return packets.NewRoomStateSnapshot(room.ID, room.MaxPlayers, room.Status, players)
}

func (lobby *LobbyManager) broadcastToRoom(room *Room, msg packets.Msg) {
	lobby.mutex.Lock()

	clients := roomClients(room)

	lobby.mutex.Unlock()

	for _, client := range clients {
		client.SocketSend(msg)
	}
}

func (lobby *LobbyManager) RoomListSnapshot() packets.Msg {
	var messages = make([]*packets.RoomSummaryMessage, 0, lobby.rooms.Size())

	lobby.rooms.Foreach(func(room *Room, u uint64) {
		var players = collectRoomPlayers(room)
		var roomSummary = packets.NewRoomSummaryMessage(room.ID, room.Name, players, room.MaxPlayers, room.Status)

		messages = append(messages, roomSummary)
	})

	return packets.NewRoomListSnapshotMessage(messages)
}

func (lobby *LobbyManager) PlayersListSnapshot(id uint64) packets.Msg {
	var room, _ = lobby.rooms.Get(id)
	var players = collectRoomPlayers(room)

	return packets.NewPlayersInRoomResponse(players)
}

func roomClients(room *Room) []contracts.ClientInterface {
	clients := make([]contracts.ClientInterface, 0, len(room.Players))
	addedClients := make(map[uint64]struct{}, len(room.Players))

	for _, userId := range room.PlayerOrder {
		player, ok := room.Players[userId]
		if !ok || player.Client == nil {
			continue
		}

		clients = append(clients, player.Client)
		addedClients[userId] = struct{}{}
	}

	for userId, player := range room.Players {
		if _, ok := addedClients[userId]; ok || player.Client == nil {
			continue
		}

		clients = append(clients, player.Client)
	}

	return clients
}

func collectRoomPlayers(room *Room) []*packets.RoomPlayerMessage {
	players := make([]*packets.RoomPlayerMessage, 0, len(room.Players))
	addedPlayers := make(map[uint64]struct{}, len(room.Players))

	for _, userId := range room.PlayerOrder {
		player, ok := room.Players[userId]
		if !ok {
			continue
		}

		players = append(players, newRoomPlayerMessage(player))
		addedPlayers[userId] = struct{}{}
	}

	for userId, player := range room.Players {
		if _, ok := addedPlayers[userId]; ok {
			continue
		}

		players = append(players, newRoomPlayerMessage(player))
	}

	return players
}

func newRoomPlayerMessage(roomPlayer *RoomPlayer) *packets.RoomPlayerMessage {
	return &packets.RoomPlayerMessage{
		UserId:   roomPlayer.UserID,
		ClientId: roomPlayer.ClientID,
		Username: roomPlayer.Username,
		IsReady:  roomPlayer.IsReady,
		Owner:    roomPlayer.IsOwner,
	}
}
