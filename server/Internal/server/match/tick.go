package match

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"time"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

func (m *Match) Run() {
	var ticker = time.NewTicker(TickDuration)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.Tick(time.Now())
		case <-m.stop:
			return
		}
	}
}

func (m *Match) Tick(now time.Time) {
	var snapshot packets.Msg
	var clients []Client

	m.Mu.Lock()

	m.ServerTick++

	var ctx = m.newSystemContextLocked(now, TimeDeltaSeconds)

	var err = m.SystemRunner.UpdateSystems(ctx)
	if err != nil {
		return
	}

	m.applySystemContextLocked()

	snapshot = BuildMatchSnapshot(m, m.World, now)
	clients = m.connectedClientsLocked()

	m.Mu.Unlock()

	for _, client := range clients {
		client.SocketSend(snapshot)
	}
}

func (m *Match) InitializeSystems(now time.Time) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	err := m.SystemRunner.BuildSystems(m.World)
	if err != nil {
		return err
	}

	var ctx = m.newSystemContextLocked(now, 0)

	if err := m.SystemRunner.InitializeSystems(ctx); err != nil {
		return err
	}

	m.applySystemContextLocked()

	return nil
}

func (m *Match) newSystemContextLocked(now time.Time, deltaSeconds float32) *esc_core.Context {
	var err = esc_core.RemoveResources[esc.PlayerInputSlice](m.Resources)
	if err != nil {
		return nil
	}

	var inputs = make(esc.PlayerInputSlice, 0, len(m.Inputs))

	for _, input := range m.Inputs {
		inputs = append(inputs, input)
	}

	err = esc_core.PutResource[esc.PlayerInputSlice](m.Resources, inputs)
	if err != nil {
		return nil
	}

	var matchPhaseResource = esc.MatchPhaseResource{
		Now:         now,
		Phase:       m.Phase,
		PhaseEndsAt: m.PhaseEndsAt,
	}

	err = esc_core.PutResource[esc.MatchPhaseResource](m.Resources, matchPhaseResource)

	return &esc_core.Context{
		Tick:         m.ServerTick,
		DeltaSeconds: deltaSeconds,
		Commands:     esc_core.NewCommandBuffer(),
		Resources:    m.Resources,
	}
}

func (m *Match) applySystemContextLocked() {
	matchPhase, err := esc_core.GetResources[esc.MatchPhaseResource](m.Resources)

	if err != nil {
		return
	}

	m.Phase = matchPhase.Phase
	m.PhaseEndsAt = matchPhase.PhaseEndsAt
}
