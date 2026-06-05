package match

import "time"

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

}
