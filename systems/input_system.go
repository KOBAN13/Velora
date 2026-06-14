package systems

import (
	"Velora/esc"
)

type InputSystem struct{}

func (*InputSystem) Name() string {
	return "InputSystem"
}

func (*InputSystem) Stage() Stage {
	return StageInput
}

func NewInputSystem() *InputSystem {
	return &InputSystem{}
}

func (i *InputSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	for _, player := range world.QueryPlayerCells() {
		var input = ctx.Resources.Inputs.Inputs[player.OwnerId.UserId]

		if player.Active.IsActive == false {
			world.SetDirection(player.EntityID(), esc.Zero())
		} else {
			world.SetDirection(player.EntityID(), esc.MoveDirection{
				X: input.MoveX,
				Y: input.MoveY,
			})
		}
	}
}
