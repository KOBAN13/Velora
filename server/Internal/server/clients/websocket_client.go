package clients

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	id       uint64
	conn     *websocket.Conn
	hub      *server.Hub
	sendChan chan *packets.Packet
	state    server.ClientStateHandler
	logger   *log.Logger
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id

	c.logger.SetPrefix(fmt.Sprintf("Client ID: %d ", c.id))

	var clientId = packets.NewId(id)

	c.SocketSend(clientId)
	c.logger.Printf("Client initialized and send to client id: %v", clientId)
}

func (c *WebSocketClient) SetState(newState server.ClientStateHandler) {
	
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}
