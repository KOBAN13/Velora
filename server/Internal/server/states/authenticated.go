package states

import (
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
	"fmt"
	"log"
)

type Authenticated struct {
	client contracts.ClientInterface
	log    *log.Logger
}

func (auth *Authenticated) Name() string {
	return "Authenticated"
}

func (auth *Authenticated) SetClient(client contracts.ClientInterface) {
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
