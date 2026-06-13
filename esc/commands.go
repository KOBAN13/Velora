package esc

type CommandBuffer struct {
	commands []Command
}

type Command interface {
	Apply(world *World, resources *Resources)
}

func (c *CommandBuffer) Add(command Command) {
	c.commands = append(c.commands, command)
}

func (c *CommandBuffer) Execute(world *World, resources *Resources) {

}

func (c *CommandBuffer) Clear(world *World, resources *Resources) {
	clear(c.commands)
	c.commands = c.commands[:0]
}

type SpawnNutrientCommand struct {
	Position Position
	Value    uint32
	Active   bool
}

func (c *SpawnNutrientCommand) Apply(world *World, resources *Resources) {
	world.CreateNutrient(resources.EntityIds.Next(), c.Position, c.Value, c.Active)
}

type RemoveEntityCommand struct {
	EntityId EntityId
}

func (c *RemoveEntityCommand) Apply(world *World, resources *Resources) {
	world.RemoveEntity(c.EntityId)
}

type SetActiveCommand struct {
	EntityId EntityId
	Active   bool
}

func (c *SetActiveCommand) Apply(world *World, resources *Resources) {
	world.SetActive(c.EntityId, Active{c.Active})
}
