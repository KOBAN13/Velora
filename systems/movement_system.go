package systems

import (
	"Velora/esc"
	"Velora/server/Internal/server/match"
)

type MovementSystem struct {
	match *match.Match
}

func NewMovementSystem(match *match.Match) *MovementSystem {
	return &MovementSystem{
		match: match,
	}
}

func (m *MovementSystem) Update(tick float64, world *esc.World) {
	for _, player := range world.PlayerCells() {
		if !player.Active.IsActive {
			continue
		}

		var direction = player.Direction

		if direction.IsZero() {
			continue
		}

		player.Position.X += direction.X * match.BaseSpeed * match.TimeDeltaSeconds
		player.Position.Y += direction.Y * match.BaseSpeed * match.TimeDeltaSeconds
	}
}
