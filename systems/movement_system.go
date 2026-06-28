package systems

import (
	"Velora/esc"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

const (
	DefaultPlayerSpeed = 5
)

type MovementSystem struct {
	players *esc_core.Query
}

func (*MovementSystem) Name() string {
	return "MovementSystem"
}

func (*MovementSystem) Stage() esc_core.StageID {
	return StageMovement
}

func (m *MovementSystem) Access() esc_core.AccessSet {
	return m.players.Access()
}

func (m *MovementSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		m.players.DebugInfo(),
	}
}

func NewMovementSystem(world *esc_core.World) (*MovementSystem, error) {
	var players, err = esc_core.
		NewQuery(world, "PlayersMovement").
		With(esc_core.Component[esc.PlayerTag]()).
		Write(esc_core.Component[esc.Position]()).
		Read(esc_core.Component[esc.MoveDirection]()).
		Read(esc_core.Component[esc.Active]()).
		Build()

	if err != nil {
		return nil, err
	}

	return &MovementSystem{players: players}, nil
}

func (m *MovementSystem) Update(ctx *esc_core.Context) error {
	var it = m.players.Iter()

	for it.Next() {
		active, err := esc_core.Read[esc.Active](it)

		if err != nil {
			return err
		}

		if !active.IsActive {
			continue
		}

		direction, err := esc_core.Read[esc.MoveDirection](it)

		if err != nil {
			return err
		}

		if direction.IsZero() {
			continue
		}

		position, err := esc_core.Read[esc.Position](it)

		if err != nil {
			return err
		}

		position.X += direction.X * DefaultPlayerSpeed * ctx.DeltaSeconds
		position.Y += direction.Y * DefaultPlayerSpeed * ctx.DeltaSeconds
	}

	return nil
}
