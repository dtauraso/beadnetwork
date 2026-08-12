// broadcast_chain.go — the lock-free broadcast primitive bead_actor.go's Bead is driven by:
// one node-owned generation chain, closed once per event, that lets every waiting Bead wake
// in a single hop with no lock/atomic. Lifted out of bead_actor.go (which keeps the guarded
// Bead.run select — see that file's header) because BroadcastChain is a self-contained,
// reusable primitive with its own construction/advance API, not part of the bead actor's own
// state machine: BeadGeometryIn/BroadcastChain have no dependency on Bead, while Bead depends
// on them (three of Bead's own fields are *BroadcastChain). The dependency runs one way, so
// this is the primitive Bead is built on, not a second copy of Bead's own concern.
package beadchain

import wire "github.com/dtauraso/wirefold/nodes/wire"

// BeadGeometryIn is the payload a node broadcasts to every bead on its edges each time its
// own transform changes: the one thing a bead's position is computed FROM. It carries
// enough for a bead to place itself directly (body force) — this node's live world
// center and its live unit aim toward the edge's far end — with no neighbour-of-neighbour
// read and no relaxation pass.
type BeadGeometryIn struct {
	Center wire.Vec3
	Aim    wire.Vec3 // unit direction from Center toward the edge's far node
}

// BroadcastChain is a lock-free, single-writer broadcast primitive: ONE close wakes every
// goroutine blocked on Fire, no matter how many, and no send-loop ever iterates the
// receivers. Because a Go channel can only be closed once, each "next generation" is a
// FRESH BroadcastChain — Next is written by the single owning (node) goroutine BEFORE Fire
// is closed, so a receiver that wakes on <-Fire can safely read Next afterward with no lock
// or atomic: Go's memory model guarantees a close on a channel happens-before the
// corresponding receive observes it, and Next was written strictly before the close that
// wakes the receiver, so publication is safe by construction, not by convention.
type BroadcastChain struct {
	Fire  chan struct{}
	Value BeadGeometryIn // meaningful only for geometry-generation chains; zero otherwise
	Next  *BroadcastChain
}

// NewBroadcastChain returns a fresh, unfired chain link, ready to be closed later.
func NewBroadcastChain() *BroadcastChain {
	return &BroadcastChain{Fire: make(chan struct{})}
}

// Advance closes c (waking every goroutine blocked on c.Fire, in one operation regardless
// of how many are waiting) and returns the fresh NEXT link, already wired into c.Next
// before the close so waiters can chain forward. Only the owning (node) goroutine may call
// this — see the type doc for why that single-writer discipline is what makes Next safe to
// read lock-free.
func (c *BroadcastChain) Advance() *BroadcastChain {
	next := NewBroadcastChain()
	c.Next = next
	close(c.Fire)
	return next
}

// AdvanceWithValue is Advance, but stamps a payload (the node's live transform) into the
// CURRENT link before firing, for waiters that need the value the broadcast carries (the
// geometry hop) rather than just the fact of it (wake/settle, which carry no payload).
func (c *BroadcastChain) AdvanceWithValue(v BeadGeometryIn) *BroadcastChain {
	c.Value = v
	return c.Advance()
}
