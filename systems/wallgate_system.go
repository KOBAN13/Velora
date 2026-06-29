package systems

import (
	"Velora/esc"
	"Velora/server/pkg/packets"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type WallGateSystem struct {
	wall *esc_core.Query
}

func NewWallGateSystem(world *esc_core.World) (*WallGateSystem, error) {
	var wall, err = esc_core.
		NewQuery(world, "WallGateSystem").
		With(esc_core.Component[esc.WallTag]()).
		Write(esc_core.Component[esc.WallState]()).
		Build()

	if err != nil {
		return nil, err
	}

	return &WallGateSystem{wall: wall}, nil
}

func (*WallGateSystem) Name() string {
	return "WallGateSystem"
}

func (*WallGateSystem) Stage() esc_core.StageID {
	return StageRules
}

func (wall *WallGateSystem) Access() esc_core.AccessSet {
	return wall.wall.Access()
}

func (wall *WallGateSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		wall.wall.DebugInfo(),
	}
}

func (wall *WallGateSystem) Update(ctx *esc_core.Context) error {
	var phaseResource, err = esc_core.GetResources[esc.MatchPhaseResource](ctx.Resources)

	if err != nil {
		return err
	}

	var openState bool

	if phaseResource.Phase == packets.MatchPhase_MATCH_PHASE_PREPARE {
		openState = false
	} else if phaseResource.Phase == packets.MatchPhase_MATCH_PHASE_ACTIVE {
		openState = true
	}

	var it = wall.wall.Iter()

	for it.Next() {
		var wallState, err = esc_core.Write[esc.WallState](it)

		if err != nil {
			return err
		}

		wallState.Open = openState
	}

	return nil
}
