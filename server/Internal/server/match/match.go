package match

import (
	"Velora/server/Internal"
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
	"math"
	"sync"
	"time"
)

type Match struct {
	mu sync.Mutex

	ID         uint64
	RoomId     uint64
	MapSeed    uint64
	ServerTick uint64

	Phase       packets.MatchPhase
	PhaseEndsAt time.Time

	entityIds *Internal.IdGenerator
	entities  *World

	players map[uint64]*PlayerRef
	inputs  map[uint64]PlayerInput

	stop chan struct{}
	sync sync.Once
}

type MatchConfig struct {
	RoomId  uint64
	MatchId uint64
	MapSeed uint64
	Players []PlayerRef
}

type PlayerRef struct {
	UserId   uint64
	ClientId uint64
	Slot     uint32
	Client   contracts.ClientInterface
}

type PlayerInput struct {
	MoveX      float32
	MoveY      float32
	ReceivedAt time.Time
}

func (m *Match) HandleInput(userId uint64, input *packets.PlayerInputMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.inputs[userId]; ok {
		return ErrPlayerNotInMatch
	}

	m.inputs[userId] = NewPlayerInput(input.MovePosition.X, input.MovePosition.Y)

	return nil
}

func (m *Match) Stop() {
	m.sync.Do(func() {
		close(m.stop)
	})
}

func (m *Match) updatePhase(now time.Time) {
	switch m.Phase {
	case packets.MatchPhase_MATCH_PHASE_PREPARE:
		if now.After(m.PhaseEndsAt) || now.Equal(m.PhaseEndsAt) {
			m.Phase = packets.MatchPhase_MATCH_PHASE_ACTIVE
			m.PhaseEndsAt = now.Add(ActiveDuration)
		}

	case packets.MatchPhase_MATCH_PHASE_ACTIVE:
		if now.After(m.PhaseEndsAt) || now.Equal(m.PhaseEndsAt) {
			m.Phase = packets.MatchPhase_MATCH_PHASE_ENDED
			m.PhaseEndsAt = time.Time{}
		}
	case packets.MatchPhase_MATCH_PHASE_ENDED:
	}
}

func (m *Match) phaseTimeLeftMs() int64 {
	if m.Phase == packets.MatchPhase_MATCH_PHASE_ENDED || m.PhaseEndsAt.IsZero() {
		return 0
	}

	left := time.Until(m.PhaseEndsAt)
	if left < 0 {
		return 0
	}

	return left.Milliseconds()
}

func NewPlayerInput(x float32, y float32) PlayerInput {
	if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
		x = 0
	}

	if math.IsNaN(float64(y)) || math.IsInf(float64(y), 0) {
		y = 0
	}

	var length = float32(math.Sqrt(float64(x*x + y*y)))

	if length > 1 {
		x /= length
		y /= length
	}

	return PlayerInput{
		MoveX:      x,
		MoveY:      y,
		ReceivedAt: time.Now(),
	}
}
