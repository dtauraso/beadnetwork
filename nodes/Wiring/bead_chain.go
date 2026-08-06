// bead_chain.go — the PRODUCTION call site for nodes/wire's bead-actor primitive
// (bead_actor.go, bead_wake_group.go): the node-owned mapping from ONE outgoing edge's
// live bead count/aim (chainBeads' own count and direction, chain_beads.go) onto a real
// chain of *wire.Bead goroutines, and the readback of their pushed positions into the
// content buffer.
//
// This is deliberately NOT inside chainBeads' own placement loop by default: chainBeads
// is exercised directly, synchronously, by bare `&nodeMover{...}` test literals
// (chain_beads_test.go) that construct no goroutine and expect a pure, deterministic
// answer on the calling goroutine with no live TickBroadcaster started as a side effect.
// The switch is m.beadTickFn (nil in every such test, set to wire.NewTickChan only by
// production's newNodeMover) — chainBeads reads it once per outgoing target and only
// calls into this file when it is non-nil.
package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// edgeBeadChain is ONE outgoing edge's live chain of bead-actor goroutines, owned by this
// node's own goroutine (nodeMover.run/handle/chainBeads — never a second goroutine reads
// or writes any field here). beads/stops/snaps/last/valid are parallel slices, one entry
// per chain-bead index, always kept the same length as each other and as the edge's
// current bead count (chain_beads.go's edgeStepCount) by reconcileBeadChain.
type edgeBeadChain struct {
	group *wire.BeadWakeGroup
	beads []*wire.Bead
	// stops holds THIS chain's own per-bead teardown channel (one per bead, individually
	// closable) — reconcileBeadChain closes stops[i] to tear down bead i's goroutine when
	// the chain shrinks, which is what keeps bead-goroutine lifetime following chain
	// length (a removed bead's goroutine must exit, never leak).
	stops []chan struct{}
	// snaps is bead i's own observe channel (see Bead.WithObserve) — this node's own
	// goroutine is the ONLY reader (Bead.observe's doc comment: a plain getter would
	// reintroduce the cross-goroutine shared-read the type exists to rule out).
	snaps []<-chan wire.BeadSnapshot
	// last is the most recently drained snapshot per bead; valid[i] is false until bead
	// i's goroutine has pushed at least once (a freshly started bead, or one that has not
	// yet serviced its first geometry broadcast) — chainBeads falls back to its own
	// inline computation for any index where valid[i] is false, so a frame is never wrong
	// while an actor is still catching up, only briefly one hop behind.
	last  []wire.BeadSnapshot
	valid []bool
	// haveAim/lastAim: the aim direction this chain's beads were last broadcast, so
	// reconcileBeadChain only issues a fresh BroadcastGeometry when the aim (or the bead
	// count) actually changed — an idle chain (nothing dragging on either end) never
	// advances its geometry generation.
	haveAim bool
	lastAim wire.Vec3
	// lattice is offsetAt(0) for the placement these beads were BUILT with. A bead's own
	// offset is baked in at construction and never changes for its lifetime
	// (bead_actor.go's Bead.offsetR), so a chain cannot be RE-SPACED by growing it: the
	// beads that already exist would keep the old spacing while appended ones used the new
	// one — which is exactly what a naive "grow on overlay flip" produces, an unchanged
	// chain with extra beads trailing off its far end. The tween overlay changes that
	// spacing, so when this value moves the chain is torn down and rebuilt rather than
	// grown. haveLattice distinguishes "never built" from "built at offset 0".
	haveLattice bool
	lattice     float64
}

