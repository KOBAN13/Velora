package match

type EntityId uint64

type World struct {
	PlayerCells map[EntityId]PlayerCell
	Cores       map[EntityId]Core
	Nutrients   map[EntityId]Nutrient
	Walls       map[EntityId]Wall
}

type Position struct {
	X float32
	Y float32
}

type PlayerCell struct {
	ID       EntityId
	OwnerId  uint64
	Position Position
	Health   Health
}

type Health struct {
	HP int32
}

type Active struct {
	IsActive bool
}

type Core struct {
	ID            EntityId
	OwnerId       uint64
	Position      Position
	CurrentHealth Health
	MaxHP         Health
}

type Nutrient struct {
	ID     EntityId
	Pos    Position
	Value  uint32
	Active Active
}

type Wall struct {
	ID   EntityId
	Open bool
}
