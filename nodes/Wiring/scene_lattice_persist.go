// scene_lattice_persist.go — persist + load the pair lattice's POINT COUNT to
// view/lattice.json (writer + loader + seed), mirroring scene_speed_persist.go's shape —
// a scene-level scalar, edited from the UI, delivered to many goroutines.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// latticePersister.schedule() below — the scene/latticePoints edit handler
// (applyUpdateScene's "latticePoints" case, stdin_reader.go) is the only trigger.
// lattice.json is scene-level and genuinely singular (there is only one lattice point
// count), so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") — same
// shape as camera.json/overlays.json/sphere.json/speed.json, not a per-node split.
//
// UNLIKE counts.json, a missing or malformed lattice.json is a PREFERENCE, not a
// structural invariant: it falls back to defaultLatticePoints quietly rather than failing
// loudly — same reasoning as speed.json.
//
// UNLIKE speed, there is no divisor/EffectiveClockSpeed-style scaling step: the lattice
// point count is delivered to nodes VERBATIM (nodes/PairNode's newRing enforces its own
// 4..64-multiple-of-4 range), so this file has no analogue to HumanEditSpeed/
// EffectiveClockSpeed — those exist only because a clock's RATE has a separate
// "setting" mode, and a point count has no such mode.

package Wiring

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// defaultLatticePoints is the point count a fresh topology (or a missing/malformed
// lattice.json) falls back to — the model's long-standing default (Wiring.FullTurnThetaIdx).
const defaultLatticePoints int32 = tiltvector.FullTurnThetaIdx

// writeSceneLattice writes the current lattice point count as the WHOLE content of
// latticePath (lattice.json) — the sole writer of that file.
func writeSceneLattice(latticePath string, points int32) error {
	obj := map[string]json.RawMessage{
		"points": json.RawMessage(formatLatticeJSON(points)),
	}
	return jsonpersist.WriteJSONAtomic(latticePath, obj)
}

// formatLatticeJSON renders points as a plain JSON integer.
func formatLatticeJSON(points int32) []byte {
	b, err := json.Marshal(points)
	if err != nil {
		b = []byte("24")
	}
	return b
}

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
	if err := writeSceneLattice(p.path, points); err != nil {
		jsonpersist.LogPersistErr("scene_lattice_persist", p.path, err)
		return
	}
}

// sceneLatticeFile is the on-disk shape of lattice.json.
type sceneLatticeFile struct {
	Points *int32 `json:"points"`
}

// loadSceneLattice reads the persisted lattice point count from latticePath
// (lattice.json). The bool return is false when the file yields no points key (fresh
// topology, or a missing/malformed file) — the caller then keeps defaultLatticePoints.
// This is a PREFERENCE, not a structural invariant, so a missing/malformed file falls
// back quietly rather than failing loudly — see jsonpersist.ReadJSONBestEffort.
func loadSceneLattice(latticePath string) (int32, bool) {
	var lf sceneLatticeFile
	jsonpersist.ReadJSONBestEffort(latticePath, &lf)
	if lf.Points == nil {
		return defaultLatticePoints, false
	}
	return *lf.Points, true
}

// loadLatticePoints reads the persisted lattice point count from lattice.json (FILE DATA)
// into ui.latticePoints. Unlike LoadSpeed, there is no broadcast-to-sinks step here: this
// runs during the build (buildMoveDispatch, before any node's own build func has claimed
// BuildArgs.LatticeIn), so a node reads the seeded value directly via
// BuildArgs.LatticePointsSeed() rather than receiving it on a channel that does not exist
// yet. A live in-process change (the scene/latticePoints edit) goes through
// BroadcastLatticePoints instead.
func (ui *uiState) loadLatticePoints(topologyPath string) {
	points, _ := loadSceneLattice(scenepaths.LatticeFilePath(topologyPath))
	ui.latticePoints = points
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

// sendLatticePointsNonBlocking delivers points to one node's buffered-1 LatticeIn
// channel WITHOUT blocking on a goroutine that may be asleep or mid-cycle. If the buffer
// already holds a stale pending value, that stale value is dropped and replaced — LATEST
// WINS is correct because the point count is absolute state, not an event stream (same
// reasoning as wire.SendSpeedNonBlocking's own doc comment). ch must be a channel this
// call's caller alone sends on (the stdin-reader goroutine, sole writer of every
// registered latticeIns channel).
func sendLatticePointsNonBlocking(ch chan int32, points int32) {
	select {
	case ch <- points:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- points:
	default:
	}
}
