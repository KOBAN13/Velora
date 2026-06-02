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
	ErrPlayerNotFound = errors.New("player not found")
)

type EntityId uint64

type NutrientSpawner struct {
	rng           *rand.Rand
	LastSpawnTick uint64
	SpawnInterval uint64
	MaxNutrients  int
	SpawnBatch    int
	MaxAttempts   int
}

type World struct {
	PlayerCells map[EntityId]PlayerCell
	Cores       map[EntityId]Core
	Nutrients   map[EntityId]Nutrient
	Walls       map[EntityId]Wall

	nutrientSpawner NutrientSpawner
}

type Position struct {
	X float32
	Y float32
}

type PlayerCell struct {
	ID        EntityId
	OwnerId   uint64
	Position  Position
	Health    Health
	MaxHealth Health
	Biomass   uint32
	Level     uint32
	Alive     Active
}

type Health struct {
	HP int32
}

type Active struct {
	IsActive bool
}

type Core struct {
	ID            EntityId
	OwnerId       uint64
	Position      Position
	CurrentHealth Health
	MaxHP         Health
}

type Nutrient struct {
	ID     EntityId
	Pos    Position
	Value  uint32
	Active Active
}

type Wall struct {
	ID   EntityId
	Open bool
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

func NewWorld(players []PlayerRef, config config.GameConfig, idGenerator *Internal.IdGenerator) (*World, error) {
	if len(players) == 0 {
		return nil, ErrPlayerNotFound
	}

	var playersCells = make(map[EntityId]PlayerCell, len(players))
	var core = make(map[EntityId]Core, len(players))
	var walls = make(map[EntityId]Wall, len(players))

	slices.SortFunc(players, func(a, b PlayerRef) int {
		return cmp.Compare(a.Slot, b.Slot)
	})

	for _, player := range players {
		var playerCell = PlayerCell{
			ID:       EntityId(idGenerator.Next()),
			OwnerId:  player.UserId,
			Position: startPositions[player.Slot].Cell,
			Health: Health{
				HP: int32(config.PlayerCell.HP),
			},
			MaxHealth: Health{
				HP: int32(config.PlayerCell.MaxHP),
			},
			Biomass: uint32(config.PlayerCell.Biomass),
			Level:   uint32(config.PlayerCell.Level),
			Alive:   Active{IsActive: true},
		}

		var coreCell = Core{
			ID:       EntityId(idGenerator.Next()),
			OwnerId:  player.UserId,
			Position: startPositions[player.Slot].Core,
		}

		var wall = Wall{
			ID:   EntityId(idGenerator.Next()),
			Open: config.Wall.Open,
		}

		walls[wall.ID] = wall
		playersCells[playerCell.ID] = playerCell
		core[coreCell.ID] = coreCell
	}

	return &World{
		PlayerCells: playersCells,
		Cores:       core,
		Nutrients:   make(map[EntityId]Nutrient),
		Walls:       walls,
	}, nil
}
