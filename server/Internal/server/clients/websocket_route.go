package clients

import "Velora/server/pkg/packets"

func (c *WebSocketClient) PassToPear(message packets.Msg, id uint64) {
	if peer, exists := c.hub.Clients[id]; exists {
		peer.ProcessPacket(id, message)
	}
}

func (c *WebSocketClient) Broadcast(message packets.Msg) {
	c.hub.Broadcast <- &packets.Packet{SenderId: c.id, Msg: message}
}

func (c *WebSocketClient) ProcessPacket(id uint64, msg packets.Msg) {

}
