package systems

import (
	"Velora/server/Internal/server/match"
	"Velora/server/pkg/packets"
	"time"
)

type PhaseSystem struct {
	match *match.Match
}

func NewPhaseSystem(match *match.Match) *PhaseSystem {
	return &PhaseSystem{
		match: match,
	}
}

func (phase *PhaseSystem) Update(tick float64, world *match.World) {
	phase.updatePhase(time.Now())
}

func (phase *PhaseSystem) updatePhase(now time.Time) {
	switch phase.match.Phase {
	case packets.MatchPhase_MATCH_PHASE_PREPARE:
		if now.After(phase.match.PhaseEndsAt) || now.Equal(phase.match.PhaseEndsAt) {
			phase.match.Phase = packets.MatchPhase_MATCH_PHASE_ACTIVE
			phase.match.PhaseEndsAt = now.Add(match.ActiveDuration)
		}

	case packets.MatchPhase_MATCH_PHASE_ACTIVE:
		if now.After(phase.match.PhaseEndsAt) || now.Equal(phase.match.PhaseEndsAt) {
			phase.match.Phase = packets.MatchPhase_MATCH_PHASE_ENDED
			phase.match.PhaseEndsAt = time.Time{}
		}
	case packets.MatchPhase_MATCH_PHASE_ENDED:
	}
}
