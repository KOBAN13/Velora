package match

import (
	"Velora/server/Internal"
	"Velora/server/Internal/server/contracts"
	"Velora/server/pkg/packets"
	"sync"
	"time"
)

type Manager struct {
	mu      sync.Mutex
	matches map[uint64]*Match
}

type MatchConfig struct {
	RoomId  uint64
	MatchId uint64
	MapSeed uint64
	Players []PlayerRef
}

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
}

type PlayerRef struct {
	UserId   uint64
	ClientId uint64
	Slot     uint32
	Client   contracts.ClientInterface
}

func (m *Manager)  *MatchConfig {

}
