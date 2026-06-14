package systems

import (
	"Velora/esc"
	"errors"
)

var (
	ErrNutrientSpawn = errors.New("nutrient spawn failed")
)

const (
	DefaultNutrientPickUpDistance = 1.5
)

type NutrientSystem struct{}

func (*NutrientSystem) Name() string {
	return "NutrientSystem"
}

func (*NutrientSystem) Stage() Stage {
	return StageSpawn
}

func NewNutrientSystem() *NutrientSystem {
	return &NutrientSystem{}
}

func (s *NutrientSystem) Update(ctx *esc.SystemContext, world *esc.World) {
	s.PlayerPickUpNutrient(world)
	s.SpawnForTick(world, ctx)
}

func (s *NutrientSystem) Start(ctx *esc.SystemContext, world *esc.World) error {
	if err := s.Fill(ctx, world); err != nil {
		return err
	}

	return nil
}

func (s *NutrientSystem) Fill(ctx *esc.SystemContext, world *esc.World) error {
	var spawner = ctx.Resources.NutrientSpawner
	var missing = spawner.MaxNutrients - s.ActiveNutrientCount(world)

	if missing <= 0 {
		return nil
	}

	if s.spawn(ctx, world, missing) != missing {
		return ErrNutrientSpawn
	}

	return nil
}

func (s *NutrientSystem) PlayerPickUpNutrient(world *esc.World) {
	for _, player := range world.QueryActivePlayerCells() {
		var playerPosition = player.Position
		for _, nutrient := range world.QueryActiveNutrients() {

			var nutrientPosition = nutrient.Position

			var distance = playerPosition.DistanceTo(nutrientPosition)

			if distance <= DefaultNutrientPickUpDistance {
				world.SetActive(nutrient.EntityID(), esc.Active{IsActive: false})

				var nutrientValue = nutrient.Value.Value
				var playerBiomass = player.Biomass.Value

				world.SetBiomass(player.EntityID(), esc.Biomass{Value: nutrientValue + playerBiomass})
				world.SetLevel(player.EntityID(), esc.Level{Value: 1 + player.Biomass.Value/100})
			}
		}
	}
}

func (s *NutrientSystem) SpawnForTick(world *esc.World, ctx *esc.SystemContext) {
	var spawner = ctx.Resources.NutrientSpawner

	var tick = float64(ctx.Tick)

	if spawner.SpawnInterval > 0 && tick-spawner.LastSpawnTick < spawner.SpawnInterval {
		return
	}

	spawner.LastSpawnTick = tick

	var missing = spawner.MaxNutrients - s.ActiveNutrientCount(world)

	if missing <= 0 {
		return
	}

	if spawner.SpawnBatch <= 0 {
		return
	}

	s.spawn(ctx, world, min(missing, spawner.SpawnBatch))
}

func (s *NutrientSystem) spawn(ctx *esc.SystemContext, world *esc.World, count int) int {
	if count <= 0 {
		return 0
	}

	var spawned = 0
	var attempts = 0
	var spawner = ctx.Resources.NutrientSpawner
	var reservedPositions = make([]esc.Position, 0, count)
	var inactiveNutrients = world.QueryInactiveNutrients()
	var newNutrientLimit = spawner.MaxNutrients - len(world.QueryNutrients())

	if newNutrientLimit < 0 {
		newNutrientLimit = 0
	}

	if count > len(inactiveNutrients)+newNutrientLimit {
		count = len(inactiveNutrients) + newNutrientLimit
	}

	var maxAttempts = count * spawner.MaxAttempts

	for spawned < count && attempts < maxAttempts {
		attempts++

		var position = s.randomPosition(ctx)

		if !s.canPlace(ctx, world, position, reservedPositions) {
			continue
		}

		if spawned < len(inactiveNutrients) {
			ctx.Commands.Add(&esc.RespawnNutrientCommand{
				EntityId: inactiveNutrients[spawned].EntityID(),
				Position: position,
				Value:    spawner.NutrientValue,
				Active:   spawner.NutrientActive,
			})
		} else {
			ctx.Commands.Add(&esc.SpawnNutrientCommand{
				Position: position,
				Value:    spawner.NutrientValue,
				Active:   spawner.NutrientActive,
			})
		}

		reservedPositions = append(reservedPositions, position)

		spawned++
	}

	return spawned
}

func (s *NutrientSystem) randomPosition(ctx *esc.SystemContext) esc.Position {
	var spawner = ctx.Resources.NutrientSpawner
	var size = spawner.ArenaHalfSize * 2

	return esc.Position{
		X: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
		Y: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
	}
}

func (s *NutrientSystem) canPlace(ctx *esc.SystemContext, world *esc.World, position esc.Position, reservedPositions []esc.Position) bool {
	var spawner = ctx.Resources.NutrientSpawner

	for _, player := range world.QueryPlayerCells() {
		if distanceSquared(position, player.Position) < square(spawner.MinPlayerDistance) {
			return false
		}
	}

	for _, core := range world.QueryCores() {
		if distanceSquared(position, core.Position) < square(spawner.MinCoreDistance) {
			return false
		}
	}

	for _, nutrient := range world.QueryActiveNutrients() {
		if distanceSquared(position, nutrient.Position) < square(spawner.MinNutrientDistance) {
			return false
		}
	}

	for _, reservedPosition := range reservedPositions {
		if distanceSquared(position, reservedPosition) < square(spawner.MinNutrientDistance) {
			return false
		}
	}

	return true
}

func (s *NutrientSystem) ActiveNutrientCount(world *esc.World) int {
	var count = 0

	for range world.QueryActiveNutrients() {
		count++
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
