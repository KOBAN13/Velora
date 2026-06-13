package spawners

import (
	"math/rand/v2"

	"Velora/esc"
	"Velora/server/Internal/server/config"
)

const (
	defaultNutrientMaxAttempts         = 1500
	defaultNutrientArenaHalfSize       = 22
	defaultNutrientMinPlayerDistance   = 4
	defaultNutrientMinCoreDistance     = 5
	defaultNutrientMinNutrientDistance = 2
	DefaultNutrientPickUpDistance      = 1.5
)

type NutrientSpawner = esc.NutrientSpawnerResource

type startPosition struct {
	Cell esc.Position
	Core esc.Position
}

var StartPositions = []startPosition{
	{Cell: esc.Position{X: -10, Y: 0}, Core: esc.Position{X: -14, Y: 0}},
	{Cell: esc.Position{X: 10, Y: 0}, Core: esc.Position{X: 14, Y: 0}},
	{Cell: esc.Position{X: 0, Y: 10}, Core: esc.Position{X: 0, Y: 14}},
	{Cell: esc.Position{X: 0, Y: -10}, Core: esc.Position{X: 0, Y: -14}},
	{Cell: esc.Position{X: 10, Y: 10}, Core: esc.Position{X: 14, Y: 14}},
}

func NewNutrientSpawn(gameConfig config.GameConfig, mapSeed uint64) (NutrientSpawner, error) {
	var spawner = NewNutrientSpawner(mapSeed, gameConfig.Nutrient)

	return spawner, nil
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
