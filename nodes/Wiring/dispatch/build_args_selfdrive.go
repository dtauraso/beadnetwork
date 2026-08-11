// build_args_selfdrive.go — BuildArgs.ClaimSelfDrive, the accessor that hands a node's own
// kind goroutine direct ownership of its own nodeGeometry instead of a separate nodeMover
// actor. Split out of build_args.go — see that file's header.

package dispatch

import "github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"

// ClaimSelfDrive hands THIS node's own kind goroutine direct ownership of its own
// nodeGeometry (nodeactor.PairNodeSelf) — geometry, outgoing wires, bead chain,
// and persistence — instead of a SEPARATE nodeMover actor (task/pair-node-owns-
// itself). Call this ONLY from a kind whose own goroutine is meant to drive its own
// geometry directly (PairNode today, the pair scene): it records this node's id in
// md.selfDriveClaimed so finalizeActors (mover_registry.go, called from build.go AFTER
// every kind's build func has run) never constructs a nodeMover for it at all — there is
// no flag on a mover to skip, because no mover is ever built for this id — and returns a
// handle the caller's own Update loop uses to run that geometry's per-cycle work (Step)
// and to apply what used to be one-way notification messages to it (SetTiltIndex/
// SetReceivedVector/ClearOutBeads) as plain method calls instead — there is no longer
// anything to notify: the caller's own goroutine already IS the driver. Returns nil on a
// bare test build with no loader (a.deps.mr == nil) or if this node has no geometry entry,
// matching the nil-safe fallback every other closure in this file takes; every method on
// a nil *nodeactor.PairNodeSelf is itself a no-op.
func (a BuildArgs) ClaimSelfDrive() *nodeactor.PairNodeSelf {
	mr := a.deps.mr
	if mr == nil {
		return nil
	}
	ng, ok := mr.nodeGeoms[a.name]
	if !ok {
		return nil
	}
	if mr.selfDriveClaimed == nil {
		mr.selfDriveClaimed = map[string]bool{}
	}
	mr.selfDriveClaimed[a.name] = true
	// A ring node's NodeMover.Run Copies clockSrc into clk once, at its own goroutine
	// start. There is no such goroutine start for a self-driven node — ClaimSelfDrive
	// runs during buildNodes, single-threaded setup, before any goroutine exists — so do
	// the same copy here instead; writeStreamFrame (Step) still reads clk directly.
	ng.CopyClockSrc()
	// geom.clk also needs its OWN speed-delivery channel, exactly like a ring node's
	// NodeMover.speedCh (finalizeActors) — otherwise this clock, copied once above, never
	// hears a later speed broadcast (SceneTab.ClockDivisor or a live slider change) at all,
	// even though the kind's own SEPARATE clock (its own SpeedCh, polled in its Update
	// loop) does. See nodeactor.PairNodeSelf's own speedCh doc comment for the full defect
	// this closes.
	var speedCh chan float64
	if a.pb.SpeedSinks != nil {
		speedCh = make(chan float64, 1)
		*a.pb.SpeedSinks = append(*a.pb.SpeedSinks, speedCh)
	}
	// NewPairNodeSelf is the exported constructor package nodeactor requires now that
	// PairNodeSelf's fields are unexported across the package boundary (§20) — this call
	// site used to build the value with a bare struct literal while both types lived in
	// package Wiring.
	return nodeactor.NewPairNodeSelf(ng, speedCh)
}
