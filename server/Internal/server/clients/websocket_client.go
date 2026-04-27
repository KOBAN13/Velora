package clients

import (
	"Velora/server/Internal/server"
	"Velora/server/Internal/server/states"
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

	c.SetState(&states.Connection{})
}

func (c *WebSocketClient) SetState(newState server.ClientStateHandler) {
	var prevStateName = "None"

	if c.state != nil {
		prevStateName = c.state.Name()
		c.state.OnLeave()
	}

	var newStateName = "None"

	if newState != nil {
		newStateName = newState.Name()
	}

	c.logger.Printf("Switch from state : %s, new state: %s", prevStateName, newStateName)

	c.state = newState
	c.state.SetClientInterface(c)
	c.state.OnEnter()
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}
