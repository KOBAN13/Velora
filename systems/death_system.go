package systems

import (
	"Velora/esc"
)

type DeathSystem struct {
}

func NewDeathSystem() *DeathSystem {
	return &DeathSystem{}
}

func (*DeathSystem) Name() string {
	return "DeathSystem"
}

func (*DeathSystem) Stage() Stage {
	return StageCleanup
}

func (*DeathSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	for _, player := range world.QueryActivePlayerCells() {
		if player.HP.Value <= 0 {
			player.Active.IsActive = false
		}
	}
}
