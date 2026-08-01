// bead_actor.go — the render (CHAIN) bead as its own goroutine, per PLAN.md
// "two clocks per bead, three channel sets" (docs/beads-are-the-edge.md names this same
// entity: the node-owned placeholder chain that IS the visual of an edge).
//
// STATUS: this is a PRIMITIVE, validated by its own tests (bead_actor_test.go) but with NO
// PRODUCTION CALL SITE yet. `nodes/Wiring/chain_beads.go`'s chainBeads() — the function
// that actually feeds the live content buffer — does not construct a Bead or a
// BeadWakeGroup; it still computes chain-bead positions the way it always has, per frame,
// as plain slice entries. Nothing in the running editor uses this file yet. See MODEL.md's
// "Chain (render/placeholder) bead" bullet, which states the same status, and
// memory/project_wire_is_straight_line_not_chain.md for why the integration was deferred
// rather than rushed. Do not read the rest of this comment as a description of live
// behaviour — it describes what this primitive DOES, in isolation, today.
//
// This is a
// DIFFERENT bead from PacedWire's in-flight VALUE beads (paced_wire.go) — that bead is a
// value in transit between two node goroutines and stays a passive delay-queue entry, per
// MODEL.md's Wire section, unchanged by this file. This file is about the chain bead: the
// placeholder that renders a traversal and that a node drag must reposition at machine
// speed with no visible per-bead lag (memory/project_wire_is_straight_line_not_chain.md —
// the O(N^2) defect that reverted the original chain model was human-clock gating plus
// momentum-free midpoint averaging, not "a chain of beads" itself).
//
// A Bead is now a GOROUTINE, not a slice entry. It is driven by two clocks over disjoint
// state, reached over three DISTINCT channel sets:
//
//   - GEOMETRY (machine time): a BroadcastChain carrying BeadGeometryIn — the owning node's
//     live transform — broadcast to every bead in ONE hop (one close, not N sends); a Bead
//     computes its own POSITION from it directly (body force, dependency depth 1 — no
//     neighbour reads, no relaxation).
//   - ANIMATION/TICK (human time, MsPerTick): a pulse from the process's one
//     TickBroadcaster (see clock.go) that advances the bead's own lit/carried-value state.
//   - MODE: two more BroadcastChains — wake (the bead is DRAGGING, on machine time) and
//     settle (DONE DRAGGING, back to human time). The mode is ONE LOCAL BOOL
//     (Bead.dragging), written only by the Bead's own goroutine, set once per drag by the
//     wake broadcast and cleared once per drag by the settle broadcast — never toggled per
//     pointer/geometry event.
//
// Position (beadGeometryState) and animation (beadAnimationState) are separate TYPES, each
// mutated from exactly one select case, so the "one writer per field" split is visible to
// the compiler and not merely a comment's promise.
//
// The goroutine's own loop (run, below) is ONE select over all three channel sets with NO
// `default:` case — see the file-level comment on RealClock.SleepCycle (clock.go) for why
// that distinction is load-bearing: `default:` makes the loop non-blocking, so a caller
// looping around it spins a core; omitting it lets the runtime park the goroutine on every
// channel's wait queue at zero CPU until one of them has something. Guarded in source by
// tools/check-no-select-default.sh (scoped to this file's run loop, fenced by
// the run-loop fence markers below — NOT a repo-wide "no default in any select" rule, since
// sendStepsNonBlocking and friends elsewhere in this package correctly rely on `default:`
// for a non-blocking latest-wins send).
package wire

// BeadGeometryIn is the payload a node broadcasts to every bead on its edges each time its
// own transform changes: the one thing a bead's position is computed FROM. It carries
// enough for a bead to place itself directly (body force) — this node's live world
// center and its live unit aim toward the edge's far end — with no neighbour-of-neighbour
// read and no relaxation pass.
type BeadGeometryIn struct {
	Center Vec3
	Aim    Vec3 // unit direction from Center toward the edge's far node
}

