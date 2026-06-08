package systems

import (
	"Velora/server/Internal/server/match"
)

type DeathSystem struct {
	match *match.Match
}

func NewDeathSystem(match *match.Match) *DeathSystem {
	return &DeathSystem{
		match: match,
	}
}

func (wall *DeathSystem) Update(tick float64, world *match.World) {
	for entityId := range world.PlayerCells {
		var hpComponent = world.Health[entityId]

		if hpComponent.HP <= 0 {
			var activeComponent = world.Active[entityId]
			activeComponent.IsActive = false
			world.Active[entityId] = activeComponent
		}
	}
}
