package systems

import "Velora/server/Internal/server/match"

type MovementSystem struct {
	match *match.Match
}

func NewMovementSystem(match *match.Match) *MovementSystem {
	return &MovementSystem{
		match: match,
	}
}

func (m *MovementSystem) Update(tick float64, world *match.World) {
	for entityId := range world.PlayerCells {
		if !world.Active[entityId].IsActive {
			continue
		}

		var direction = world.Directions[entityId]

		if direction.IsZero() {
			continue
		}

		var position = world.Positions[entityId]

		position.X += direction.X * match.BaseSpeed * match.TimeDeltaSeconds
		position.Y += direction.Y * match.BaseSpeed * match.TimeDeltaSeconds
		world.Positions[entityId] = position
	}
}
