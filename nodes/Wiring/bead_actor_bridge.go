package Wiring

import wire "github.com/dtauraso/wirefold/nodes/wire"

// bead_actor_bridge.go — wires nodes/wire's bead-actor primitive (bead_actor.go,
// bead_wake_group.go) into chainBeads (chain_beads.go) as the SOLE source of a chain
// bead's POSITION, per PLAN.md "two clocks per bead, three channel sets".
//
// chainBeads still computes, per outgoing edge, the values a bead's position is a
// function OF (this node's live center, the unit aim toward the target, and the edge's
// bead-step count) — that is input derivation, not position computation, and it stays on
// this node's own goroutine same as before. What moved here is the LAST step: turning
// (center, aim) into N cartesian points. That step now happens inside each bead's own
// goroutine (Bead.applyTransform, bead_actor.go), reached over ONE broadcast hop
// (BeadWakeGroup.BroadcastGeometry) regardless of N, and chainBeads only ever READS the
// result back — never derives it, and never WAITS for it either (see broadcastAndRead).
//
// beadEdgeActors is THIS node's own live bead-actor group for ONE outgoing edge, keyed
// by target id in nodeMover.beadEdges. Owned and touched exclusively by this node's own
// goroutine (chainBeads runs on it, same as everything else in this file) — no lock, no
// atomic.
type beadEdgeActors struct {
	group *wire.BeadWakeGroup
	beads []*wire.Bead
	obs   []<-chan wire.BeadSnapshot
	stops []chan struct{}
	// last is this group's own CACHE of each bead's most recently observed snapshot —
	// the node's own copy, updated by a non-blocking drain in broadcastAndRead, never by
	// blocking on a bead. Seeded at bead construction (SeedGeometry) so index i is always
	// a valid position, even before that bead's own goroutine has served its first
	// broadcast.
	last []wire.BeadSnapshot
}

// ensureBeadEdgeActors returns this node's bead-actor group for target `to`, first
// growing or shrinking its goroutine set to exactly `count` beads. CRUD is add/remove AT
// THE CHAIN END only (MODEL.md "moving a node is CRUD on the edge beads that touch it"):
// growing starts new goroutines for the newly added far-end beads; shrinking tears down
// the far-end beads first (closes their stop channel — no leaked goroutines). offsetR for
// a newly created bead at index i is fixed forever at construction, exactly the same
// formula chainBeads used to apply directly: selfTorusR + wire.BeadTorusOuterR +
// i*wire.BeadStepR (docs/bead-lattice.md "Placement") — only the AIM a bead is asked to
// apply that fixed offset against ever changes, via later geometry broadcasts.
//
// xf is this edge's CURRENT transform, used only to seed a newly created bead's initial
// position (Bead.SeedGeometry, called before Start()) so it has something valid to report
// before it has ever serviced a live broadcast — it is not otherwise consulted here.
func (m *nodeMover) ensureBeadEdgeActors(to string, count int, selfTorusR float64, xf wire.BeadGeometryIn) *beadEdgeActors {
	if m.beadEdges == nil {
		m.beadEdges = map[string]*beadEdgeActors{}
	}
	ea, ok := m.beadEdges[to]
	if !ok {
		ea = &beadEdgeActors{group: wire.NewBeadWakeGroup()}
		m.beadEdges[to] = ea
	}
	for len(ea.beads) < count {
		i := len(ea.beads)
		offsetR := selfTorusR + wire.BeadTorusOuterR + float64(i)*wire.BeadStepR
		geom, wake, settle := ea.group.Current()
		tickCh := wire.NewTickChan()
		stop := make(chan struct{})
		b := wire.NewBead(offsetR, geom, wake, settle, tickCh, stop)
		obs := b.WithObserve()
		snap := b.SeedGeometry(xf)
		b.Start()
		ea.beads = append(ea.beads, b)
		ea.obs = append(ea.obs, obs)
		ea.stops = append(ea.stops, stop)
		ea.last = append(ea.last, snap)
	}
	for len(ea.beads) > count {
		last := len(ea.beads) - 1
		close(ea.stops[last])
		ea.beads = ea.beads[:last]
		ea.obs = ea.obs[:last]
		ea.stops = ea.stops[:last]
		ea.last = ea.last[:last]
	}
	return ea
}

// startAllBeadDrags wakes every bead-actor group THIS node owns (one group per outgoing
// edge) — sets every one of that edge's beads onto machine time — with one close per
// group, never a loop of per-bead sends (BeadWakeGroup.StartDrag). Called exactly once
// per drag gesture, from THIS node's own moveMsgKindDragStart handler: this node is the
// one being dragged, and PLAN.md "either endpoint can wake" means the SOURCE-owned groups
// for its own outgoing edges wake here; the mirror case (this node being the TARGET of a
// neighbor's outgoing edge) is that neighbor's own moveMsgKindDragStart handler waking
// its own groups — no both-at-once case exists (only one node is ever dragged at a time).
func (m *nodeMover) startAllBeadDrags() {
	for _, ea := range m.beadEdges {
		ea.group.StartDrag()
	}
}

// endAllBeadDrags is startAllBeadDrags' mirror, called from moveMsgKindDragEnd — settles
// every bead this node's groups woke, on EVERY path a drag can end by (PLAN.md "'done
// dragging' is not optional").
func (m *nodeMover) endAllBeadDrags() {
	for _, ea := range m.beadEdges {
		ea.group.EndDrag()
	}
}

// broadcastAndRead delivers this edge's live transform to every bead in ea in ONE hop (a
// single BroadcastGeometry call, i.e. one close — not a loop of N sends) and then reads
// back each bead's own CURRENTLY-HELD snapshot, non-blocking.
//
// This does NOT wait for that broadcast to be applied. Every read here is a `select` with
// `default:` against that bead's own observe channel: if the bead has already pushed a
// fresher snapshot, take it and cache it in ea.last; if not — because its goroutine hasn't
// been scheduled yet, or is mid-tick, or is behind on a run of broadcasts issued faster
// than it drains (a fast drag can easily issue several BroadcastGeometry calls between two
// scheduler slices of one bead) — keep the PREVIOUSLY cached value in ea.last and move on.
// There is no generation match, no loop, no blocking receive anywhere in this function: a
// bead that is arbitrarily far behind still returns instantly, with a slightly stale
// position, rather than making this node's own goroutine wait on it.
//
// A returned position may therefore be up to one broadcast (or a few, under a fast drag)
// stale. THAT IS CORRECT under this model, not a bug being tolerated: the bead owns its
// position and this node's goroutine reads what it currently is, exactly like reading any
// other bead-owned state (Lit/LitVal already work this way) — it does not stop and
// interrogate the bead for a synchronous answer. The staleness is bounded by one tick of
// scheduler latency (microseconds) and self-heals on the very next call, which reads
// whatever the bead has caught up to by then; nothing accumulates, because ea.last is
// always overwritten with the newest available value, never advanced by waiting.
func (ea *beadEdgeActors) broadcastAndRead(xf wire.BeadGeometryIn) []wire.Vec3 {
	ea.group.BroadcastGeometry(xf)
	positions := make([]wire.Vec3, len(ea.beads))
	for i, obs := range ea.obs {
		select {
		case snap := <-obs:
			ea.last[i] = snap
		default:
		}
		positions[i] = ea.last[i].Position
	}
	return positions
}
