package systems

import (
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

func (wall *WallGateSystem) Update(tick float64, world *match.World) {
	var openState bool

	if wall.match.Phase == packets.MatchPhase_MATCH_PHASE_PREPARE {
		openState = false
	} else if wall.match.Phase == packets.MatchPhase_MATCH_PHASE_ACTIVE {
		openState = true
	}

	for id := range world.Walls {
		var state = world.WallStates[id]
		state.Open = openState
		world.WallStates[id] = state
	}
}
