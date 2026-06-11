package esc

type PlayerCell struct {
	Id        EntityId
	OwnerId   uint64
	Position  Position
	Direction MoveDirection
	HP        int32
	MaxHP     int32
	Biomass   uint32
	Level     uint32
	Active    bool
}

func (e *PlayerCell) EntityID() EntityId {
	return e.Id
}

func (e *PlayerCell) EntityKind() EntityKind {
	return EntityKindPlayerCell
}

type Core struct {
	Id       EntityId
	OwnerId  uint64
	Position Position
	HP       int32
	MaxHP    int32
}

func (e *Core) EntityID() EntityId {
	return e.Id
}

func (e *Core) EntityKind() EntityKind {
	return EntityKindCore
}

type Nutrient struct {
	Id       EntityId
	Position Position
	Value    uint32
	Active   bool
}

func (e *Nutrient) EntityID() EntityId {
	return e.Id
}

func (e *Nutrient) EntityKind() EntityKind {
	return EntityKindNutrient
}

type Wall struct {
	Id   EntityId
	Open bool
}

func (e *Wall) EntityID() EntityId {
	return e.Id
}

func (e *Wall) EntityKind() EntityKind {
	return EntityKindWall
}
