package states

import (
	"Velora/server/Internal/server"
	"Velora/server/pkg/packets"
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

}

func (auth *Authenticated) HandleMessage(id uint64, msg packets.Msg) {

}

func (auth *Authenticated) OnEnter() {

}
func (auth *Authenticated) OnLeave() {

}
