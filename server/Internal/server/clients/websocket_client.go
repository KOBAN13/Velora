package clients

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/Internal/server/db"
	"Velora/server/Internal/server/states"
	"Velora/server/pkg/packets"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	id       uint64
	conn     *websocket.Conn
	hub      contracts.Hub
	sendChan chan *packets.Packet
	state    contracts.ClientStateHandler
	logger   *log.Logger
	dBtX     *db.DbTx
	user     *db.User
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id

	c.logger.SetPrefix(fmt.Sprintf("Client ID: %d ", c.id))

	c.SetState(&states.Connection{})
}

func (c *WebSocketClient) SetUser(user *db.User) {
	c.user = user
}

func (c *WebSocketClient) GetUser() *db.User {
	return c.user
}

func (c *WebSocketClient) IsAuthenticated() bool {
	return c.state != nil && c.state.Name() == "Authenticated"
}

func (c *WebSocketClient) Lobby() contracts.LobbyService {
	return c.hub.GetLobby()
}

func (c *WebSocketClient) DbTx() *db.DbTx {
	return c.dBtX
}

func (c *WebSocketClient) SetState(newState contracts.ClientStateHandler) {
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

	if c.state != nil {
		c.state.SetClient(c)
		c.state.OnEnter()
	}
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}
