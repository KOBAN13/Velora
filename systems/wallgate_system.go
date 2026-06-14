package systems

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
)

type WallGateSystem struct {
}

func NewWallGateSystem() *WallGateSystem {
	return &WallGateSystem{}
}

func (*WallGateSystem) Name() string {
	return "WallGateSystem"
}

func (*WallGateSystem) Stage() Stage {
	return StageRules
}

func (wall *WallGateSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	var openState bool

	if ctx.Phase == packets.MatchPhase_MATCH_PHASE_PREPARE {
		openState = false
	} else if ctx.Phase == packets.MatchPhase_MATCH_PHASE_ACTIVE {
		openState = true
	}

	for _, wallEntity := range world.QueryWalls() {
		ctx.Commands.Add(&esc.SetActiveCommand{EntityId: wallEntity.EntityID(), Active: openState})
	}
}
