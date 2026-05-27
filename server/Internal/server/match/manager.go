package match

import (
	"Velora/server/Internal"
	"Velora/server/Internal/server/contracts"
	"Velora/server/Internal/server/lobby"
	"Velora/server/pkg/packets"
	"errors"
	"sync"
	"time"
)

var (
	ErrCreateMatch      = errors.New("Create Match Error")
	ErrMatchNotFound    = errors.New("Match Not Found")
	ErrPlayerNotInMatch = errors.New("Player Not In Match")
)

const (
	PrepareDuration = 3 * time.Second
	ActiveDuration  = 180 * time.Second
)

type Manager struct {
	mu      sync.Mutex
	matches map[uint64]*Match

	clientMatches map[uint64]uint64
	userMatches   map[uint64]uint64
}

func (m *Manager) CreateMatch(config MatchConfig) (*Match, error) {
	if len(config.Players) == 0 {
		return nil, ErrCreateMatch
	}

	var match = &Match{
		mu:          sync.Mutex{},
		ID:          config.MatchId,
		MapSeed:     config.MapSeed,
		ServerTick:  uint64(0),
		Phase:       packets.MatchPhase_MATCH_PHASE_PREPARE,
		PhaseEndsAt: time.Now().Add(PrepareDuration),
		entityIds:   &Internal.IdGenerator{},
		players:     make(map[uint64]*PlayerRef),
		inputs:      make(map[uint64]PlayerInput),

		stop: make(chan struct{}),
		sync: sync.Once{},
	}

	for _, player := range config.Players {
		if _, ok := match.players[player.UserId]; ok {
			continue
		}

		match.players[player.UserId] = &PlayerRef{
			UserId:   player.UserId,
			ClientId: player.ClientId,
			Slot:     player.Slot,
			Client:   player.Client,
		}

		m.clientMatches[player.ClientId] = match.ID
		m.userMatches[player.UserId] = match.ID
	}

	return match, nil
}

func (m *Manager) HandleInput(client contracts.ClientInterface, input *packets.PlayerInputMessage) error {
	if !client.IsAuthenticated() {
		return lobby.ErrUserIsNotAuthenticated
	}

	var userId = client.GetUser().ID

	m.mu.Lock()
	var match, ok = m.matches[input.MatchId]
	m.mu.Unlock()

	if !ok {
		return ErrMatchNotFound
	}

	return match.HandleInput(userId, input)
}

func (m *Manager) RemoveClient(client contracts.ClientInterface) {
	if !client.IsAuthenticated() {
		return
	}

	var userID = client.GetUser().ID
	var clientId = client.Id()

	m.mu.Lock()
	var matchId, ok = m.clientMatches[clientId]
	var match = m.matches[matchId]
	m.mu.Unlock()

	if !ok {
		return
	}

	match.RemoveClient(userID, clientId)

	m.mu.Lock()
	delete(m.clientMatches, clientId)
	delete(m.userMatches, userID)

	if !match.HasConnectedClients() {
		match.Stop()
		delete(m.matches, matchId)
	}

	m.mu.Unlock()
}

func (m *Manager) StopMatch(matchId uint64) {

}
