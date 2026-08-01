// bead_wake_group.go — the endpoint-node side of the bead broadcast model (PLAN.md "Each
// node on either end of an edge owns a wakeup channel to all of that edge's beads").
//
// STATUS: NO PRODUCTION CALL SITE yet — see bead_actor.go's file header. No node in the
// running editor constructs a BeadWakeGroup today.
//
// A
// BeadWakeGroup is owned by ONE node goroutine, for ONE outgoing edge's chain of Beads; it
// holds the CURRENT generation of each of the three BroadcastChains (geometry, wake,
// settle) and is the only thing that ever calls Advance on them — a single-writer
// discipline, same as every other owned-state type in this package.
package wire

// BeadWakeGroup is one edge's broadcast surface, owned by the edge's SOURCE node's own
// goroutine (an edge's target node reaches the same beads through its OWN BeadWakeGroup —
// PLAN.md "either endpoint can wake" — see StartDrag's doc comment; there is exactly one
// BeadWakeGroup per (node, edge, which-endpoint) triple, not a shared one between the two
// endpoints, so each endpoint advances its own generations independently and neither
// blocks on the other).
type BeadWakeGroup struct {
	geom   *BroadcastChain
	wake   *BroadcastChain
	settle *BroadcastChain
	anim   *BroadcastChain
}

// NewBeadWakeGroup returns a group with all four chains freshly armed (unfired), ready for
// beads to be constructed against via Current() and for the owning node to drive via
// BroadcastGeometry/StartDrag/EndDrag/BroadcastAnim.
func NewBeadWakeGroup() *BeadWakeGroup {
	return &BeadWakeGroup{
		geom:   NewBroadcastChain(),
		wake:   NewBroadcastChain(),
		settle: NewBroadcastChain(),
		anim:   NewBroadcastChain(),
	}
}

// Current returns the group's live generation of each chain, for constructing a Bead that
// starts out waiting on exactly what every other bead on this edge is waiting on.
func (g *BeadWakeGroup) Current() (geom, wake, settle, anim *BroadcastChain) {
	return g.geom, g.wake, g.settle, g.anim
}

// BroadcastGeometry delivers this node's fresh transform to every bead in this group in ONE
// hop: a single Advance call, not a loop over N beads. Called once per pointer/geometry
// event while dragging (or whenever this node's transform otherwise changes) — the N
// per-bead position computations happen inside each bead's own goroutine, concurrently,
// not here.
func (g *BeadWakeGroup) BroadcastGeometry(xf BeadGeometryIn) {
	g.geom = g.geom.AdvanceWithValue(xf)
}

// BroadcastAnim delivers this edge's fresh lit set to every bead in this group in ONE hop —
// BroadcastGeometry's sibling for colour instead of position (PLAN.md: a woken bead does
// exactly two things, move and send its own colour). Each bead reads its own index out of
// v.LitVals to decide its own Lit/LitVal; this call never loops over beads.
func (g *BeadWakeGroup) BroadcastAnim(v BeadAnimIn) {
	g.anim = g.anim.AdvanceWithAnim(v)
}

// StartDrag wakes every bead in this group — sets each one's mode flag to dragging — with a
// SINGLE close, regardless of how many beads are waiting on g.wake.Fire. Called exactly
// once per drag gesture (PLAN.md "the flag is set once per drag, not per pointer event"),
// from the gesture FSM's drag-start edge (only one node is ever the drag target at a time —
// PLAN.md "do not build reference counting... the input model cannot produce that case").
// Both the SOURCE and the TARGET node of an edge own their own BeadWakeGroup for that edge,
// so dragging either endpoint reaches the same beads through its own group — one test each,
// never a both-at-once test.
func (g *BeadWakeGroup) StartDrag() {
	g.wake = g.wake.Advance()
}

// EndDrag settles every bead in this group — clears each one's mode flag — with a single
// close. Called exactly once per drag gesture, on EVERY path a drag can end by, including
// one abandoned without a clean "drag end" event (PLAN.md "'done dragging' is not
// optional"): the gesture FSM's reset path calls this unconditionally whenever g.wake has
// been advanced without a matching EndDrag, so a bead can never be left on machine time.
func (g *BeadWakeGroup) EndDrag() {
	g.settle = g.settle.Advance()
}
