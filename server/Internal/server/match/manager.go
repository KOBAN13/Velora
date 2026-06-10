package match

import (
	"Velora/esc"
	"Velora/server/Internal"
	"Velora/server/Internal/server/config"
	"Velora/server/pkg/packets"
	"Velora/systems"
	"errors"
	"sync"
	"time"
)

var (
	ErrCreateMatch            = errors.New("Create Match Error")
	ErrMatchNotFound          = errors.New("Match Not Found")
	ErrPlayerNotInMatch       = errors.New("Player Not In Match")
	ErrUserIsNotAuthenticated = errors.New("user is not authenticated")
)

type Manager struct {
	mu         sync.Mutex
	matches    map[uint64]*Match
	gameConfig config.GameConfig

	clientMatches map[uint64]uint64
	userMatches   map[uint64]uint64
}

func NewManager(gameConfig config.GameConfig) *Manager {
	return &Manager{
		mu:            sync.Mutex{},
		matches:       make(map[uint64]*Match),
		clientMatches: make(map[uint64]uint64),
		userMatches:   make(map[uint64]uint64),
		gameConfig:    gameConfig,
	}
}

func (m *Manager) CreateMatch(config MatchConfig) (*Match, error) {
	if len(config.Players) == 0 {
		return nil, ErrCreateMatch
	}

	var entityIds = &Internal.IdGenerator{}
	var players, err = sortedValidPlayers(config.Players)

	if err != nil {
		return nil, err
	}

	var capacity = esc.WorldCapacity{
		PlayerCells: len(players),
		Cores:       len(players),
		Nutrients:   m.gameConfig.Nutrient.MaxNutrients,
		Walls:       len(players),
	}

	var world = esc.NewWorld(capacity)

	for _, player := range players {
		var start = startPositions[player.Slot]

		world.CreatePlayerCell(esc.EntityId(entityIds.Next()), player.UserId, start.Cell, m.gameConfig.PlayerCell)
		world.CreateCore(esc.EntityId(entityIds.Next()), player.UserId, start.Core, m.gameConfig.Core)
	}

	world.CreateWall(esc.EntityId(entityIds.Next()), m.gameConfig.Wall)

	nutrientSpawner, err := NewNutrientSpawn(config.Players, m.gameConfig, config.MapSeed)

	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.matches[config.MatchId]; ok {
		return nil, ErrCreateMatch
	}

	var match = &Match{
		Mu:          sync.Mutex{},
		ID:          config.MatchId,
		RoomId:      config.RoomId,
		MapSeed:     config.MapSeed,
		ServerTick:  uint64(0),
		Phase:       packets.MatchPhase_MATCH_PHASE_PREPARE,
		PhaseEndsAt: time.Now().Add(PrepareDuration),
		players:     make(map[uint64]*PlayerRef),
		Inputs:      make(map[uint64]PlayerInput),
		Entities:    world,

		EntityIds:       entityIds,
		NutrientSpawner: nutrientSpawner,

		stop: make(chan struct{}),
		sync: sync.Once{},
	}

	var systemRunner = systems.NewSystemRunner(match)

	match.SystemRunner = systemRunner

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

		match.Inputs[player.UserId] = NewPlayerInput(0, 0)

		m.clientMatches[player.ClientId] = match.ID
		m.userMatches[player.UserId] = match.ID
	}

	m.matches[match.ID] = match

	return match, nil
}

func (m *Manager) HandleInput(client Client, input *packets.PlayerInputMessage) error {
	if !client.IsAuthenticated() {
		return ErrUserIsNotAuthenticated
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

func (m *Manager) RemoveClient(client Client) {
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

	if !match.RemoveClient(userID, clientId) {
		return
	}

	m.mu.Lock()
	delete(m.clientMatches, clientId)
	delete(m.userMatches, userID)
	m.mu.Unlock()

	if !match.HasConnectedClients() {
		m.StopMatch(matchId)
	}
}

func (m *Manager) StopMatch(matchId uint64) {
	m.mu.Lock()

	var match, ok = m.matches[matchId]

	if !ok {
		m.mu.Unlock()
		return
	}

	delete(m.matches, matchId)

	for _, player := range match.players {
		if player == nil {
			continue
		}

		delete(m.clientMatches, player.ClientId)
		delete(m.userMatches, player.UserId)
	}

	m.mu.Unlock()

	match.Stop()
}
