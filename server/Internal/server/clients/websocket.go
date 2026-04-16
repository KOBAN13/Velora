package clients

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	ReadBufferSize     = 1024
	WriteBufferSize    = 1024
	SendChanBufferSize = 256
)

type WebSocketClient struct {
	id       uint64
	conn     *websocket.Conn
	hub      *server.Hub
	sendChan chan *packets.Packet
	logger   *log.Logger
}

func (c *WebSocketClient) NewWebsocketConnection(hub *server.Hub, writer *http.ResponseWriter, request *http.Request) (server.ClientInterface, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     checkOrigin,
	}

	var connection, err = upgrader.Upgrade(*writer, request, nil)

	if err != nil {
		return nil, err
	}

	var client = &WebSocketClient{
		conn:     connection,
		hub:      hub,
		sendChan: make(chan *packets.Packet, SendChanBufferSize),
		logger:   log.New(*writer, "", log.LstdFlags),
	}

	return client, nil
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id

	c.logger.SetPrefix(fmt.Sprintf("Client ID: %d", c.id))
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) SendPacket(packet *packets.Packet) {

}

func (c *WebSocketClient) ProcessPacket(packet *packets.Packet) {

}

func checkOrigin(request *http.Request) bool {
	return true
}
