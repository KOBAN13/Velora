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

type PhaseSystem struct{}

func NewPhaseSystem() (*PhaseSystem, error) {
	return &PhaseSystem{}, nil
}

func (*PhaseSystem) Name() string {
	return "PhaseSystem"
}

func (*PhaseSystem) Stage() esc_core.StageID {
	return StagePhase
}

func (phase *PhaseSystem) Update(ctx *esc_core.Context) error {
	return phase.updatePhase(ctx)
}

func (phase *PhaseSystem) updatePhase(ctx *esc_core.Context) error {
	var phaseResource, err = esc_core.GetResources[esc.MatchPhaseResource](ctx.Resources)

	if err != nil {
		return err
	}

	switch phaseResource.Phase {
	case packets.MatchPhase_MATCH_PHASE_PREPARE:
		if phaseExpired(phaseResource.Now, phaseResource.PhaseEndsAt) {
			phaseResource.Phase = packets.MatchPhase_MATCH_PHASE_ACTIVE
			phaseResource.PhaseEndsAt = phaseResource.Now.Add(ActiveDuration)
		}

	case packets.MatchPhase_MATCH_PHASE_ACTIVE:
		if phaseExpired(phaseResource.Now, phaseResource.PhaseEndsAt) {
			phaseResource.Phase = packets.MatchPhase_MATCH_PHASE_ENDED
			phaseResource.PhaseEndsAt = time.Time{}
		}
	case packets.MatchPhase_MATCH_PHASE_ENDED:
	}

	return nil
}

func phaseExpired(now time.Time, endsAt time.Time) bool {
	return !endsAt.IsZero() && !now.Before(endsAt)
}
