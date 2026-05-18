package lobby

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
)

func (lobby *LobbyManager) buildSnapshot(room *Room) packets.Msg {
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

func (lobby *LobbyManager) RoomSnapshot(room *Room) {
	packets.Pa
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

func newRoomPlayerMessage(roomPlayer *RoomPlayer) *packets.RoomPlayerMessage {
	return &packets.RoomPlayerMessage{
		UserId:   roomPlayer.UserID,
		ClientId: roomPlayer.ClientID,
		Username: roomPlayer.Username,
		IsReady:  roomPlayer.IsReady,
		Owner:    roomPlayer.IsOwner,
	}
}
