// mover_registry.go — the nodeMover/edgeMover directory owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): moverRegistry owns
// nodeMovers/edgeMovers/edgeOut and the Bind/Start/EdgeOut/sendMove/enqueueFor/
// centerOfNode logic. MoveDispatch's public Bind/Start/EdgeOut stay as thin delegators so
// the external API is unchanged; sendMove threads through md.ctx (owned elsewhere, NOT
// part of this extraction) as a parameter. The test-only message tap is per-mover
// (nodeMover.tap, node_mover.go) — enqueueFor no longer takes or threads a shared tap.

package Wiring

import (
	"context"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"sync"
)

// moverInboxDepth is the declared capacity of every per-mover moveMsg inbox: an
// edgeMover's extIn/srcIn/dstIn (edge_mover.go), a nodeMover's extIn
// (node_mover.go), and each directed neighborIn (node_move.go). Previously the same
// bare 8 repeated at six construction sites — the largest group of magic numbers in
// the network.
//
// WHY THIS NUMBER, honestly: it is a chosen depth, NOT derived. What IS derivable
// from the topology is the COUNT of these channels (one triple per edge, one extIn
// per node, two directed neighborIn per edge — 10/9/20 for the shipped graph), and
// the loader already fixes that by construction. The DEPTH is a queue for a burst of
// move messages during a drag, so it is bounded by gesture rate, not by anything in
// the spec files (docs/planning/visual-editor/session-log.md classifies it DYNAMIC for exactly this reason).
// 8 is "a few frames of drag messages"; it has held in practice and no measurement
// yet contradicts it.
//
// Deliberately NOT asserted, unlike maxPendingEvents. Filling one of these inboxes is
// not a bug: the sender keeps the message in its own pending queue and retries, which
// is the designed backpressure. The thing that actually grows unbounded when an inbox
// stays full is that retry queue (nodeMover.pending) — ITS bound (maxPendingSends,
// below) IS asserted, and enforced where nm.pending is checked in flushPending. Naming
// this constant is the "declared" half of the rule; the "asserted" half belongs to
// that queue, not to this capacity.
const moverInboxDepth = 8

// maxPendingSends is the declared, asserted upper bound on len(nm.pending) between
// flushPending calls (enforced below; see pending_bound_test.go). GENEROUS ceiling,
// not a tight derivation, same honesty as maxPendingEvents/maxInflightBeads in
// nodes/wire/paced_wire.go: nm.pending's own drain rate isn't a function of
// moverInboxDepth (flushPending only ever attempts ONE real send per blocked
// destination per cycle — every later item to that same destination is
// retained WITHOUT an attempt, to preserve FIFO — so a bigger peer inbox
// would not drain this queue any faster). Reusing moverInboxDepth twice-over
// (its square) is a way to get a value definitely too large to reach under
// ordinary drag traffic, not a claim that this is the tightest possible bound.
const maxPendingSends = moverInboxDepth * moverInboxDepth

// moverRegistry is the pure registry that owns every mover and wires their dedicated
// channels together — there is no shared dispatch map; nodeMovers/edgeMovers themselves
// are the directories a mover's resolveDest closure and the external-entry helpers below
// look up. It also retains the per-edge source Outs so out-of-package test/verifier
// callers can read an edge's loaded geometry (EdgeOut) without going through a central
// coordinator.
type moverRegistry struct {
	nodeMovers map[string]*nodeMover
	edgeMovers map[string]*edgeMover
	// edgeOut: edgeId → source *Out, for read-only access by tests/verifiers.
	edgeOut map[string]*wire.Out
	// centerMirror is the DISPATCH goroutine's OWN mirror of every node's last-known
	// world center, kept current by messages from each node's own goroutine.
	// Seeded once at construction (newMoveDispatch, single-threaded
	// setup, from each node's load-time geom) so the first framing read has every
	// center before any push arrives, then kept current by drainCenterMirror pulling
	// each nodeMover's own centerOut channel. Written and read ONLY from the dispatch/
	// gesture goroutine (centerOfNode is, after the quantize call sites moved to each
	// node's own partnerCenters map, called only from that goroutine) — no lock.
	centerMirror map[string]vec3
}

