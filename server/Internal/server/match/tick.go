package match

import (
	"time"

	"Velora/esc"
	"Velora/server/pkg/packets"

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
	var clients []Client
	var snapshot packets.Msg

	var ok = func() bool {
		m.Mu.Lock()
		defer m.Mu.Unlock()

		m.ServerTick++

		ctx, err := m.newSystemContextLocked(now, TimeDeltaSeconds)
		if err != nil {
			return false
		}

		err = m.SystemRunner.UpdateSystems(ctx)

		if err != nil {
			return false
		}

		m.applySystemContextLocked()

		players, err := esc_core.NewQuery(m.World, "SnapshotPlayers").With(esc_core.Component[esc.PlayerTag]()).Build()
		if err != nil {
			return false
		}

		nutrients, err := esc_core.NewQuery(m.World, "SnapshotNutrients").With(esc_core.Component[esc.NutrientTag]()).Build()
		if err != nil {
			return false
		}

		cores, err := esc_core.NewQuery(m.World, "SnapshotCores").With(esc_core.Component[esc.CoreTag]()).Build()
		if err != nil {
			return false
		}

		walls, err := esc_core.NewQuery(m.World, "SnapshotWalls").With(esc_core.Component[esc.WallTag]()).Build()
		if err != nil {
			return false
		}

		var snapshotQueries = &SnapshotQueries{
			players:   players,
			cores:     cores,
			nutrients: nutrients,
			walls:     walls,
		}

		snapshot = BuildMatchSnapshot(m, snapshotQueries, now)
		clients = m.connectedClientsLocked()

		return snapshot != nil
	}()

	if !ok {
		return
	}

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

	ctx, err := m.newSystemContextLocked(now, 0)
	if err != nil {
		return err
	}

	if err := m.SystemRunner.InitializeSystems(ctx); err != nil {
		return err
	}

	m.applySystemContextLocked()

	return nil
}

func (m *Match) newSystemContextLocked(now time.Time, deltaSeconds float32) (*esc_core.Context, error) {
	_ = esc_core.RemoveResources[esc.PlayerInputSlice](m.Resources)
	var inputs = make(esc.PlayerInputSlice, 0, len(m.Inputs))

	for _, input := range m.Inputs {
		inputs = append(inputs, input)
	}

	err := esc_core.PutResource[esc.PlayerInputSlice](m.Resources, inputs)
	if err != nil {
		return nil, err
	}

	var matchPhaseResource = esc.MatchPhaseResource{
		Now:         now,
		Phase:       m.Phase,
		PhaseEndsAt: m.PhaseEndsAt,
	}

	err = esc_core.PutResource[esc.MatchPhaseResource](m.Resources, matchPhaseResource)
	if err != nil {
		return nil, err
	}

	return &esc_core.Context{
		Tick:         m.ServerTick,
		DeltaSeconds: deltaSeconds,
		World:        m.World,
		Commands:     esc_core.NewCommandBuffer(),
		Resources:    m.Resources,
	}, nil
}

func (m *Match) applySystemContextLocked() {
	matchPhase, err := esc_core.GetResources[esc.MatchPhaseResource](m.Resources)

	if err != nil {
		return
	}

	m.Phase = matchPhase.Phase
	m.PhaseEndsAt = matchPhase.PhaseEndsAt
}
