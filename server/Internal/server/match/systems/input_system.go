package systems

import "Velora/server/Internal/server/match"

type InputSystem struct {
	match *match.Match
}

func NewInputSystem(match *match.Match) *InputSystem {
	return &InputSystem{
		match: match,
	}
}

func (i *InputSystem) Update(tick float64, world *match.World) {
	for entityId := range world.PlayerCells {
		var owner = world.Owners[entityId]
		var input = i.match.Inputs[owner.UserId]

		if world.Active[entityId].IsActive == false {
			world.Directions[entityId] = match.Zero()
		} else {
			world.Directions[entityId] = match.MoveDirection{
				X: input.MoveX,
				Y: input.MoveY,
			}
		}
	}
}
