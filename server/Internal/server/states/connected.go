package states

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
)

type Connection struct {
	client server.ClientInterface
	logger *log.Logger
}

func (conn *Connection) Name() string {
	return "Connection"
}

func (conn *Connection) SetClientInterface(client server.ClientInterface) {
	conn.client = client

	var loggerPrefix = fmt.Sprintf("Client %d [%s]", client.Id(), conn.Name())

	conn.logger = log.New(log.Writer(), loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (conn *Connection) HandleMessage(id uint64, msg packets.Msg) {
	if id == conn.client.Id() {
		conn.client.Broadcast(msg)
	} else {
		conn.client.SocketSendAs(msg, id)
	}
}

func (conn *Connection) OnEnter() {
	var id = conn.client.Id()

	var clientId = packets.NewId(id)

	conn.client.SocketSend(clientId)
	conn.logger.Printf("Client initialized and send to client id: %v", clientId)
}

func (conn *Connection) OnLeave() {

}
