package Buffer

// bufLayoutChainBead defines one row of the chain-bead column block — the node-owned
// placeholder sequence that IS the visual of an edge (docs/bead-model/beads-are-the-edge.md). One row
// per placeholder bead on one of this node's OUTGOING edges, in that edge's order outward
// from this node.
//
// OX/OY/OZ are NODE-LOCAL, exactly like the Interior block's: the offset from this node's
// own center, with the renderer adding the center to get the world position. That is not the
// renderer owning positions (Go owns the offsets); it is what makes moving a node constant
// time, because the offsets do not change when the node's center does.
//
// A chain bead has no absolute position column ON PURPOSE. The old moving bead (the now-gone
// Bead block, see layout.go) carried recomputed world X/Y/Z every tick; a chain does not
// move, so there is nothing to recompute.
//
// NOTE: no bead position here depends on another bead's position. That is the line separating
// this from the reverted bead-chain wire (memory/project_wire_is_straight_line_not_chain.md),
// whose spacing came from neighbour midpoints and therefore followed a drag in O(N²). Each
// offset is index × spacing along this node's own aim at the target — dependency depth 1.
type bufLayoutChainBead struct {
	OX float32 `buf:"f32"` // node-local offset x from this node's center
	OY float32 `buf:"f32"` // node-local offset y
	OZ float32 `buf:"f32"` // node-local offset z
	// Lit is the ANIMATION: 1 marks the bead a traversal has currently reached on this
	// chain, 0 every other bead. This replaces a bead MOVING along a wire — the chain is
	// fixed and the lighting is what advances (docs/bead-model/beads-are-the-edge.md).
	//
	// Go owns it: the source node drives its own outgoing wires and reads their in-flight
	// fractional progress t on its own goroutine, then lights index = t × count. The
	// renderer only colours what this column says.
	Lit uint8 `buf:"u8"` // 1 = a traversal has reached this bead
	// LitValue is the VALUE (0|1) of the traversal occupying this bead, meaningful only when
	// Lit==1. The lit bead is drawn in bead 0's or bead 1's own fill, so the renderer needs
	// the value, not just the fact of being lit — the whole animation is ONE fill-colour
	// change against the grey resting chain.
	LitValue int32 `buf:"i32"` // traversing bead's value (0|1); meaningful when Lit==1
	// There is no per-bead Radius column. Under bead CRUD (MODEL.md "Moving a node is
	// CRUD on the edge beads that touch it", nodes/Wiring/bead_crud.go) the single global
	// wire.BeadRadius/wire.BeadStepR lattice constants already make every chain's beads
	// touch their own neighbours on the chain exactly — a per-edge radius (added in
	// commit d50fab83, removed here) is unnecessary and was removed along with the
	// residue it existed to absorb.
}
