package match

import (
	"Velora/esc"
	"Velora/server/Internal/server/config"
	"cmp"
	"errors"
	"math/rand/v2"
	"slices"
)

var (
	ErrPlayerNotFound    = errors.New("player not found")
	ErrPlayerSlotInvalid = errors.New("player slot invalid")
	ErrPlayerSlotUsed    = errors.New("player slot already used")
	ErrPlayerUsed        = errors.New("player already used")
)

type startPosition struct {
	Cell esc.Position
	Core esc.Position
}

var startPositions = []startPosition{
	{Cell: esc.Position{X: -10, Y: 0}, Core: esc.Position{X: -14, Y: 0}},
	{Cell: esc.Position{X: 10, Y: 0}, Core: esc.Position{X: 14, Y: 0}},
	{Cell: esc.Position{X: 0, Y: 10}, Core: esc.Position{X: 0, Y: 14}},
	{Cell: esc.Position{X: 0, Y: -10}, Core: esc.Position{X: 0, Y: -14}},
	{Cell: esc.Position{X: 10, Y: 10}, Core: esc.Position{X: 14, Y: 14}},
}

func NewNutrientSpawn(players []PlayerRef, gameConfig config.GameConfig, mapSeed uint64) (NutrientSpawner, error) {
	if _, err := sortedValidPlayers(players); err != nil {
		return NutrientSpawner{}, err
	}

	var spawner = NewNutrientSpawner(mapSeed, gameConfig.Nutrient)

	return spawner, nil
}

func sortedValidPlayers(players []PlayerRef) ([]PlayerRef, error) {
	if len(players) == 0 {
		return nil, ErrPlayerNotFound
	}

	var sortedPlayers = slices.Clone(players)

	slices.SortFunc(sortedPlayers, func(a, b PlayerRef) int {
		return cmp.Compare(a.Slot, b.Slot)
	})

	var usedPlayers = make(map[uint64]struct{}, len(sortedPlayers))
	var usedSlots = make(map[uint32]struct{}, len(sortedPlayers))

	for _, player := range sortedPlayers {
		if int(player.Slot) >= len(startPositions) {
			return nil, ErrPlayerSlotInvalid
		}

		if _, ok := usedPlayers[player.UserId]; ok {
			return nil, ErrPlayerUsed
		}

		if _, ok := usedSlots[player.Slot]; ok {
			return nil, ErrPlayerSlotUsed
		}

		usedPlayers[player.UserId] = struct{}{}
		usedSlots[player.Slot] = struct{}{}
	}

	return sortedPlayers, nil
}

type NutrientSpawner struct {
	Rng *rand.Rand

	LastSpawnTick float64
	SpawnInterval float64
	MaxNutrients  int
	SpawnBatch    int
	MaxAttempts   int

	ArenaHalfSize       float32
	MinPlayerDistance   float32
	MinCoreDistance     float32
	MinNutrientDistance float32

	NutrientActive bool
}

func NewNutrientSpawner(mapSeed uint64, nutrientConfig config.NutrientConfig) NutrientSpawner {
	return NutrientSpawner{
		Rng: rand.New(rand.NewPCG(mapSeed, mapSeed^0x9e3779b97f4a7c15)),

		LastSpawnTick: 0,
		SpawnInterval: nutrientConfig.SpawnInterval,
		MaxNutrients:  nutrientConfig.MaxNutrients,
		SpawnBatch:    nutrientConfig.SpawnBatch,
		MaxAttempts:   defaultNutrientMaxAttempts,

		ArenaHalfSize:       defaultNutrientArenaHalfSize,
		MinPlayerDistance:   defaultNutrientMinPlayerDistance,
		MinCoreDistance:     defaultNutrientMinCoreDistance,
		MinNutrientDistance: defaultNutrientMinNutrientDistance,

		NutrientActive: nutrientConfig.Alive,
	}
}
