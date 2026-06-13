package esc

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
