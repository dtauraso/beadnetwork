// mover_registry.go — the nodeMover/edgeMover directory owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): moverRegistry owns
// nodeMovers/edgeMovers/edgeOut and the bind/start/sendMove/enqueueFor/
// centerOfNode logic. In-package callers address md.mr.X directly (bind/edgeOutFor/
// centerOfNode/enqueueFor/finalizeActors have no MoveDispatch-level delegator); only
// Start stays on MoveDispatch, since it also sets md.ctx. sendMove threads through
// md.ctx (owned elsewhere, NOT part of this extraction) as a parameter. The test-only
// message tap is per-mover (nodeMover.tap, node_mover.go) — enqueueFor no longer takes
// or threads a shared tap.

package Wiring

import (
	"context"
	"fmt"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// moverInboxDepth is the declared capacity of every per-mover movemsg.Msg inbox: an
// edgeMover's extIn/srcIn/dstIn (edge_mover.go), a nodeMover's extIn
// (node_mover.go), and each directed neighborIn (move_dispatch_construct.go). Previously the same
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
// below) IS asserted, and enforced where nm.msg.pending is checked in flushPending. Naming
// this constant is the "declared" half of the rule; the "asserted" half belongs to
// that queue, not to this capacity.
const moverInboxDepth = 8

// maxPendingSends is the declared, asserted upper bound on len(nm.msg.pending) between
// flushPending calls (enforced below; see pending_bound_test.go). GENEROUS ceiling,
// not a tight derivation, same honesty as maxPendingEvents/maxInflightBeads in
// nodes/wire/paced_wire.go: nm.msg.pending's own drain rate isn't a function of
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
// look up. It also retains the per-edge source Outs so in-package callers can read an
// edge's loaded geometry (edgeOutFor) without going through a central coordinator.
type moverRegistry struct {
	// nodeGeoms is the UNIVERSAL per-node directory — every node's own *nodeGeometry,
	// ring and pair alike. This is what routing (resolveDest, sendMove, centerOfNode,
	// NodeKind, drag/commit) looks up: a message addressed to a node must arrive and be
	// handled regardless of which goroutine drives that node's geometry.
	nodeGeoms map[string]*nodeGeometry
	// nodeMovers is the RING-ONLY actor directory: one entry per node whose OWN kind did
	// NOT claim BuildArgs.ClaimSelfDrive, populated by finalizeActors AFTER buildNodes has
	// run (so every ClaimSelfDrive call has already happened). Used ONLY by start() to
	// launch a dedicated goroutine per ring node — a PAIR node (PairNode) has NO entry
	// here at all, by construction, not by a flag that says "launch nothing for me".
	nodeMovers map[string]*nodeMover
	// selfDriveClaimed holds, for each node id whose OWN kind claimed
	// BuildArgs.ClaimSelfDrive at build time (PairNode — the pair scene), true. Written
	// ONCE per entry by ClaimSelfDrive (build_args.go), on the single-threaded build path,
	// before finalizeActors runs and before any goroutine exists. finalizeActors reads it
	// to decide which nodes get a nodeMover actor at all — an id present here gets NONE.
	selfDriveClaimed map[string]bool
	edgeMovers       map[string]*edgeMover
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
func (mr *moverRegistry) bind(outSink map[string]*wire.Out, slotReg inputcodec.SlotRegistry) {
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
			// (docs/bead-model/beads-are-the-edge.md step 3). The wire is no longer driven by a
			// goroutine of its own — that is what "the wire goroutine is removed" means
			// concretely, and it is why the node can read the fraction without touching
			// another goroutine's state.
			if srcNM, ok := mr.nodeGeoms[em.srcID]; ok {
				srcNM.outs.outWires = append(srcNM.outs.outWires, pw)
				srcNM.outs.outWireTargets = append(srcNM.outs.outWireTargets, em.dstID)
				// Parallel to outWires: the source *Out this edge's step count is
				// PUBLISHED through (chainBeads calls PublishSteps on it — see
				// outWireOuts' doc comment). o may be nil if this edge's source handle
				// wasn't found in outSink; chainBeads then just skips publishing for
				// this edge (it still lays the chain out — the step count is computed
				// locally either way, see edgeStepCount).
				srcNM.outs.outWireOuts = append(srcNM.outs.outWireOuts, o)
				// Parallel to outWires/outWireOuts: this edge's OWN edgeMover.stepsIn
				// channel (edge_mover.go's doc comment) — the second delivery
				// chainBeads makes alongside PublishSteps, so the edgeMover's own
				// goroutine (which cannot read the Out directly — see stepsIn's doc
				// comment) can revise an in-flight bead's remaining travel against the
				// same freshly computed count.
				srcNM.outs.outStepsIn = append(srcNM.outs.outStepsIn, em.stepsIn)
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
	// mr.nodeMovers holds ONLY ring nodes by construction (finalizeActors never builds
	// one for a node that claimed BuildArgs.ClaimSelfDrive) — there is nothing to skip
	// here, unlike the old selfDriven-flag check this replaced.
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

// finalizeActors builds the RING actor directory (mr.nodeMovers) from mr.nodeGeoms, AFTER
// buildNodes has run every kind's own build func — which is when a pair kind calls
// BuildArgs.ClaimSelfDrive (build_args.go) and so is the earliest point "which nodes
// self-drive" is fully known. Every node id NOT in claimed gets wrapped in a nodeMover and
// a fresh speed channel (per-goroutine-clock.md "Delivery"), appended to speedSinks; every
// id IN claimed gets no nodeMover at all — nothing to skip launching later, by
// construction, not by a flag. clockSrc is copied into that node's own geometry.clk lazily
// by nodeMover.run at its own goroutine start (mirrors every other per-goroutine clock use).
func (mr *moverRegistry) finalizeActors(speedSinks *[]chan float64) {
	mr.nodeMovers = map[string]*nodeMover{}
	for id, ng := range mr.nodeGeoms {
		if mr.selfDriveClaimed[id] {
			continue
		}
		nm := newNodeMover(ng)
		if speedSinks != nil {
			nodeSpeedCh := make(chan float64, 1)
			nm.speedCh = nodeSpeedCh
			*speedSinks = append(*speedSinks, nodeSpeedCh)
		}
		mr.nodeMovers[id] = nm
	}
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
	for id, nm := range mr.nodeGeoms {
		select {
		case c := <-nm.msg.centerOut:
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

// sendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn), if
// the id is a known node. This is the EXTERNAL-caller path — RootMove (drag) and
// gesture.go's dragStart send — not a mover-to-mover send (those go through a node's
// own pending/flushPending onto its OWN dedicated channel, never through this
// function), so it has no owning geometry to fire a tap through — this bare path never
// fires the test-only tap (see nodeGeometry.tap's doc comment; only enqueueFor, a node's
// own send, does). Looks up mr.nodeGeoms (every node, ring and pair alike — a drag/
// select/hover addressed to a pair node must still arrive and be handled, on that
// node's own goroutine) — a read-only directory once construction finishes, safe to
// read from any goroutine. ctx is threaded through from MoveDispatch (not part of
// moverRegistry).
func (mr *moverRegistry) sendMove(ctx context.Context, id string, msg movemsg.Msg) {
	nm, ok := mr.nodeGeoms[id]
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
		nm.msg.extIn <- msg
		return
	}
	select {
	case nm.msg.extIn <- msg:
	case <-ctx.Done():
	}
}

// enqueueFor returns nm's own non-blocking send function: it fires nm's own tap (at
// enqueue time, so tap-based tests' counts/ordering match today's behavior — a plain
// nil check + direct call, since nm.msg.tap is owned and read only by nm's own goroutine,
// which is the only caller of the closure returned here), appends the message to nm's
// own pending retry queue, and attempts an immediate flush — never blocking the calling
// handler goroutine. Bound once per node at construction (nm.sendMove = md.mr.enqueueFor(nm))
// so every send a node's own handle performs — including the ones
// fanEdgesAndPartners/requantizeLocalPolars make on that node's behalf — goes through
// nm's own retry queue, never a raw blocking channel write and never a second node's
// queue (there is no shared outbox to route through anymore).
func (mr *moverRegistry) enqueueFor(nm *nodeGeometry) func(id string, msg movemsg.Msg) {
	return func(id string, msg movemsg.Msg) {
		if nm.msg.tap != nil {
			nm.msg.tap(id, msg)
		}
		nm.msg.pending = append(nm.msg.pending, pendingSend{destID: id, msg: msg})
		nm.flushPending()
		if len(nm.msg.pending) > maxPendingSends {
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
				"nodeGeometry(%s): pending exceeded %d retry-queued sends; either a "+
					"destination's own goroutine has stopped draining its inbox "+
					"(wedged or dead), or this node is enqueueing to a peer faster "+
					"than that peer drains, cycle over cycle",
				nm.id, maxPendingSends))
		}
	}
}
