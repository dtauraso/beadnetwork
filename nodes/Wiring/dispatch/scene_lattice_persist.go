// scene_lattice_persist.go — MoveDispatch-facing side of the pair lattice's POINT COUNT
// persister (view/lattice.json): the latticePersister type + BroadcastLatticePoints. Pure
// read/write helpers (WriteSceneLattice/LoadSceneLattice/LoadLatticePoints/
// SendLatticePointsNonBlocking/DefaultLatticePoints) live in
// nodes/Wiring/scenepersist/scene_lattice_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// the lattice Persister's Schedule() below — the scene/latticePoints edit handler
// (applyUpdateScene's "latticePoints" case, stdin_reader.go) is the only trigger.
// lattice.json is scene-level and genuinely singular (there is only one lattice point
// count), so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") — same
// shape as camera.json/overlays.json/sphere.json/speed.json, not a per-node split.
package dispatch

// The lattice's own file persister (view/lattice.json) is one instantiation of
// scenepersist.Persister[int32] (the shared debounce-then-write actor shape, see that type's
// own doc comment), bound to WriteSceneLattice, constructed in move_persist.go's
// EnableEditPersist and held at md.persist.lattice. Armed by EnableEditPersist, then called
// exclusively by the view-owner goroutine (RunStdinReader). Its Path == "" (tests that never
// arm) → Schedule is a no-op.

// BroadcastLatticePoints sends a new lattice point count to every node id's own dedicated
// LatticeIn channel. Thin nil-guarded delegator to md.inboxes (move_dispatch.go); the nil
// check on md itself (not on inboxes) is preserved from before this pure single-owner
// forward moved, so a bare nil *MoveDispatch still no-ops instead of panicking.
func (md *MoveDispatch) BroadcastLatticePoints(points int32) {
	if md == nil {
		return
	}
	md.inboxes.broadcastLatticePoints(points)
}
