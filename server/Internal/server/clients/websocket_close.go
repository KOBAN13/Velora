package clients

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client: %s", reason)

	c.SetState(nil)

	c.hub.Unregister <- c
	c.conn.Close()

	if _, isClosed := <-c.sendChan; !isClosed {
		close(c.sendChan)
	}
}
