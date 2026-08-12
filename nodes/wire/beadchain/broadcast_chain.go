package beadchain

import wire "github.com/dtauraso/wirefold/nodes/wire"

type BeadGeometryIn struct {
	Center wire.Vec3
	Aim    wire.Vec3
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
