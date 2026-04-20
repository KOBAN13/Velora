package clients

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const (
	ReadBufferSize     = 1024
	WriteBufferSize    = 1024
	SendChanBufferSize = 256
)

func NewWebsocketConnection(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterface, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin:     checkOrigin,
	}

	var connection, err = upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return nil, err
	}

	var client = &WebSocketClient{
		conn:     connection,
		hub:      hub,
		sendChan: make(chan *packets.Packet, SendChanBufferSize),
		logger:   log.New(os.Stderr, "", log.LstdFlags),
	}

	return client, nil
}

func checkOrigin(request *http.Request) bool {
	return true
}
