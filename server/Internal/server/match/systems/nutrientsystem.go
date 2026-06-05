package systems

import (
	"Velora/server/Internal/server/match"
	"errors"
)

var (
	ErrNutrientSpawn = errors.New("nutrient spawn failed")
)

type NutrientSystem struct {
	match *match.Match
}

func NewNutrientSystem(match *match.Match) *NutrientSystem {
	return &NutrientSystem{
		match: match,
	}
}

func (s *NutrientSystem) Update(tick float64, world *match.World) {
	s.SpawnForTick(world, tick)
}

func (s *NutrientSystem) Start(world *match.World) error {
	if err := s.Fill(world); err != nil {
		return err
	}

	return nil
}

func (s *NutrientSystem) Fill(world *match.World) error {
	var spawner = &s.match.NutrientSpawner
	var missing = spawner.MaxNutrients - s.ActiveNutrientCount(world)

	if missing <= 0 {
		return nil
	}

	if s.spawn(world, missing) != missing {
		return ErrNutrientSpawn
	}

	return nil
}

func (s *NutrientSystem) SpawnForTick(world *match.World, serverTick float64) {
	var spawner = &s.match.NutrientSpawner

	if spawner.SpawnInterval > 0 && serverTick-spawner.LastSpawnTick < spawner.SpawnInterval {
		return
	}

	spawner.LastSpawnTick = serverTick

	var missing = spawner.MaxNutrients - s.ActiveNutrientCount(world)

	if missing <= 0 {
		return
	}

	if spawner.SpawnBatch <= 0 {
		return
	}

	s.spawn(world, min(missing, spawner.SpawnBatch))
}

func (s *NutrientSystem) spawn(world *match.World, count int) int {
	if count <= 0 {
		return 0
	}

	var spawned = 0
	var attempts = 0
	var spawner = &s.match.NutrientSpawner
	var maxAttempts = count * spawner.MaxAttempts

	for spawned < count && attempts < maxAttempts {
		attempts++

		var position = s.randomPosition()
		if !s.canPlace(world, position) {
			continue
		}

		world.CreateNutrient(match.EntityId(s.match.EntityIds.Next()), position, 0, spawner.NutrientActive)
		spawned++
	}

	return spawned
}

func (s *NutrientSystem) randomPosition() match.Position {
	var spawner = &s.match.NutrientSpawner
	var size = spawner.ArenaHalfSize * 2

	return match.Position{
		X: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
		Y: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
	}
}

func (s *NutrientSystem) canPlace(world *match.World, position match.Position) bool {
	var spawner = &s.match.NutrientSpawner

	for id := range world.PlayerCells {
		if distanceSquared(position, world.Positions[id]) < square(spawner.MinPlayerDistance) {
			return false
		}
	}

	for id := range world.Cores {
		if distanceSquared(position, world.Positions[id]) < square(spawner.MinCoreDistance) {
			return false
		}
	}

	for id := range world.Nutrients {
		if !world.Active[id].IsActive {
			continue
		}

		if distanceSquared(position, world.Positions[id]) < square(spawner.MinNutrientDistance) {
			return false
		}
	}

	return true
}

func (s *NutrientSystem) ActiveNutrientCount(world *match.World) int {
	var count = 0

	for id := range world.Nutrients {
		if world.Active[id].IsActive {
			count++
		}
	}

	return count
}

func distanceSquared(a match.Position, b match.Position) float32 {
	var dx = a.X - b.X
	var dy = a.Y - b.Y

	return dx*dx + dy*dy
}

func square(value float32) float32 {
	return value * value
}
