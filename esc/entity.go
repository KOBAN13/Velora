package esc

type EntityId uint64

type EntityAllocator interface {
	Next() uint64
}
