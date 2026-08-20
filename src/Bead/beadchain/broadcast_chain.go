package beadchain

import "github.com/dtauraso/wirefold/src/Node/spatial"

type BeadGeometryIn struct {
	Center spatial.Vec3
	Aim    spatial.Vec3
}

type BroadcastChain struct {
	Fire  chan struct{}
	Value BeadGeometryIn
	Next  *BroadcastChain
}

func NewBroadcastChain() *BroadcastChain {
	return &BroadcastChain{Fire: make(chan struct{})}
}

func (c *BroadcastChain) Advance() *BroadcastChain {
	next := NewBroadcastChain()
	c.Next = next
	close(c.Fire)
	return next
}

func (c *BroadcastChain) AdvanceWithValue(v BeadGeometryIn) *BroadcastChain {
	c.Value = v
	return c.Advance()
}
