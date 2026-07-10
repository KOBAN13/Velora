package systems

import (
	"Velora/esc"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type InputSystem struct {
	input *esc_core.Query
}

func NewInputSystem(world *esc_core.World) (*InputSystem, error) {
	var input, err = esc_core.
		NewQuery(world, "PlayersInput").
		With(esc_core.Component[esc.PlayerTag]()).
		Read(esc_core.Component[esc.Active]()).
		Write(esc_core.Component[esc.MoveDirection]()).
		Build()

	if err != nil {
		return nil, err
	}

	return &InputSystem{input}, nil
}

func (*InputSystem) Name() string {
	return "InputSystem"
}

func (*InputSystem) Stage() esc_core.StageID {
	return StageInput
}

func (i *InputSystem) Access() esc_core.AccessSet {
	return i.input.Access()
}

func (i *InputSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		i.input.DebugInfo(),
	}
}

func (i *InputSystem) Update(ctx *esc_core.Context) error {
	var iter = i.input.Iter()

	for iter.Next() {
		var inputSlice, err = esc_core.GetResources[esc.PlayerInputSlice](ctx.Resources)

		if err != nil {
			return err
		}

		for _, playerInput := range *inputSlice {
			active, err := esc_core.Read[esc.Active](iter)

			if err != nil {
				return err
			}

			direction, err := esc_core.Write[esc.MoveDirection](iter)

			if err != nil {
				return err
			}

			if active.IsActive == false {
				direction.X = 0
				direction.Y = 0
			} else {
				direction.X = playerInput.MoveX
				direction.Y = playerInput.MoveY
			}
		}
	}

	return nil
}
