package match

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"time"
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

	m.SystemRunner.UpdateSystems(ctx, m.Entities)
	m.applySystemContextLocked(ctx)

	snapshot = BuildMatchSnapshot(m, m.Entities, now)
	clients = m.connectedClientsLocked()

	m.Mu.Unlock()

	for _, client := range clients {
		client.SocketSend(snapshot)
	}
}

func (m *Match) InitializeSystems(now time.Time) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	m.SystemRunner.BuildSystems()

	var ctx = m.newSystemContextLocked(now, 0)

	if err := m.SystemRunner.InitializeSystems(ctx, m.Entities); err != nil {
		return err
	}

	m.applySystemContextLocked(ctx)

	return nil
}

func (m *Match) newSystemContextLocked(now time.Time, deltaSeconds float32) *esc.SystemContext {
	clear(m.Resources.Inputs.Inputs)

	for userId, input := range m.Inputs {
		m.Resources.Inputs.Inputs[userId] = input
	}

	return &esc.SystemContext{
		Tick:         m.ServerTick,
		DeltaSeconds: deltaSeconds,
		Now:          now,
		Phase:        m.Phase,
		PhaseEndsAt:  m.PhaseEndsAt,
		Commands:     &esc.CommandBuffer{},
		Resources:    m.Resources,
	}
}

func (m *Match) applySystemContextLocked(ctx *esc.SystemContext) {
	m.Phase = ctx.Phase
	m.PhaseEndsAt = ctx.PhaseEndsAt
}
