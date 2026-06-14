package systems

import "Velora/esc"

const (
	DefaultPlayerSpeed = 5
)

type MovementSystem struct {
}

func (*MovementSystem) Name() string {
	return "MovementSystem"
}

func (*MovementSystem) Stage() Stage {
	return StageMovement
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{}
}

func (m *MovementSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	for _, player := range world.QueryActivePlayerCells() {
		var direction = player.Direction

		if direction.IsZero() {
			continue
		}

		var position = player.Position

		position.X += direction.X * DefaultPlayerSpeed * ctx.DeltaSeconds
		position.Y += direction.Y * DefaultPlayerSpeed * ctx.DeltaSeconds

		world.SetPosition(player.EntityID(), position)
	}
}
