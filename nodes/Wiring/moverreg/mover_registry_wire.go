// mover_registry_wire.go — MoverRegistry's construction-time wiring (Bind) and actor
// launch (Start, FinalizeActors), split out of mover_registry.go by concern (same
// same-package, no-API-change split Trace/Trace.go used: struct+directory accessors stay
// in mover_registry.go, wiring/launch moves here, read-only query/lookup moves to
// mover_registry_query.go, the pure link-refusal decision moves to
// mover_registry_linkrefusal.go).
package moverreg

import (
	"context"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// Bind wires the per-edge source Outs (keyed "source.sourceHandle" in outSink) and dest
// wires (slotReg, keyed "target.targetHandle") into each edgeMover. Call once after node
// construction.
func (mr *MoverRegistry) Bind(outSink map[string]*wire.Out, slotReg inputcodec.SlotRegistry) {
	for edgeID, em := range mr.edgeMovers {
		var o *wire.Out
		if oo, ok := outSink[em.SrcID()+"."+em.SrcHandle()]; ok {
			o = oo
			em.SetOut(oo)
			mr.edgeOut[edgeID] = oo
		}
		if pw, ok := slotReg[em.DstID()+"."+em.DstHandle()]; ok {
			em.SetDest(pw)
			// The SOURCE node also takes this wire, paired with the outTargets entry for
			// the same edge: the source node's own goroutine drives it (NodeMover.Run)
			// and reads its in-flight fractions to light its own chain
			// (docs/bead-model/beads-are-the-edge.md step 3). The wire is no longer driven by a
			// goroutine of its own — that is what "the wire goroutine is removed" means
			// concretely, and it is why the node can read the fraction without touching
			// another goroutine's state.
			// o may be nil if this edge's source handle wasn't found in outSink;
			// chainBeads then just skips publishing for this edge (it still lays the
			// chain out — the step count is computed locally either way, see
			// edgeStepCount). em.SendSteps is the second delivery chainBeads makes
			// alongside PublishSteps, so the edgeMover's own goroutine (which cannot
			// read the Out directly — see nodeOuts.outStepsIn's own doc comment) can
			// revise an in-flight bead's remaining travel against the same freshly
			// computed count.
			if srcNM, ok := mr.nodeGeoms[em.SrcID()]; ok {
				srcNM.AddOutWire(pw, em.DstID(), o, em.SendSteps)
			}
		}
	}
}

// Start launches every mover's goroutine — ONE goroutine per node and ONE per edge, no
// dedicated sender/watcher goroutines (an earlier shared-outbox-plus-sender-goroutine
// design was removed: each mover's own run loop drains its own inbox AND retries its own
// pending sends, non-blockingly, every cycle).
//
// Returns a *sync.WaitGroup covering every launched goroutine, so a caller that wants a
// complete shutdown (main.go: "wait for everything, then close" — see
// the wait-for-everything-then-close change) can wg.Wait() on it after cancelling
// ctx. Both nm.Run and em.Run select on ctx.Done() at the top of their loop (their only
// blocking call is SleepCycle, which also selects on ctx), so cancel-to-return is one
// clock tick, worst case. Callers that don't care about shutdown completeness (most
// existing tests) can ignore the return value — Start(ctx) alone still compiles and
// still launches every goroutine exactly as before.
func (mr *MoverRegistry) Start(ctx context.Context) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	// mr.nodeMovers holds ONLY ring nodes by construction (FinalizeActors never builds
	// one for a node that claimed BuildArgs.ClaimSelfDrive) — there is nothing to skip
	// here, unlike the old selfDriven-flag check this replaced.
	for _, nm := range mr.nodeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nm.Run(ctx)
		}()
	}
	for _, em := range mr.edgeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Run(ctx)
		}()
	}
	return wg
}

// FinalizeActors builds the RING actor directory (mr.nodeMovers) from mr.nodeGeoms, AFTER
// buildNodes has run every kind's own build func — which is when a pair kind calls
// BuildArgs.ClaimSelfDrive (build_args_selfdrive.go) and so is the earliest point "which
// nodes self-drive" is fully known. Every node id NOT in claimed gets wrapped in a
// nodeactor.NodeMover and a fresh speed channel (per-goroutine-clock.md "Delivery"),
// appended to speedSinks; every id IN claimed gets no NodeMover at all — nothing to skip
// launching later, by construction, not by a flag. clockSrc is copied into that node's own
// geometry.clk lazily by NodeMover.Run at its own goroutine start (mirrors every other
// per-goroutine clock use).
func (mr *MoverRegistry) FinalizeActors(speedSinks *[]chan float64) {
	mr.nodeMovers = map[string]*nodeactor.NodeMover{}
	for id, ng := range mr.nodeGeoms {
		if mr.selfDriveClaimed[id] {
			continue
		}
		nm := nodeactor.NewNodeMover(ng)
		if speedSinks != nil {
			nodeSpeedCh := make(chan float64, 1)
			nm.SetSpeedCh(nodeSpeedCh)
			*speedSinks = append(*speedSinks, nodeSpeedCh)
		}
		mr.nodeMovers[id] = nm
	}
}
