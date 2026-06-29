package systems

import (
	"Velora/esc"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

func (s *NutrientSystem) spawn(
	ctx *esc_core.Context,
	spawner *esc.NutrientSpawnerResource,
	count int,
	inactive []nutrientSlot,
	newLimit int,
) (int, error) {
	if count <= 0 {
		return 0, nil
	}

	if newLimit < 0 {
		newLimit = 0
	}

	if count > len(inactive)+newLimit {
		count = len(inactive) + newLimit
	}

	spawned := 0
	attempts := 0
	reserved := make([]esc.Position, 0, count)
	maxAttempts := count * spawner.MaxAttempts

	for spawned < count && attempts < maxAttempts {
		attempts++

		position := randomPosition(spawner)
		canPlace, err := s.canPlace(spawner, position, reserved)

		if err != nil {
			return spawned, err
		}

		if !canPlace {
			continue
		}

		if spawned < len(inactive) {
			respawnNutrient(inactive[spawned], position, spawner.NutrientValue, spawner.NutrientActive)
		} else if err := ctx.Commands.Spawn().
			Bundle(esc.NewNutrientBundle(position, spawner.NutrientValue, spawner.NutrientActive)).
			Commit(); err != nil {
			return spawned, err
		}

		reserved = append(reserved, position)
		spawned++
	}

	return spawned, nil
}

func (s *NutrientSystem) canPlace(
	spawner *esc.NutrientSpawnerResource,
	position esc.Position,
	reserved []esc.Position,
) (bool, error) {
	if ok, err := s.farFromPlayers(spawner, position); err != nil || !ok {
		return ok, err
	}

	if ok, err := s.farFromCores(spawner, position); err != nil || !ok {
		return ok, err
	}

	if ok, err := s.farFromActiveNutrients(spawner, position); err != nil || !ok {
		return ok, err
	}

	for _, reservedPosition := range reserved {
		if distanceSquared(position, reservedPosition) < square(spawner.MinNutrientDistance) {
			return false, nil
		}
	}

	return true, nil
}

func (s *NutrientSystem) farFromPlayers(spawner *esc.NutrientSpawnerResource, position esc.Position) (bool, error) {
	players := s.players.Iter()

	for players.Next() {
		playerPosition, err := esc_core.Read[esc.Position](players)

		if err != nil {
			return false, err
		}

		if distanceSquared(position, playerPosition) < square(spawner.MinPlayerDistance) {
			return false, nil
		}
	}

	return true, nil
}

func (s *NutrientSystem) farFromCores(spawner *esc.NutrientSpawnerResource, position esc.Position) (bool, error) {
	cores := s.cores.Iter()

	for cores.Next() {
		corePosition, err := esc_core.Read[esc.Position](cores)

		if err != nil {
			return false, err
		}

		if distanceSquared(position, corePosition) < square(spawner.MinCoreDistance) {
			return false, nil
		}
	}

	return true, nil
}

func (s *NutrientSystem) farFromActiveNutrients(spawner *esc.NutrientSpawnerResource, position esc.Position) (bool, error) {
	nutrients := s.nutrients.Iter()

	for nutrients.Next() {
		active, err := esc_core.Write[esc.Active](nutrients)

		if err != nil {
			return false, err
		}

		if !active.IsActive {
			continue
		}

		nutrientPosition, err := esc_core.Write[esc.Position](nutrients)

		if err != nil {
			return false, err
		}

		if distanceSquared(position, *nutrientPosition) < square(spawner.MinNutrientDistance) {
			return false, nil
		}
	}

	return true, nil
}

func (s *NutrientSystem) nutrientStats() (activeCount int, totalCount int, inactive []nutrientSlot, err error) {
	nutrients := s.nutrients.Iter()

	for nutrients.Next() {
		totalCount++

		active, err := esc_core.Write[esc.Active](nutrients)

		if err != nil {
			return 0, 0, nil, err
		}

		if active.IsActive {
			activeCount++
			continue
		}

		position, err := esc_core.Write[esc.Position](nutrients)

		if err != nil {
			return 0, 0, nil, err
		}

		value, err := esc_core.Write[esc.NutrientValue](nutrients)

		if err != nil {
			return 0, 0, nil, err
		}

		inactive = append(inactive, nutrientSlot{
			position: position,
			value:    value,
			active:   active,
		})
	}

	return activeCount, totalCount, inactive, nil
}

type nutrientSlot struct {
	position *esc.Position
	value    *esc.NutrientValue
	active   *esc.Active
}

func respawnNutrient(slot nutrientSlot, position esc.Position, value uint32, active bool) {
	*slot.position = position
	slot.value.Value = value
	slot.active.IsActive = active
}

func randomPosition(spawner *esc.NutrientSpawnerResource) esc.Position {
	size := spawner.ArenaHalfSize * 2

	return esc.Position{
		X: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
		Y: spawner.Rng.Float32()*size - spawner.ArenaHalfSize,
	}
}

func distanceSquared(a esc.Position, b esc.Position) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y

	return dx*dx + dy*dy
}

func square(value float32) float32 {
	return value * value
}
