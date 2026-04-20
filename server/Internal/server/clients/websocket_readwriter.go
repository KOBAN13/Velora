package clients

import (
	"Velora/server/pkg/packets"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (c *WebSocketClient) SocketSend(message packets.Msg) {
	c.SocketSendAs(message, c.Id())
}

func (c *WebSocketClient) SocketSendAs(message packets.Msg, id uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: id, Msg: message}:
	default:
		c.logger.Printf("Send channel full, drop packet")
	}
}

func (c *WebSocketClient) WritePump() {
	defer func() {
		c.logger.Println("Closing write pump")
		c.Close("write pump closed")
	}()

	for packet := range c.sendChan {
		var writer, err = c.conn.NextWriter(websocket.BinaryMessage)

		if err != nil {
			c.logger.Printf("error writing packet %T, closing client: %v", packet.Msg, err)
			return
		}

		data, err := proto.Marshal(packet)

		if err != nil {
			c.logger.Printf("error marshalling packet %T, closing client: %v", packet.Msg, err)
			return
		}

		_, err = writer.Write(data)

		if err != nil {
			c.logger.Printf("error writing packet %T, closing client: %v", packet.Msg, err)
			return
		}

		if err := writer.Close(); err != nil {
			c.logger.Printf("error closing client writer: %v", err)
		}
	}
}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("Error: %v", err)
			}
			break
		}

		var packet packets.Packet

		err = proto.Unmarshal(data, &packet)

		if err != nil {
			c.logger.Printf("Error: %v", err)
			continue
		}

		if packet.SenderId == 0 {
			packet.SenderId = c.Id()
		}

		c.ProcessPacket(packet.SenderId, packet.Msg)
	}
}
