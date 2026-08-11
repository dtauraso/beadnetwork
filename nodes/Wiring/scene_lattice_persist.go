// scene_lattice_persist.go — MoveDispatch-facing side of the pair lattice's POINT COUNT
// persister (view/lattice.json): the latticePersister type + BroadcastLatticePoints. Pure
// read/write helpers (WriteSceneLattice/LoadSceneLattice/LoadLatticePoints/
// SendLatticePointsNonBlocking/DefaultLatticePoints) live in
// nodes/Wiring/scenepersist/scene_lattice_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// latticePersister.schedule() below — the scene/latticePoints edit handler
// (applyUpdateScene's "latticePoints" case, stdin_reader.go) is the only trigger.
// lattice.json is scene-level and genuinely singular (there is only one lattice point
// count), so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") — same
// shape as camera.json/overlays.json/sphere.json/speed.json, not a per-node split.
package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// latticePersister writes the lattice point count to lattice.json as it changes. Armed by
// EnableEditPersist, then called exclusively by the view-owner goroutine (RunStdinReader).
// path == "" (tests that never arm) → no-op.
type latticePersister struct {
	path string // lattice.json path (scenepaths.LatticeFilePath(topologyPath))
}

// schedule writes the given point count to lattice.json synchronously.
func (p *latticePersister) schedule(points int32) {
	if p == nil || p.path == "" {
		return
	}
	if err := scenepersist.WriteSceneLattice(p.path, points); err != nil {
		jsonpersist.LogPersistErr("scene_lattice_persist", p.path, err)
		return
	}
}

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