// bind wires the per-edge source Outs (keyed "source.sourceHandle" in outSink) and dest
// wires (slotReg, keyed "target.targetHandle") into each edgeMover. Call once after node
// construction.
func (mr *moverRegistry) bind(outSink map[string]*wire.Out, slotReg SlotRegistry) {
	for edgeID, em := range mr.edgeMovers {
		var o *wire.Out
		if oo, ok := outSink[em.srcID+"."+em.srcH]; ok {
			o = oo
			em.out = oo
			mr.edgeOut[edgeID] = oo
		}
		if pw, ok := slotReg[em.dstID+"."+em.dstH]; ok {
			em.dest = pw
			// The SOURCE node also takes this wire, paired with the outTargets entry for
			// the same edge: the source node's own goroutine drives it (nodeMover.run)
			// and reads its in-flight fractions to light its own chain
			// (docs/beads-are-the-edge.md step 3). The wire is no longer driven by a
			// goroutine of its own — that is what "the wire goroutine is removed" means
			// concretely, and it is why the node can read the fraction without touching
			// another goroutine's state.
			if srcNM, ok := mr.nodeMovers[em.srcID]; ok {
				srcNM.outWires = append(srcNM.outWires, pw)
				srcNM.outWireTargets = append(srcNM.outWireTargets, em.dstID)
				// Parallel to outWires: the source *Out this edge's step count is
				// PUBLISHED through (chainBeads calls PublishSteps on it — see
				// outWireOuts' doc comment). o may be nil if this edge's source handle
				// wasn't found in outSink; chainBeads then just skips publishing for
				// this edge (it still lays the chain out — the step count is computed
				// locally either way, see edgeStepCount).
				srcNM.outWireOuts = append(srcNM.outWireOuts, o)
				// Parallel to outWires/outWireOuts: this edge's OWN edgeMover.stepsIn
				// channel (edge_mover.go's doc comment) — the second delivery
				// chainBeads makes alongside PublishSteps, so the edgeMover's own
				// goroutine (which cannot read the Out directly — see stepsIn's doc
				// comment) can revise an in-flight bead's remaining travel against the
				// same freshly computed count.
				srcNM.outStepsIn = append(srcNM.outStepsIn, em.stepsIn)
			}
		}
	}
}

// start launches every mover's goroutine — ONE goroutine per node and ONE per edge, no
// dedicated sender/watcher goroutines (an earlier shared-outbox-plus-sender-goroutine
// design was removed: each mover's own run loop drains its own inbox AND retries its own
// pending sends, non-blockingly, every cycle).
//
// Returns a *sync.WaitGroup covering every launched goroutine, so a caller that wants a
// complete shutdown (main.go: "wait for everything, then close" — see
// the wait-for-everything-then-close change) can wg.Wait() on it after cancelling
// ctx. Both nm.run and em.run select on ctx.Done() at the top of their loop (their only
// blocking call is SleepCycle, which also selects on ctx), so cancel-to-return is one
// clock tick, worst case. Callers that don't care about shutdown completeness (most
// existing tests) can ignore the return value — start(ctx) alone still compiles and
// still launches every goroutine exactly as before.
func (mr *moverRegistry) start(ctx context.Context) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	for _, nm := range mr.nodeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nm.run(ctx)
		}()
	}
	for _, em := range mr.edgeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.run(ctx)
		}()
	}
	return wg
}

// edgeOutFor returns the source *Out bound to the given edge label, or nil if unknown.
// Read-only accessor for out-of-package verifiers (the headless cascade reads an
// edge's per-edge in-flight time from the loaded geometry).
func (mr *moverRegistry) edgeOutFor(edgeID string) *wire.Out {
	return mr.edgeOut[edgeID]
}

// drainCenterMirror drains every nodeMover's centerOut channel non-blockingly,
// updating mr.centerMirror with whatever's newest for each node. Called before every
// dispatch-side framing read (centerOfNode) so those reads always see the latest
// pushed center. This is EVENTUALLY CONSISTENT (a read may be one push behind a node
// that just moved on its own goroutine) — acceptable for camera/framing reads, which
// is the only remaining caller class (see moverRegistry.centerMirror's doc comment).
// Must only be called from the dispatch/gesture goroutine — it is the sole
// reader of every nodeMover.centerOut channel.
func (mr *moverRegistry) drainCenterMirror() {
	if mr.centerMirror == nil {
		mr.centerMirror = map[string]vec3{}
	}
	for id, nm := range mr.nodeMovers {
		select {
		case c := <-nm.centerOut:
			mr.centerMirror[id] = c
		default:
		}
	}
}

