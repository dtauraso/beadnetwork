// mover_registry.go — the nodeMover/edgeMover directory owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): moverRegistry owns
// nodeMovers/edgeMovers/edgeOut and the Bind/Start/EdgeOut/sendMove/enqueueFor/
// centerOfNode logic. MoveDispatch's public Bind/Start/EdgeOut stay as thin delegators so
// the external API is unchanged; sendMove/enqueueFor thread through md.msgTap/md.ctx
// (owned elsewhere, NOT part of this extraction) as parameters.

package Wiring

import (
	"context"
	"sync"
	"sync/atomic"
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
	edgeOut map[string]*Out
}

// bind wires the per-edge source Outs (keyed "source.sourceHandle" in outSink) and dest
// wires (slotReg, keyed "target.targetHandle") into each edgeMover. Call once after node
// construction.
func (mr *moverRegistry) bind(outSink map[string]*Out, slotReg SlotRegistry) {
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
func (mr *moverRegistry) edgeOutFor(edgeID string) *Out {
	return mr.edgeOut[edgeID]
}

// centerOfNode returns the current world center for a node id by loading the
// nodeMover's atomically-published snapshot. Safe to call from any goroutine
// without synchronization — the snap is published via atomic.Pointer after each
// center update so this never races with the mover's live geom writes.
func (mr *moverRegistry) centerOfNode(id string) (vec3, bool) {
	if nm, ok := mr.nodeMovers[id]; ok {
		if s := nm.snap.Load(); s != nil {
			return s.c, true
		}
	}
	return vec3{}, false
}

// sendMove routes one moveMsg to a node's dedicated external-entry channel (extIn), if
// the id is a known node. This is the EXTERNAL-caller path — RootMove (drag) and
// gesture.go's dragStart send — not a mover-to-mover send (those go through a mover's
// own nm.pending/flushPending onto its OWN dedicated channel, never through this
// function). mr.nodeMovers is a read-only directory once construction finishes, safe to
// read from any goroutine. msgTap/ctx are threaded through from MoveDispatch (not part
// of moverRegistry).
func (mr *moverRegistry) sendMove(msgTap *atomic.Pointer[func(destID string, msg moveMsg)], ctx context.Context, id string, msg moveMsg) {
	if tap := msgTap.Load(); tap != nil {
		(*tap)(id, msg)
	}
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

// enqueueFor returns nm's own non-blocking send function: it fires the msgTap (at enqueue time, so tap-based tests'
// counts/ordering match today's behavior), appends the message to nm's own pending
// retry queue, and attempts an immediate flush — never blocking the calling handler
// goroutine. Bound once per node at construction (nm.sendMove = md.enqueueFor(nm)) so
// every send a nodeMover's own handle performs — including the ones
// fanEdgesAndPartners/requantizeLocalPolars make on that node's behalf — goes through
// nm's own retry queue, never a raw blocking channel write and never a second mover's
// queue (there is no shared outbox to route through anymore). msgTap is threaded through
// from MoveDispatch (not part of moverRegistry).
func (mr *moverRegistry) enqueueFor(msgTap *atomic.Pointer[func(destID string, msg moveMsg)], nm *nodeMover) func(id string, msg moveMsg) {
	return func(id string, msg moveMsg) {
		if tap := msgTap.Load(); tap != nil {
			(*tap)(id, msg)
		}
		nm.pending = append(nm.pending, pendingSend{destID: id, msg: msg})
		nm.flushPending()
	}
}