// reconcileBeadChain grows or shrinks this node's own bead-actor chain for outgoing edge
// `to` to `count` beads, then broadcasts fresh geometry (one hop, one close) when the
// count or the aim changed. offsetAt(i) is chainBeads' own fixed per-index offset
// (selfTorusR + wire.BeadTorusOuterR + i*wire.BeadStepR) — unchanged for the life of bead
// i, exactly as bead_actor.go's Bead.offsetR documents. Called only from chainBeads, only
// on this node's own goroutine.
func (m *nodeGeometry) reconcileBeadChain(to string, count int, offsetAt func(i int) float64, aim wire.Vec3) *edgeBeadChain {
	if m.beadChains == nil {
		m.beadChains = map[string]*edgeBeadChain{}
	}
	c := m.beadChains[to]
	if c == nil {
		c = &edgeBeadChain{group: wire.NewBeadWakeGroup()}
		m.beadChains[to] = c
	}
	// Re-space: the per-index offsets changed (the tween overlay flipping the lattice), so
	// every existing bead holds an offset it cannot be told to update. Tear the chain down
	// to empty here and let the grow loop below rebuild it at the new spacing — each removed
	// bead's own stop channel closes, so no goroutine leaks, exactly as the shrink path
	// does. Rare and cheap: this runs only when the lattice constant itself moves, never on
	// a move, a drag, or an ordinary count change.
	if lat := offsetAt(0); !c.haveLattice || lat != c.lattice {
		for i := range c.stops {
			close(c.stops[i])
		}
		c.beads, c.stops, c.snaps, c.last, c.valid = nil, nil, nil, nil, nil
		c.haveLattice, c.lattice = true, lat
		// Force the geometry broadcast below: the rebuilt beads have never been aimed.
		c.haveAim = false
	}
	// Grow: add beads at the chain END (bead CRUD's own convention, bead_crud.go) —
	// index len(c.beads) is always the next one appended, never inserted mid-chain.
	for len(c.beads) < count {
		i := len(c.beads)
		geom, wake, settle := c.group.Current()
		stop := make(chan struct{})
		b := wire.NewBead(offsetAt(i), geom, wake, settle, m.beadTickFn(), stop)
		snap := b.WithObserve()
		b.Start()
		c.beads = append(c.beads, b)
		c.stops = append(c.stops, stop)
		c.snaps = append(c.snaps, snap)
		c.last = append(c.last, wire.BeadSnapshot{})
		c.valid = append(c.valid, false)
	}
	// Shrink: remove beads from the chain END, closing each removed bead's OWN stop
	// channel so its goroutine exits (Bead.run's `case <-b.stop: return`) — this is the
	// half of bead-goroutine lifetime that keeps a removed bead from leaking.
	for len(c.beads) > count {
		last := len(c.beads) - 1
		close(c.stops[last])
		c.beads = c.beads[:last]
		c.stops = c.stops[:last]
		c.snaps = c.snaps[:last]
		c.last = c.last[:last]
		c.valid = c.valid[:last]
	}
	if !c.haveAim || aim != c.lastAim {
		// One broadcast hop, not N sends: every existing (and freshly-grown) bead in
		// this chain resolves its own position directly from this single value —
		// dependency depth 1, no neighbour read (PLAN.md "Position updates: body
		// force"). Aim is broadcast in NODE-LOCAL terms (unit direction only, offsetAt
		// already carries the distance) so a bead's resolved position IS the node-local
		// offset chainBeads has always returned — no separate absolute/local
		// conversion is needed here or downstream.
		c.group.BroadcastGeometry(wire.BeadGeometryIn{Aim: aim})
		c.lastAim = aim
		c.haveAim = true
	}
	// Drain every bead's own latest pushed snapshot — non-blocking, latest-wins, this
	// node's own goroutine is the sole reader. A bead that has not pushed since the last
	// drain (nothing changed, or it has not yet serviced the broadcast above) simply
	// leaves last[i]/valid[i] at whatever they already were.
	for i, ch := range c.snaps {
		select {
		case s := <-ch:
			c.last[i] = s
			c.valid[i] = true
		default:
		}
	}
	return c
}

// startBeadDrag arms every one of this node's OWN outgoing-edge bead chains onto machine
// time — one close per edge (StartDrag), never a per-bead send-loop. Called from handle's
// moveMsgKindDragStart case, the same edge armDragAnchor already runs on. A chain that
// does not exist yet (this node has never emitted a frame with beads on that edge) is
// simply absent from m.beadChains and is skipped — its beads will be constructed already
// resting (dragging=false) by the first reconcileBeadChain call this drag makes, which is
// harmless: the mode flag does not gate geometry delivery (see bead_actor.go's Bead.run —
// the geometry case applies unconditionally, dragging or not), only observability.
func (m *nodeGeometry) startBeadDrag() {
	for _, c := range m.beadChains {
		c.group.StartDrag()
	}
}

// endBeadDrag is startBeadDrag's mirror: settles every one of this node's own outgoing
// bead chains — clears the dragging flag with one close per edge. Called from handle's
// moveMsgKindDragEnd case, on EVERY path a drag ends by (see that message kind's own doc
// comment), so no bead this node woke is ever left on machine time.
func (m *nodeGeometry) endBeadDrag() {
	for _, c := range m.beadChains {
		c.group.EndDrag()
	}
}
