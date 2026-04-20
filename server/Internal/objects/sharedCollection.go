package objects

import (
	"Velora/server/Internal"
	"maps"
	"sync"
)

type SharedCollection[T any] struct {
	objectsMap map[uint64]T
	nextID     uint64
	mapMutex   sync.Mutex
}

func NewSharedCollection[T any](idGenerator *Internal.IdGenerator, capacity ...int) *SharedCollection[T] {
	var objectsMap map[uint64]T

	if len(capacity) > 0 {
		objectsMap = make(map[uint64]T, capacity[0])
	} else {
		objectsMap = make(map[uint64]T)
	}

	return &SharedCollection[T]{
		objectsMap: objectsMap,
		nextID:     idGenerator.Next(),
	}
}

func (c *SharedCollection[T]) Add(object T, idGenerator *Internal.IdGenerator) uint64 {
	defer c.mapMutex.Unlock()
	c.mapMutex.Lock()

	var id = idGenerator.Next()

	c.objectsMap[id] = object

	return id
}

func (c *SharedCollection[T]) Remove(id uint64) {
	defer c.mapMutex.Unlock()
	c.mapMutex.Lock()

	delete(c.objectsMap, id)
}

func (c *SharedCollection[T]) Get(id uint64) (T, bool) {
	defer c.mapMutex.Unlock()
	c.mapMutex.Lock()

	var object, isOk = c.objectsMap[id]

	return object, isOk
}

func (c *SharedCollection[T]) Foreach(callback func(T, uint64)) {
	c.mapMutex.Lock()

	var localMap = make(map[uint64]T, len(c.objectsMap))

	maps.Copy(localMap, c.objectsMap)

	c.mapMutex.Unlock()

	for id, object := range localMap {
		callback(object, id)
	}
}

func (c *SharedCollection[T]) Size() int {
	defer c.mapMutex.Unlock()
	c.mapMutex.Lock()
	return len(c.objectsMap)
}
