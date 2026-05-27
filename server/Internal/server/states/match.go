package states

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
)

type Match struct {
	client contracts.ClientInterface
	logger *log.Logger
}

func (m *Match) Name() string {
	return "Match"
}

func (m *Match) SetClient(client contracts.ClientInterface) {
	m.client = client

	var loggerPrefix = fmt.Sprintf("Client %d [%s]", client.Id(), m.Name())

	m.logger = log.New(log.Writer(), loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (m *Match) HandleMessage(id uint64, msg packets.Msg) {
	switch msg.(type) {
	case *packets.Packet_Chat:
		m.handleChatMessage(id, msg)
	}
}

func (m *Match) OnEnter() {

}

func (m *Match) OnLeave() {

}

func (m *Match) handleChatMessage(id uint64, msg packets.Msg) {
	if id == m.client.Id() {
		m.client.Broadcast(msg)
		return
	}

	m.client.SocketSendAs(msg, id)
}
