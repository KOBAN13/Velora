package esc

import "Velora/server/Internal"

type EntityId uint64

type EntityKind uint8

const (
	EntityKindUnknown EntityKind = iota
	EntityKindPlayerCell
	EntityKindCore
	EntityKindNutrient
	EntityKindWall
)

type Entity interface {
	EntityID() EntityId
	EntityKind() EntityKind
}

type EntityAllocator interface {
	Next() EntityId
}

type EntityIdAllocator struct {
	Generator *Internal.IdGenerator
}

func (a EntityIdAllocator) Next() EntityId {
	return EntityId(a.Generator.Next())
}
