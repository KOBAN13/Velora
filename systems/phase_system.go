package systems

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"time"
)

const (
	ActiveDuration = 180 * time.Second
)

type PhaseSystem struct {
}

func (*PhaseSystem) Name() string {
	return "PhaseSystem"
}

func (*PhaseSystem) Stage() Stage {
	return StagePhase
}

func NewPhaseSystem() *PhaseSystem {
	return &PhaseSystem{}
}

func (phase *PhaseSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	phase.updatePhase(ctx)
}

func (phase *PhaseSystem) updatePhase(ctx *esc.SystemContext) {
	switch ctx.Phase {
	case packets.MatchPhase_MATCH_PHASE_PREPARE:
		if ctx.Now.After(ctx.PhaseEndsAt) || ctx.Now.Equal(ctx.PhaseEndsAt) {
			ctx.Phase = packets.MatchPhase_MATCH_PHASE_ACTIVE
			ctx.PhaseEndsAt = ctx.Now.Add(ActiveDuration)
		}

	case packets.MatchPhase_MATCH_PHASE_ACTIVE:
		if ctx.Now.After(ctx.PhaseEndsAt) || ctx.Now.Equal(ctx.PhaseEndsAt) {
			ctx.Phase = packets.MatchPhase_MATCH_PHASE_ENDED
			ctx.PhaseEndsAt = time.Time{}
		}
	case packets.MatchPhase_MATCH_PHASE_ENDED:
	}
}