// centerOfNode returns the current world center for a node id by draining the center
// mirror (drainCenterMirror) and reading mr.centerMirror. Must only be called from the
// dispatch/gesture goroutine.
func (mr *moverRegistry) centerOfNode(id string) (vec3, bool) {
	mr.drainCenterMirror()
	c, ok := mr.centerMirror[id]
	return c, ok
}

// sendMove routes one moveMsg to a node's dedicated external-entry channel (extIn), if
// the id is a known node. This is the EXTERNAL-caller path — RootMove (drag) and
// gesture.go's dragStart send — not a mover-to-mover send (those go through a mover's
// own nm.pending/flushPending onto its OWN dedicated channel, never through this
// function), so it has no owning mover to fire a tap through — this bare path never
// fires the test-only tap (see nodeMover.tap's doc comment; only enqueueFor, a mover's
// own send, does). mr.nodeMovers is a read-only directory once construction finishes,
// safe to read from any goroutine. ctx is threaded through from MoveDispatch (not part
// of moverRegistry).
func (mr *moverRegistry) sendMove(ctx context.Context, id string, msg moveMsg) {
	nm, ok := mr.nodeMovers[id]
	if !ok {
		return
	}
	// Blocking send with a ctx-cancel escape hatch: this is the bare external-entry
	// send used by callers (RootMove, gesture.go) that have no owning mover goroutine
	// to thread a ctx from. Without the ctx.Done() arm, a send into a torn-down/full
	// extIn on shutdown parks this goroutine forever (the target's own run loop has
	// already returned on the same ctx cancel, so nothing will ever drain it). ctx
	// is nil only in tests that build a bare MoveDispatch without Start — a nil
	// Context's Done() channel would panic, so guard it and fall back to a plain
	// blocking send there (matches prior test behavior; no shutdown path exists in
	// that setting anyway).
	if ctx == nil {
		nm.extIn <- msg
		return
	}
	select {
	case nm.extIn <- msg:
	case <-ctx.Done():
	}
}

// enqueueFor returns nm's own non-blocking send function: it fires nm's own tap (at
// enqueue time, so tap-based tests' counts/ordering match today's behavior — a plain
// nil check + direct call, since nm.tap is owned and read only by nm's own goroutine,
// which is the only caller of the closure returned here), appends the message to nm's
// own pending retry queue, and attempts an immediate flush — never blocking the calling
// handler goroutine. Bound once per node at construction (nm.sendMove = md.enqueueFor(nm))
// so every send a nodeMover's own handle performs — including the ones
// fanEdgesAndPartners/requantizeLocalPolars make on that node's behalf — goes through
// nm's own retry queue, never a raw blocking channel write and never a second mover's
// queue (there is no shared outbox to route through anymore).
func (mr *moverRegistry) enqueueFor(nm *nodeMover) func(id string, msg moveMsg) {
	return func(id string, msg moveMsg) {
		if nm.tap != nil {
			nm.tap(id, msg)
		}
		nm.pending = append(nm.pending, pendingSend{destID: id, msg: msg})
		nm.flushPending()
		if len(nm.pending) > maxPendingSends {
			// Named causes, checked against flushPending's actual behaviour (not
			// guessed): an item whose destID doesn't resolve is DROPPED, not
			// retained (flushPending's `!ok` branch), so an unresolvable
			// destination can never grow this queue — it is deliberately not
			// named below. What CAN: (1) a peer whose own goroutine has
			// stopped draining its inbox entirely (wedged or dead) — every
			// later item to that same destination piles up behind it,
			// unattempted, to preserve FIFO; (2) this node enqueueing to a
			// live-but-slower peer faster, cycle over cycle, than that peer's
			// own goroutine drains its inbox — flushPending retries only ONE
			// send per blocked destination per cycle, so a persistent
			// per-cycle surplus accumulates even without a dead peer.
			panic(fmt.Sprintf(
				"nodeMover(%s): pending exceeded %d retry-queued sends; either a "+
					"destination's own goroutine has stopped draining its inbox "+
					"(wedged or dead), or this node is enqueueing to a peer faster "+
					"than that peer drains, cycle over cycle",
				nm.id, maxPendingSends))
		}
	}
}