// BeadAnimIn is the payload a node broadcasts to every bead on its edges each time the
// edge's lit set changes: which bead indices are currently lit and, for each, the VALUE
// the traversal reaching it carries. A single value broadcast in ONE hop (never a per-bead
// loop) — each bead looks up its OWN index in the map to decide its own lit/value, the same
// "one hop, bead computes its own state" shape BeadGeometryIn/applyTransform already use
// for position. This is what routes "which bead is lit" OFF chainBeads and onto the bead
// itself (PLAN.md: a woken bead does exactly two things, move and send its own colour) —
// chainBeads keeps computing litBeadIndex per in-flight traversal (that is edge-owned
// progress data, not a bead's own state) but hands the result to the beads instead of
// writing it into the output row itself.
type BeadAnimIn struct {
	LitVals map[int]int32
}

// beadGeometryState is the bead's POSITION — written from exactly one place: Bead.run's
// geometry-broadcast case, via applyTransform. Nothing else touches it (not the mode case,
// not the tick case), which is the "disjoint writers" test PLAN.md asks for made a
// separate Go type instead of a shared struct field.
type beadGeometryState struct {
	position Vec3
}

// applyTransform computes this bead's position directly from the node's broadcast
// transform and this bead's own fixed node-local offset (index*wire.BeadStepR along the
// node's aim, docs/bead-lattice.md "Placement") — ONE hop, no dependency on any other
// bead's position. offsetR is this bead's fixed distance from the node center along Aim.
func (g *beadGeometryState) applyTransform(xf BeadGeometryIn, offsetR float64) {
	g.position = xf.Center.Add(xf.Aim.Scale(offsetR))
}

// beadAnimationState is the bead's ANIMATION (lit / carried value) — written from exactly
// one place: Bead.run's tick-channel case, via tick. Disjoint from beadGeometryState: a
// human-clock pulse never touches position, and a geometry broadcast never touches this.
type beadAnimationState struct {
	lit    bool
	litVal int32
}

