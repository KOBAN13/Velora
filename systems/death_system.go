package systems

import (
	"Velora/esc"
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

func (wall *DeathSystem) Update(tick float64, world *esc.World) {
	for _, player := range world.PlayerCells() {
		if player.HP.HP <= 0 {
			player.Active.IsActive = false
		}
	}
}
