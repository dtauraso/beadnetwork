// build_args_selfdrive.go — BuildArgs.ClaimSelfDrive, the accessor that hands a node's own
// kind goroutine direct ownership of its own nodeGeometry instead of a separate nodeMover
// actor. Split out of build_args.go — see that file's header.

package Wiring

// ClaimSelfDrive hands THIS node's own kind goroutine direct ownership of its own
// nodeGeometry (PairNodeSelf, pair_node_self.go) — geometry, outgoing wires, bead chain,
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
// bare test build with no loader (currentBuildMD == nil) or if this node has no geometry entry,
// matching the nil-safe fallback every other closure in this file takes; every method on
// a nil *PairNodeSelf is itself a no-op.
func (a BuildArgs) ClaimSelfDrive() *PairNodeSelf {
	md := currentBuildMD
	if md == nil {
		return nil
	}
	ng, ok := md.mr.nodeGeoms[a.name]
	if !ok {
		return nil
	}
	if md.mr.selfDriveClaimed == nil {
		md.mr.selfDriveClaimed = map[string]bool{}
	}
	md.mr.selfDriveClaimed[a.name] = true
	// A ring node's nodeMover.run Copies clockSrc into clk once, at its own goroutine
	// start. There is no such goroutine start for a self-driven node — ClaimSelfDrive
	// runs during buildNodes, single-threaded setup, before any goroutine exists — so do
	// the same copy here instead; writeStreamFrame (Step) still reads clk directly.
	if ng.clocks.clockSrc != nil {
		ng.clocks.clk = ng.clocks.clockSrc.Copy()
	}
	// geom.clk also needs its OWN speed-delivery channel, exactly like a ring node's
	// nodeMover.speedCh (finalizeActors) — otherwise this clock, copied once above, never
	// hears a later speed broadcast (SceneTab.ClockDivisor or a live slider change) at all,
	// even though the kind's own SEPARATE clock (its own SpeedCh, polled in its Update
	// loop) does. See PairNodeSelf.speedCh's own doc comment for the full defect this closes.
	var self *PairNodeSelf
	if a.pb.SpeedSinks != nil {
		speedCh := make(chan float64, 1)
		*a.pb.SpeedSinks = append(*a.pb.SpeedSinks, speedCh)
		self = &PairNodeSelf{geom: ng, speedCh: speedCh}
	} else {
		self = &PairNodeSelf{geom: ng}
	}
	return self
}
