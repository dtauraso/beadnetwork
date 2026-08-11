// build_args_lattice.go — BuildArgs methods for a node's own LATTICE point count (its ring's
// point seed, and the channel that delivers a later scene-level change to it). Split out of
// build_args.go — see that file's header.

package Wiring

import "github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"

// LatticePointsSeed returns the scene's currently-loaded lattice point count
// (a.deps.latticePoints, seeded from view/lattice.json by LoadLatticePoints BEFORE
// buildNodes runs) — the load-time seed a node builds its FIRST ring at. nil-safe: on a
// bare test build with no loader (a.deps.inboxes == nil) this returns defaultLatticePoints
// (24), matching every other build-time fallback in this file.
func (a BuildArgs) LatticePointsSeed() int32 {
	if a.deps.inboxes == nil {
		return scenepersist.DefaultLatticePoints
	}
	return a.deps.latticePoints
}

// LatticeIn claims this node's dedicated inbound channel for a scene-level
// lattice-point-count change, registering it in MoveDispatch.latticeIns so
// BroadcastLatticePoints (scene_lattice_persist.go) delivers a new count to this node.
// Call this ONLY from a kind whose own goroutine owns its own lattice (PairNode) — every
// other kind simply never calls this. nil-safe: a.deps.inboxes is nil on a bare test build
// with no loader, in which case this returns a channel that is never written to
// (PollRecv-style non-blocking reads on it always find nothing, matching every other
// build-time fallback in this file).
func (a BuildArgs) LatticeIn() <-chan int32 {
	inboxes := a.deps.inboxes
	if inboxes == nil {
		return make(chan int32)
	}
	sceneToNodeLatticeIn := make(chan int32, moverInboxDepth)
	if inboxes.lattice == nil {
		inboxes.lattice = map[string]chan int32{}
	}
	inboxes.lattice[a.name] = sceneToNodeLatticeIn
	return sceneToNodeLatticeIn
}
