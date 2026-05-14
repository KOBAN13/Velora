package states

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
)

type Authenticated struct {
	client server.ClientInterface
	log    *log.Logger
}

func (auth *Authenticated) Name() string {
	return "Authenticated"
}

func (auth *Authenticated) SetClientInterface(client server.ClientInterface) {
	auth.client = client

	var loggerPrefix = fmt.Sprintf("Client %d [%s]", client.Id(), auth.Name())

	auth.log = log.New(log.Writer(), loggerPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (auth *Authenticated) HandleMessage(id uint64, msg packets.Msg) {

}

func (auth *Authenticated) OnEnter() {

}
func (auth *Authenticated) OnLeave() {

}

func CreateRoomRequestMessage() {
	packets.NewReadyRequest()
}
