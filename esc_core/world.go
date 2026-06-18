package esc_core

import (
	"fmt"
	"reflect"
)

type WorldOption func(*World)

type World struct {
	slots             []entitySlot
	freeEntityIndexes []uint32
	archetypes        []*archetype
	archetypeByKey    map[string]*archetype
	registry          *ComponentRegistry
	mutationPhase     MutationPhase
	archetypeVersion  uint64
}

func WithEntityCapacity(capacity int) WorldOption {
	return func(world *World) {
		world.slots = make([]entitySlot, 0, capacity)
	}
}

func NewWorld(options ...WorldOption) *World {
	var world = &World{
		archetypeByKey: make(map[string]*archetype),
		registry:       NewComponentRegistry(),
	}

	for _, option := range options {
		option(world)
	}

	return world
}

func Spawn(world *World, components ...any) (Entity, error) {
	if world.mutationPhase == MutationRunningSystem {
		return Entity{}, ErrInvalidMutationPhase
	}

	values, signature, err := world.collectComponentValues(components)
	if err != nil {
		return Entity{}, err
	}

	archetype, err := world.archetypeFor(signature)
	if err != nil {
		return Entity{}, err
	}

	entity := world.allocateEntity()

	row, err := archetype.appendEntity(entity, values)
	if err != nil {
		world.releaseEntity(entity)
		return Entity{}, err
	}

	world.slots[entity.index].location = entityLocation{
		archetype: archetype,
		row:       row,
	}

	return entity, nil
}

func (world *World) allocateEntity() Entity {
	if n := len(world.freeEntityIndexes); n > 0 {
		var index = world.freeEntityIndexes[n-1]
		world.freeEntityIndexes = world.freeEntityIndexes[:n-1]
		var slot = &world.slots[index]
		slot.alive = true
		return Entity{index: index, generation: slot.generation}
	}

	var index = uint32(len(world.slots))
	world.slots = append(world.slots, entitySlot{alive: true})
	return Entity{index: index}
}

func (world *World) releaseEntity(entity Entity) {
	var slot = &world.slots[entity.index]
	slot.alive = false
	slot.location = entityLocation{}
	slot.generation++
	world.freeEntityIndexes = append(world.freeEntityIndexes, entity.index)
}

func (world *World) validateAlive(entity Entity) (*entitySlot, error) {
	if int(entity.index) >= len(world.slots) {
		return nil, ErrInvalidEntity
	}

	slot := &world.slots[entity.index]

	if !slot.alive || slot.generation != entity.generation {
		return nil, ErrInvalidEntity
	}

	return slot, nil
}

func (world *World) collectComponentValues(components []any) (map[ComponentID]any, componentSignature, error) {
	var values = make(map[ComponentID]any, len(components))
	var ids = make([]ComponentID, 0, len(components))

	for _, component := range components {
		if component == nil {
			return nil, componentSignature{}, ErrInvalidComponentType
		}

		var token = ComponentToken{
			Type: reflect.TypeOf(component),
			Name: reflect.TypeOf(component).Name(),
		}

		var info, err = world.registry.Info(token)

		if err != nil {
			return nil, componentSignature{}, err
		}

		if _, exists := values[info.Id]; exists {
			return nil, componentSignature{}, fmt.Errorf("%w: %s", ErrDuplicateComponent, info.Name)
		}

		values[info.Id] = component
		ids = append(ids, info.Id)
	}

	var signature = newComponentSignature(ids)

	return values, *signature, nil
}

func (world *World) archetypeFor(signature componentSignature) (*archetype, error) {
	if archetype, ok := world.archetypeByKey[signature.key]; ok {
		return archetype, nil
	}

	var columns = make(map[ComponentID]column, len(signature.ids))

	for _, id := range signature.ids {
		var info, ok = world.registry.InfoById(id)

		if !ok {
			return nil, fmt.Errorf("%w: component id %d", ErrComponentNotFound, id)
		}

		columns[id] = newReflectColumn(info.Type)
	}

	archetype := newArchetype(
		archetypeID(len(world.archetypes)),
		signature,
		columns,
	)

	world.archetypes = append(world.archetypes, archetype)
	world.archetypeByKey[signature.key] = archetype
	world.archetypeVersion++

	return archetype, nil
}

func (world *World) moveEntity(entity Entity, target *archetype, values map[ComponentID]any) error {
	slot, err := world.validateAlive(entity)

	if err != nil {
		return err
	}

	var source = slot.location.archetype
	var sourceRow = slot.location.row

	var targetValues = make(map[ComponentID]any, len(target.signature.ids))

	for _, id := range target.signature.ids {
		if value, ok := values[id]; ok {
			targetValues[id] = value
			continue
		}

		var sourceColumn, ok = source.column(id)

		if !ok {
			continue
		}

		targetValues[id] = sourceColumn.ValueAny(sourceRow)
	}

	targetRow, err := target.appendEntity(entity, targetValues)

	if err != nil {
		return err
	}

	if source != nil {
		moved, movedRow, hadMove := source.removeEntity(sourceRow)
		if hadMove {
			world.slots[moved.index].location = entityLocation{
				archetype: source,
				row:       movedRow,
			}
		}
	}

	slot.location = entityLocation{
		archetype: target,
		row:       targetRow,
	}

	return nil
}

func (world *World) removeFromCurrentArchetype(slot *entitySlot) {
	var source = slot.location.archetype

	var moved, movedRow, hadMove = source.removeEntity(slot.location.row)

	if hadMove {
		world.slots[moved.index].location = entityLocation{
			archetype: source,
			row:       movedRow,
		}
	}

	slot.location = entityLocation{}
}
