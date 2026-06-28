package systems

import (
	"Velora/esc"
	"Velora/server/pkg/packets"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type WallGateSystem struct {
	wall *esc_core.Query
}

func NewWallGateSystem() *WallGateSystem {
	return &WallGateSystem{}
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
	var openState bool

	if ctx.Phase == packets.MatchPhase_MATCH_PHASE_PREPARE {
		openState = false
	} else if ctx.Phase == packets.MatchPhase_MATCH_PHASE_ACTIVE {
		openState = true
	}

	for _, wallEntity := range world.QueryWalls() {
		ctx.Commands.Add(&esc.SetWallStateCommand{EntityId: wallEntity.EntityID(), Open: openState})
	}
}
