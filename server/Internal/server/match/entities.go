package match

import (
	"Velora/server/Internal"
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

type EntityId uint64

type World struct {
	PlayerCells map[EntityId]PlayerCell
	Cores       map[EntityId]Core
	Nutrients   map[EntityId]Nutrient
	Walls       map[EntityId]Wall

	Directions     map[EntityId]MoveDirection
	Positions      map[EntityId]Position
	Owners         map[EntityId]Owner
	Health         map[EntityId]Health
	MaxHealth      map[EntityId]Health
	Biomass        map[EntityId]Biomass
	Levels         map[EntityId]Level
	Active         map[EntityId]Active
	NutrientValues map[EntityId]NutrientValue
	WallStates     map[EntityId]WallState
}

type PlayerCell struct{}

type Core struct{}

type Nutrient struct{}

type Wall struct{}

type Position struct {
	X float32
	Y float32
}

type MoveDirection struct {
	X float32
	Y float32
}

type Owner struct {
	UserId uint64
}

type Health struct {
	HP int32
}

type Biomass struct {
	Value uint32
}

type Level struct {
	Value uint32
}

type Active struct {
	IsActive bool
}

type NutrientValue struct {
	Value uint32
}

type WallState struct {
	Open bool
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

type startPosition struct {
	Cell Position
	Core Position
}

var startPositions = []startPosition{
	{Cell: Position{X: -10, Y: 0}, Core: Position{X: -14, Y: 0}},
	{Cell: Position{X: 10, Y: 0}, Core: Position{X: 14, Y: 0}},
	{Cell: Position{X: 0, Y: 10}, Core: Position{X: 0, Y: 14}},
	{Cell: Position{X: 0, Y: -10}, Core: Position{X: 0, Y: -14}},
	{Cell: Position{X: 10, Y: 10}, Core: Position{X: 14, Y: 14}},
}

func NewWorld(
	players []PlayerRef,
	gameConfig config.GameConfig,
	mapSeed uint64,
	idGenerator *Internal.IdGenerator,
) (*World, NutrientSpawner, error) {
	if len(players) == 0 {
		return nil, NutrientSpawner{}, ErrPlayerNotFound
	}

	var world = NewEmptyWorld(len(players))
	var sortedPlayers = slices.Clone(players)

	slices.SortFunc(sortedPlayers, func(a, b PlayerRef) int {
		return cmp.Compare(a.Slot, b.Slot)
	})

	var usedPlayers = make(map[uint64]struct{}, len(sortedPlayers))
	var usedSlots = make(map[uint32]struct{}, len(sortedPlayers))

	for _, player := range sortedPlayers {
		if int(player.Slot) >= len(startPositions) {
			return nil, NutrientSpawner{}, ErrPlayerSlotInvalid
		}

		if _, ok := usedPlayers[player.UserId]; ok {
			return nil, NutrientSpawner{}, ErrPlayerUsed
		}

		if _, ok := usedSlots[player.Slot]; ok {
			return nil, NutrientSpawner{}, ErrPlayerSlotUsed
		}

		usedPlayers[player.UserId] = struct{}{}
		usedSlots[player.Slot] = struct{}{}

		var start = startPositions[player.Slot]

		world.CreatePlayerCell(EntityId(idGenerator.Next()), player.UserId, start.Cell, gameConfig.PlayerCell)
		world.CreateCore(EntityId(idGenerator.Next()), player.UserId, start.Core, gameConfig.Core)
	}

	world.CreateWall(EntityId(idGenerator.Next()), gameConfig.Wall)

	var spawner = NewNutrientSpawner(mapSeed, gameConfig.Nutrient)

	return world, spawner, nil
}

func NewEmptyWorld(playerCount int) *World {
	return &World{
		PlayerCells: make(map[EntityId]PlayerCell, playerCount),
		Cores:       make(map[EntityId]Core, playerCount),
		Nutrients:   make(map[EntityId]Nutrient),
		Walls:       make(map[EntityId]Wall, 1),

		Directions:     make(map[EntityId]MoveDirection),
		Positions:      make(map[EntityId]Position, playerCount*2+1),
		Owners:         make(map[EntityId]Owner, playerCount*2),
		Health:         make(map[EntityId]Health, playerCount*2),
		MaxHealth:      make(map[EntityId]Health, playerCount*2),
		Biomass:        make(map[EntityId]Biomass, playerCount),
		Levels:         make(map[EntityId]Level, playerCount),
		Active:         make(map[EntityId]Active, playerCount),
		NutrientValues: make(map[EntityId]NutrientValue),
		WallStates:     make(map[EntityId]WallState, 1),
	}
}

func (w *World) CreatePlayerCell(
	id EntityId,
	ownerId uint64,
	position Position,
	playerConfig config.PlayerCellConfig,
) {
	w.PlayerCells[id] = PlayerCell{}
	w.Positions[id] = position
	w.Owners[id] = Owner{UserId: ownerId}
	w.Health[id] = Health{HP: int32(playerConfig.HP)}
	w.MaxHealth[id] = Health{HP: int32(playerConfig.MaxHP)}
	w.Biomass[id] = Biomass{Value: uint32(playerConfig.Biomass)}
	w.Levels[id] = Level{Value: uint32(playerConfig.Level)}
	w.Active[id] = Active{IsActive: playerConfig.Alive}
}

func (w *World) CreateCore(
	id EntityId,
	ownerId uint64,
	position Position,
	coreConfig config.CoreConfig,
) {
	w.Cores[id] = Core{}
	w.Positions[id] = position
	w.Owners[id] = Owner{UserId: ownerId}
	w.Health[id] = Health{HP: int32(coreConfig.HP)}
	w.MaxHealth[id] = Health{HP: int32(coreConfig.MaxHP)}
}

func (w *World) CreateNutrient(
	id EntityId,
	position Position,
	value uint32,
	active bool,
) {
	w.Nutrients[id] = Nutrient{}
	w.Positions[id] = position
	w.NutrientValues[id] = NutrientValue{Value: value}
	w.Active[id] = Active{IsActive: active}
}

func (w *World) CreateWall(id EntityId, wallConfig config.WallConfig) {
	w.Walls[id] = Wall{}
	w.WallStates[id] = WallState{Open: wallConfig.Open}
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
