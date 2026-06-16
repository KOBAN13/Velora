package esc_core

type World struct {
	slots             []entitySlot
	freeEntityIndexes []uint32
	archetypes        []*archetype
	archetypeByKey    map[string]*archetype
	registry          *ComponentRegistry
	mutationPhase     MutationPhase
	archetypeVersion  uint64
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
