package systems

import (
	"Velora/esc"
	"Velora/server/Internal/server/match"
	"Velora/server/pkg/packets"
)

type WallGateSystem struct {
	match *match.Match
}

func NewWallGateSystem(match *match.Match) *WallGateSystem {
	return &WallGateSystem{
		match: match,
	}
}

func (wall *WallGateSystem) Update(tick float64, world *esc.World) {
	var openState bool

	if wall.match.Phase == packets.MatchPhase_MATCH_PHASE_PREPARE {
		openState = false
	} else if wall.match.Phase == packets.MatchPhase_MATCH_PHASE_ACTIVE {
		openState = true
	}

	for _, wallEntity := range world.Walls() {
		wallEntity.Open.Open = openState
	}
}
