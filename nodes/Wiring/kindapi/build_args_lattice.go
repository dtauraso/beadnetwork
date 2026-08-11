// build_args_lattice.go — BuildArgs methods for a node's own LATTICE point count (its ring's
// point seed, and the channel that delivers a later scene-level change to it). Split out of
// build_args.go — see that file's header.

package kindapi

import "github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"

// LatticePointsSeed returns the scene's currently-loaded lattice point count
// (a.deps.LatticePoints, seeded from view/lattice.json by LoadLatticePoints BEFORE
// buildNodes runs) — the load-time seed a node builds its FIRST ring at. nil-safe: on a
// bare test build with no loader (a.deps.ClaimLatticeIn == nil) this returns
// scenepersist.DefaultLatticePoints (24), matching every other build-time fallback in this
// file.
func (a BuildArgs) LatticePointsSeed() int32 {
	if a.deps.ClaimLatticeIn == nil {
		return scenepersist.DefaultLatticePoints
	}
	return a.deps.LatticePoints
}

// LatticeIn claims this node's dedicated inbound channel for a scene-level
// lattice-point-count change, via the dispatch core's bound a.deps.ClaimLatticeIn (which
// registers it in the dispatch core's inbox directory so BroadcastLatticePoints delivers a
// new count to this node — see BuildDeps' own doc comment for why this crosses the package
// boundary as a func value rather than a pointer into the dispatch core). Call this ONLY
// from a kind whose own goroutine owns its own lattice (PairNode) — every other kind simply
// never calls this. nil-safe: a.deps.ClaimLatticeIn is nil on a bare test build with no
// loader, in which case this returns a channel that is never written to (PollRecv-style
// non-blocking reads on it always find nothing, matching every other build-time fallback in
// this file).
func (a BuildArgs) LatticeIn() <-chan int32 {
	if a.deps.ClaimLatticeIn == nil {
		return make(chan int32)
	}
	return a.deps.ClaimLatticeIn(a.name)
}
