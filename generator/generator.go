package generator

import "sync/atomic"

type Generator struct {
	current uint32
}

func NewGenerator() Generator {
	return Generator{}
}

func (s *Generator) Next() uint {
	atomic.AddUint32(&s.current, 1)

	return uint(s.current)
}