// tick applies one animation update — either a human-clock pulse (test/primitive use, see
// bead_actor_test.go) or a real animation broadcast from the owning node (applyAnim below,
// production use). The actual traversal-lit rule (which index is lit,
// docs/beads-are-the-edge.md) is owned by the node's chainBeads/litBeadIndex computation
// (nodes/Wiring/chain_beads.go), which decides the VALUES fed in here — but no longer
// decides this bead's own Lit/LitVal directly; this bead does, from what it is told. The
// point under test is that this write happens ONLY from the animation channel set (tick or
// animIn), never from a geometry or mode event.
func (a *beadAnimationState) tick(lit bool, val int32) {
	a.lit = lit
	a.litVal = val
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
	Anim  BeadAnimIn     // meaningful only for animation-generation chains; zero otherwise
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

// AdvanceWithAnim is AdvanceWithValue's sibling for the animation broadcast: stamps a
// BeadAnimIn payload (this edge's live lit set) into the CURRENT link before firing.
func (c *BroadcastChain) AdvanceWithAnim(v BeadAnimIn) *BroadcastChain {
	c.Anim = v
	return c.Advance()
}

// Bead is a placeholder chain bead, now a goroutine. offsetR is fixed at construction
// (index*wire.BeadStepR along the owning node's aim) and never changes — only the AIM the
// offset is applied against changes, via geometry broadcasts, which is what makes a node
// move reposition every bead in one hop instead of by neighbour-following.
type Bead struct {
	offsetR float64
	index   int // this bead's own fixed position in its edge's chain (docs/bead-lattice.md); set once at construction, used only to read its own key out of a broadcast BeadAnimIn.LitVals

	geom   *BroadcastChain // current geometry generation this bead is waiting on
	tickCh <-chan struct{} // this bead's own subscription to the human clock
	wake   *BroadcastChain // current "dragging" generation this bead is waiting on
	settle *BroadcastChain // current "done dragging" generation this bead is waiting on
	animIn *BroadcastChain // current animation generation this bead is waiting on (BeadAnimIn — the edge's live lit set); nil-safe: a nil animIn is never selected on (see run's guard)
	stop   <-chan struct{} // closed to tear the goroutine down (tests/edge teardown)

	// local state — owned and mutated ONLY by this Bead's own goroutine (run). Nothing
	// else reads or writes these; there is no lock and no atomic because there is exactly
	// one goroutine that ever touches them. A second goroutine (a test, or a future render
	// consumer) may NOT read these fields directly — that would be exactly the
	// cross-goroutine shared-read this package's ownership model forbids. Observe below is
	// the one sanctioned way out: a value PUSHED by this goroutine, never pulled by another.
	geomState beadGeometryState
	anim      beadAnimationState
	dragging  bool // the ONE mode flag: set by <-wake.Fire, cleared by <-settle.Fire

	// geomGen counts how many geometry broadcasts THIS bead has applied so far.
	// Written only by the geom case in run(), alongside geomState — a production
	// reader (the owning node's own goroutine, chain_beads.go) that just issued its
	// Nth BroadcastGeometry call can block-read this bead's observe channel until it
	// sees GeomGen==N, which is race-free against interleaved tick/mode pushes on the
	// same buffered-1 observe channel (those pushes carry a smaller/equal GeomGen,
	// never a larger one, since only the geom case increments it).
	geomGen int

	// observe is an OPTIONAL, buffered-1, latest-wins outbox this goroutine pushes its own
	// snapshot onto after every state change (the same non-blocking drain-then-send shape
	// SendLatestNonBlocking/SendSpeedNonBlocking already use in clock.go for "one owner
	// pushes, one reader always sees the latest"). Nil in normal production use (nothing
	// outside this goroutine needs to read a bead's state yet); tests set it to observe
	// state changes without reading b's fields from another goroutine.
	observe chan BeadSnapshot
}

// BeadSnapshot is a bead's state as of the last change, PUSHED by the bead's own goroutine
// — the only way another goroutine may observe a Bead's local state (see observe above).
type BeadSnapshot struct {
	Position Vec3
	Dragging bool
	Lit      bool
	LitVal   int32
	GeomGen  int
}

// NewBead constructs a bead bound to its owning node's channel sets. geom/wake/settle/animIn
// are the CURRENT generation of each of the node's four BroadcastChains (BeadWakeGroup owns
// all four); tickCh is this bead's own dedicated human-clock subscription (the same
// per-goroutine-clock convention RealClock.tickCh already uses in this package). index is
// this bead's own fixed position in its edge's chain, used only to read its own key out of
// a broadcast BeadAnimIn.LitVals.
func NewBead(offsetR float64, index int, geom, wake, settle, animIn *BroadcastChain, tickCh <-chan struct{}, stop <-chan struct{}) *Bead {
	return &Bead{
		offsetR: offsetR,
		index:   index,
		geom:    geom,
		wake:    wake,
		settle:  settle,
		animIn:  animIn,
		tickCh:  tickCh,
		stop:    stop,
	}
}

// SeedGeometry applies xf directly to this bead's own geometry state and returns the
// resulting snapshot — NOT a channel operation, and safe only because it is called
// BEFORE Start(): there is no goroutine running yet to race it, so a plain field write
// is exactly as safe as any other single-threaded setup step. It exists so a freshly
// constructed bead's position is valid from construction (PLAN.md: "a first frame never
// has nothing to read") — without it, a bead started with a still-unfired geometry chain
// would read as the zero Vec3 until its first broadcast is serviced.
func (b *Bead) SeedGeometry(xf BeadGeometryIn) BeadSnapshot {
	b.geomState.applyTransform(xf, b.offsetR)
	b.geomGen++
	return BeadSnapshot{Position: b.geomState.position, Dragging: b.dragging, Lit: b.anim.lit, LitVal: b.anim.litVal, GeomGen: b.geomGen}
}

// (There is deliberately no Position()/Dragging() getter on Bead: any read from a second
// goroutine of state owned by this one must go through observe/BeadSnapshot — a plain
// getter would silently reintroduce the cross-goroutine shared-read this type exists to
// rule out. `go test -race` catches a violation of this immediately; see bead_actor_test.go
// for the one time it did, before observe existed.)

// WithObserve arms a buffered-1 observation channel and returns it, for a caller (test-only
// today) that wants to watch this bead's own pushed state changes instead of ever reading
// b's fields from another goroutine. Call before Start().
func (b *Bead) WithObserve() <-chan BeadSnapshot {
	b.observe = make(chan BeadSnapshot, 1)
	return b.observe
}

// pushObserve is the send half of observe — non-blocking, latest-wins, called by this
// bead's own goroutine only, after every state change. A nil observe (production default)
// makes this a no-op.
func (b *Bead) pushObserve() {
	if b.observe == nil {
		return
	}
	snap := BeadSnapshot{Position: b.geomState.position, Dragging: b.dragging, Lit: b.anim.lit, LitVal: b.anim.litVal, GeomGen: b.geomGen}
	select {
	case b.observe <- snap:
		return
	default:
	}
	select {
	case <-b.observe:
	default:
	}
	select {
	case b.observe <- snap:
	default:
	}
}

// run is THIS bead's entire lifetime: one goroutine, one select, no default. Idle (nothing
// dragging, nothing ticking, no geometry event) means every case is unready, so the
// runtime parks this goroutine at zero CPU — see the file header.
//
// BEAD-SELECT-START
func (b *Bead) run() {
	for {
		select {
		case <-b.geom.Fire:
			// Geometry channel: the ONLY writer of b.geomState. One broadcast hop from
			// the node, applied directly — no neighbour read, no relaxation.
			b.geomState.applyTransform(b.geom.Value, b.offsetR)
			b.geomGen++
			b.geom = b.geom.Next
			b.pushObserve()
		case <-b.wake.Fire:
			// Mode message 1/2: dragging. Set once per drag by the node's single close
			// (BroadcastChain.Advance), not per pointer event.
			b.dragging = true
			b.wake = b.wake.Next
			b.pushObserve()
		case <-b.settle.Fire:
			// Mode message 2/2: done dragging. Clears the flag; back to resting on the
			// human clock. A drag abandoned without a clean "drag end" still reaches this
			// case because the node always advances settle when a drag concludes by ANY
			// path (see BeadWakeGroup.EndDrag) — there is no separate "abandoned" branch.
			b.dragging = false
			b.settle = b.settle.Next
			b.pushObserve()
		case <-b.tickCh:
			// Animation/tick channel: a writer of b.anim, alongside animIn below (both are
			// on the animation channel SET, disjoint from geom/wake/settle — PLAN.md's
			// table has one clock per set, not one channel per set). Human-clock pulse; a
			// resting (non-dragging) bead spends nearly all its life parked here.
			b.anim.tick(!b.anim.lit, b.anim.litVal)
			b.pushObserve()
		case <-b.animIn.Fire:
			// Animation channel: this bead's own colour, told to it by the owning node
			// (chainBeads, nodes/Wiring/chain_beads.go) rather than decided there — the
			// node still computes the traversal's progress (litBeadIndex), but which BEAD
			// displays it is this bead's own read of its own index against the broadcast
			// lit set (one hop, no per-bead loop — same shape as geometry's applyTransform).
			val, litHere := b.animIn.Anim.LitVals[b.index]
			b.anim.tick(litHere, val)
			b.animIn = b.animIn.Next
			b.pushObserve()
		case <-b.stop:
			return
		}
	}
}

// BEAD-SELECT-END

// Start launches the bead's goroutine. Call once per Bead.
func (b *Bead) Start() { go b.run() }
