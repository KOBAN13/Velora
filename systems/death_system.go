package systems

import (
	"Velora/esc"
)

type DeathSystem struct {
}

func NewDeathSystem() *DeathSystem {
	return &DeathSystem{}
}

func (wall *DeathSystem) Update(tick float64, world *esc.World) {
	for _, player := range world.PlayerCells() {
		if player.HP.HP <= 0 {
			player.Active.IsActive = false
		}
	}
}
