package esc

type PlayerCell struct {
	Id        EntityId
	OwnerId   Owner
	Position  Position
	Direction MoveDirection
	HP        Health
	MaxHP     Health
	Biomass   Biomass
	Level     Level
	Active    Active
}

func (e *PlayerCell) EntityID() EntityId {
	return e.Id
}

func (e *PlayerCell) EntityKind() EntityKind {
	return EntityKindPlayerCell
}

type Core struct {
	Id       EntityId
	OwnerId  Owner
	Position Position
	HP       Health
	MaxHP    Health
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
	Value    NutrientValue
	Active   Active
}

func (e *Nutrient) EntityID() EntityId {
	return e.Id
}

func (e *Nutrient) EntityKind() EntityKind {
	return EntityKindNutrient
}

type Wall struct {
	Id   EntityId
	Open WallState
}

func (e *Wall) EntityID() EntityId {
	return e.Id
}

func (e *Wall) EntityKind() EntityKind {
	return EntityKindWall
}
