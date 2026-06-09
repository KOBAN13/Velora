package systems

import (
	"Velora/esc"
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

func (s *NutrientSystem) Update(tick float64, world *esc.World) {
	s.PlayerPickUpNutrient(world)
	s.SpawnForTick(world, tick)
}

func (s *NutrientSystem) Start(world *esc.World) error {
	if err := s.Fill(world); err != nil {
		return err
	}

	return nil
}

func (s *NutrientSystem) Fill(world *esc.World) error {
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

func (s *NutrientSystem) PlayerPickUpNutrient(world *esc.World) {
	for entityId := range world.PlayerCells {
		if !world.Active[entityId].IsActive {
			continue
		}

		var playerPosition = world.Positions[entityId]

		for nutrientId := range world.Nutrients {
			if !world.Active[nutrientId].IsActive {
				continue
			}

			var nutrientPosition = world.Positions[nutrientId]

			var distance = playerPosition.DistanceTo(nutrientPosition)

			if distance <= match.DefaultNutrientPickUpDistance {
				var activeComponent = world.Active[nutrientId]
				activeComponent.IsActive = false
				world.Active[nutrientId] = activeComponent

				var nutrientValue = world.NutrientValues[nutrientId].Value
				var biomassComponent = world.Biomass[entityId]

				biomassComponent.Value += nutrientValue

				world.Biomass[entityId] = biomassComponent

				var levelComponent = world.Levels[entityId]

				levelComponent.Value = 1 + biomassComponent.Value/100

				world.Levels[entityId] = levelComponent
			}
		}
	}
}

func (s *NutrientSystem) SpawnForTick(world *esc.World, serverTick float64) {
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

func (s *NutrientSystem) spawn(world *esc.World, count int) int {
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

		world.CreateNutrient(esc.EntityId(s.match.EntityIds.Next()), position, 0, spawner.NutrientActive)
		spawned++
	}

	return spawned
}

func (s *NutrientSystem) randomPosition() esc.Position {
	var spawner = &s.match.NutrientSpawner
	var size = spawner.ArenaHalfSize * 2

	return esc.Position{
		X: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
		Y: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
	}
}

func (s *NutrientSystem) canPlace(world *esc.World, position esc.Position) bool {
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

func (s *NutrientSystem) ActiveNutrientCount(world *esc.World) int {
	var count = 0

	for id := range world.Nutrients {
		if world.Active[id].IsActive {
			count++
		}
	}

	return count
}

func distanceSquared(a esc.Position, b esc.Position) float32 {
	var dx = a.X - b.X
	var dy = a.Y - b.Y

	return dx*dx + dy*dy
}

func square(value float32) float32 {
	return value * value
}
