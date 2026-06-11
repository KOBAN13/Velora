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
	for _, player := range world.PlayerCells() {
		var input = i.match.Inputs[player.OwnerId.UserId]

		if player.Active.IsActive == false {
			player.Direction = esc.Zero()
		} else {
			player.Direction = esc.MoveDirection{
				X: input.MoveX,
				Y: input.MoveY,
			}
		}
	}
}
