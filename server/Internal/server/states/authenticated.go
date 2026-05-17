package states

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
)

type Authenticated struct {
	client contracts.ClientInterface
	logger *log.Logger
}

func (auth *Authenticated) Name() string {
	return "Authenticated"
}

func (auth *Authenticated) SetClient(client contracts.ClientInterface) {
	auth.client = client

	var loggerPrefix = fmt.Sprintf("Client %d [%s]", client.Id(), auth.Name())

	auth.logger = log.New(log.Writer(), loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (auth *Authenticated) HandleMessage(id uint64, msg packets.Msg) {
	auth.logger.Printf("Handle message id=%d msg=%s", id, msg)

	switch message := msg.(type) {
	case *packets.Packet_CreateRoomRequest:
		auth.createRoomRequestMessage(message.CreateRoomRequest.MaxPlayer)

	case *packets.Packet_JoinRoomRequest:
		auth.joinRoomRequestMessage(message.JoinRoomRequest.RoomId)

	case *packets.Packet_LeaveRoomRequest:
		auth.leaveRoomRequestMessage()

	case *packets.Packet_ReadyRequest:
		auth.readyRequestMessage(message.ReadyRequest.IsReady)

	case *packets.Packet_StartGame:
		auth.startGameRequestMessage()
	}
}

func (auth *Authenticated) OnEnter() {
	var id = auth.client.Id()

	var clientId = packets.NewId(id)

	auth.client.SocketSend(clientId)
	auth.logger.Printf("Client authenticated and send to client id: %v", clientId)
}

func (auth *Authenticated) OnLeave() {

}

func (auth *Authenticated) createRoomRequestMessage(maxPlayers uint32) {
	var lobbyService = auth.client.Lobby()

	var err = lobbyService.CreateRoom(auth.client, maxPlayers)

	if err != nil {
		var dennyMessage = packets.NewDenyResponse(err.Error())

		auth.logger.Printf("Error create room: %v", err)
		auth.client.SocketSend(dennyMessage)
	}

	auth.client.SocketSend(packets.NewOkResponse())
}

func (auth *Authenticated) joinRoomRequestMessage(roomId uint64) {
	var lobbyService = auth.client.Lobby()

	var err = lobbyService.JoinRoom(auth.client, roomId)

	if err != nil {
		var dennyMessage = packets.NewDenyResponse(err.Error())

		auth.logger.Printf("Error join room: %v", err)
		auth.client.SocketSend(dennyMessage)
	}

	auth.client.SocketSend(packets.NewOkResponse())
}

func (auth *Authenticated) leaveRoomRequestMessage() {
	var lobbyService = auth.client.Lobby()

	var err = lobbyService.LeaveRoom(auth.client)

	if err != nil {
		var dennyMessage = packets.NewDenyResponse(err.Error())

		auth.logger.Printf("Error leave room: %v", err)
		auth.client.SocketSend(dennyMessage)
	}

	auth.client.SocketSend(packets.NewOkResponse())
}

func (auth *Authenticated) readyRequestMessage(isReady bool) {
	var lobbyService = auth.client.Lobby()

	var err = lobbyService.SetReady(auth.client, isReady)

	if err != nil {
		var dennyMessage = packets.NewDenyResponse(err.Error())

		auth.logger.Printf("Error ready request room: %v", err)
		auth.client.SocketSend(dennyMessage)
	}

	auth.client.SocketSend(packets.NewOkResponse())
}

func (auth *Authenticated) startGameRequestMessage() {
	var lobbyService = auth.client.Lobby()

	var err = lobbyService.StartGame(auth.client)

	if err != nil {
		var dennyMessage = packets.NewDenyResponse(err.Error())

		auth.logger.Printf("Error start game: %v", err)
		auth.client.SocketSend(dennyMessage)
	}

	auth.client.SocketSend(packets.NewOkResponse())
}
