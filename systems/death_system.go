package systems

import (
	"Velora/esc"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type DeathSystem struct {
	death *esc_core.Query
}

func NewDeathSystem(world *esc_core.World) (*DeathSystem, error) {
	var death, err = esc_core.
		NewQuery(world, "PlayerDeath").
		With(esc_core.Component[esc.PlayerTag]()).
		Read(esc_core.Component[esc.Health]()).
		Write(esc_core.Component[esc.Active]()).
		Build()

	if err != nil {
		return nil, err
	}

	return &DeathSystem{death: death}, nil
}

func (*DeathSystem) Name() string {
	return "DeathSystem"
}

func (*DeathSystem) Stage() esc_core.StageID {
	return StageCleanup
}

func (d *DeathSystem) Access() esc_core.AccessSet {
	return d.death.Access()
}

func (d *DeathSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		d.death.DebugInfo(),
	}
}

func (d *DeathSystem) Update(ctx *esc_core.Context) error {
	var it = d.death.Iter()

	for it.Next() {
		health, err := esc_core.Read[esc.Health](it)

		if err != nil {
			return err
		}

		active, err := esc_core.Write[esc.Active](it)

		if err != nil {
			return err
		}

		if health.Value <= 0 && active.IsActive {
			active.IsActive = false
		}
	}

	return nil
}
