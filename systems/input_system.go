package systems

import (
	"Velora/esc"
	"Velora/server/Internal/server/match"
)

type InputSystem struct {
	match *match.Match
}

func NewInputSystem(match *match.Match) *InputSystem {
	return &InputSystem{
		match: match,
	}
}

func (i *InputSystem) Update(tick float64, world *esc.World) {
	for entityId := range world.PlayerCells {
		var owner = world.Owners[entityId]
		var input = i.match.Inputs[owner.UserId]

		if world.Active[entityId].IsActive == false {
			world.Directions[entityId] = esc.Zero()
		} else {
			world.Directions[entityId] = esc.MoveDirection{
				X: input.MoveX,
				Y: input.MoveY,
			}
		}
	}
}
