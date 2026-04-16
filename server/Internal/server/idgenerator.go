package server

import "sync/atomic"

type IdGenerator struct {
	next atomic.Uint64
}

func (idGenerator *IdGenerator) Next() uint64 {
	return idGenerator.next.Add(1)
}
