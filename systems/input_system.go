package systems

import (
	"Velora/esc"
)

type InputSystem struct{}

func (*InputSystem) Name() string {
	return "DeathSystem"
}

func (*InputSystem) Stage() Stage {
	return StageInput
}

func NewInputSystem() *InputSystem {
	return &InputSystem{}
}

func (i *InputSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	for _, player := range world.QueryActivePlayerCells() {
		var input = ctx.Resources.Inputs.Inputs[player.OwnerId.UserId]

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
