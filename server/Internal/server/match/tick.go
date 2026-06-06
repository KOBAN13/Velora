package match

import (
	"Velora/server/pkg/packets"
	"time"
)

const (
	TickRate         = 20
	TickDuration     = 50 * time.Millisecond
	TimeDeltaSeconds = float32(0.05)
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
	currentTick := float64(m.ServerTick)

	m.SystemRunner.UpdateSystems(currentTick, m.Entities)

	snapshot = BuildMatchSnapshot(m, m.Entities, now)
	clients = m.connectedClientsLocked()

	m.Mu.Unlock()

	for _, client := range clients {
		client.SocketSend(snapshot)
	}
}
