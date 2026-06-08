package match

import (
	"Velora/server/Internal"
	"Velora/server/Internal/server/db"
	"Velora/server/Internal/server/match/systems"
	"Velora/server/pkg/packets"
	"math"
	"sync"
	"time"
)

type Match struct {
	Mu sync.Mutex

	ID         uint64
	RoomId     uint64
	MapSeed    uint64
	ServerTick uint64

	SystemRunner *systems.SystemRunner
	Phase        packets.MatchPhase
	PhaseEndsAt  time.Time

	Entities *World

	players map[uint64]*PlayerRef
	Inputs  map[uint64]PlayerInput

	EntityIds       *Internal.IdGenerator
	NutrientSpawner NutrientSpawner

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
	Client   Client
}

type PlayerInput struct {
	MoveX      float32
	MoveY      float32
	ReceivedAt time.Time
}

type Client interface {
	Id() uint64
	GetUser() *db.User
	IsAuthenticated() bool
	SocketSend(message packets.Msg)
}

func (m *Match) HandleInput(userId uint64, input *packets.PlayerInputMessage) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, ok := m.Inputs[userId]; !ok {
		return ErrPlayerNotInMatch
	}

	if input.MovePosition == nil {
		m.Inputs[userId] = NewPlayerInput(0, 0)
		return nil
	}

	m.Inputs[userId] = NewPlayerInput(input.MovePosition.X, input.MovePosition.Y)

	return nil
}

func (m *Match) Stop() {
	m.sync.Do(func() {
		close(m.stop)
	})
}

func (m *Match) HasConnectedClients() bool {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	for _, player := range m.players {
		if player != nil {
			return true
		}
	}

	return false
}

func (m *Match) RemoveClient(userId uint64, clientId uint64) bool {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	var player, ok = m.players[userId]

	if !ok {
		return false
	}

	if player.ClientId != clientId {
		return false
	}

	player.Client = nil

	m.Inputs[player.UserId] = PlayerInput{
		ReceivedAt: time.Now(),
		MoveY:      0,
		MoveX:      0,
	}

	return true
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

func (m *Match) connectedClientsLocked() []Client {
	var clients = make([]Client, 0, len(m.players))

	for _, player := range m.players {
		if player == nil || player.Client == nil {
			continue
		}

		clients = append(clients, player.Client)
	}

	return clients
}
