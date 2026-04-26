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

}

func (conn *Connection) OnEnter() {
	conn.client.SocketSend(packets.NewId(conn.client.Id()))
}

func (conn *Connection) OnLeave() {

}
