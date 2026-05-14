package lobby

import "Velora/server/pkg/packets"

type roomSocketSender interface {
	SocketSend(message packets.Msg)
}

func buildSnapshot(room *Room) packets.Msg {
	players := make([]*packets.RoomPlayerMessage, 0, len(room.Players))
	addedPlayers := make(map[uint64]struct{}, len(room.Players))

	for _, userId := range room.PlayerOrder {
		player, ok := room.Players[userId]
		if !ok {
			continue
		}

		players = append(players, packets.NewRoomPlayerMessage(player))
		addedPlayers[userId] = struct{}{}
	}

	for userId, player := range room.Players {
		if _, ok := addedPlayers[userId]; ok {
			continue
		}

		players = append(players, packets.NewRoomPlayerMessage(player))
	}

	return packets.NewRoomStateSnapshot(room.ID, room.MaxPlayers, room.Status, players)
}

func (lobby *LobbyManager) broadcastToRoom(room *Room, msg packets.Msg) {
	lobby.mutex.Lock()

	if room == nil {
		lobby.mutex.Unlock()
		return
	}

	if msg == nil {
		msg = buildSnapshot(room)
	}

	clients := roomClients(room)

	lobby.mutex.Unlock()

	for _, client := range clients {
		client.SocketSend(msg)
	}
}

func roomClients(room *Room) []roomSocketSender {
	clients := make([]roomSocketSender, 0, len(room.Players))
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
