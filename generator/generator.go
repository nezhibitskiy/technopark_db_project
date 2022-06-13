package generator

import "sync/atomic"

type Generator struct {
	current uint32
}

func NewGenerator() Generator {
	return Generator{}
}

func (s *Generator) Next() uint32 {
	return atomic.AddUint32(&s.current, 1)
}
