package esc

type CommandBuffer struct {
	commands []Command
}

type Command interface {
	Apply(world *World, resources *Resources) error
}

type SpawnNutrientCommand struct {
	Position Position
	Value    uint32
	Active   bool
}

type RemoveEntityCommand struct {
	EntityId EntityId
}

type SetActiveCommand struct {
	EntityId EntityId
	Active   bool
}
