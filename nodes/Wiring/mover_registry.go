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
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"sync"
)

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
		if o, ok := outSink[em.srcID+"."+em.srcH]; ok {
			em.out = o
			mr.edgeOut[edgeID] = o
		}
		if pw, ok := slotReg[em.dstID+"."+em.dstH]; ok {
			em.dest = pw
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
	}
}
