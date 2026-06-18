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
	var entity Entity
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

}

func (world *World) moveEntity(entity Entity, target *archetype, values map[ComponentID]any) error {

}

func (world *World) removeFromCurrentArchetype(entity Entity, slot *entitySlot) {

}
