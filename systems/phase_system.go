package systems

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"time"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

const (
	ActiveDuration = 180 * time.Second
)

type PhaseSystem struct {
	phase *esc_core.Query
}

func NewPhaseSystem(world *esc_core.World) (*PhaseSystem, error) {
	var phase esc_core.Query
}

func (*PhaseSystem) Name() string {
	return "PhaseSystem"
}

func (*PhaseSystem) Stage() esc_core.StageID {
	return StagePhase
}

func (phase *PhaseSystem) Access() esc_core.AccessSet {
	return phase.phase.Access()
}

func (phase *PhaseSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		phase.phase.DebugInfo(),
	}
}

func (phase *PhaseSystem) Update(ctx *esc_core.Context) error {
	phase.updatePhase(ctx)
}

func (phase *PhaseSystem) updatePhase(ctx *esc_core.Context) {
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
