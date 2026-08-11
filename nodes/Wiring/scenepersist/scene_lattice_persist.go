// scene_lattice_persist.go — pure read/write helpers for the pair lattice's POINT COUNT in
// view/lattice.json. nodes/Wiring/dispatch's own scene_lattice_persist.go (the
// MoveDispatch-facing latticePersister type + BroadcastLatticePoints wrapper) was deleted
// (docs/planning/movedispatch-decomposition.md, the remainder cluster): the Persister[int32]
// this file's own writer is bound to is armed by EnableEditPersist (move_persist.go) and
// reached via md.Persist.Lattice(); BroadcastLatticePoints itself was a pure forward onto
// nodeinbox.NodeInboxes — every caller now addresses md.Inboxes.BroadcastLatticePoints
// directly.
//
// UNLIKE counts.json, a missing or malformed lattice.json is a PREFERENCE, not a
// structural invariant: it falls back to DefaultLatticePoints quietly rather than failing
// loudly — same reasoning as speed.json.
package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// DefaultLatticePoints is the point count a fresh topology (or a missing/malformed
// lattice.json) falls back to — the model's long-standing default (Wiring.FullTurnThetaIdx).
const DefaultLatticePoints int32 = tiltvector.FullTurnThetaIdx

// WriteSceneLattice writes the current lattice point count as the WHOLE content of
// latticePath (lattice.json) — the sole writer of that file.
func WriteSceneLattice(latticePath string, points int32) error {
	obj := map[string]json.RawMessage{
		"points": json.RawMessage(FormatLatticeJSON(points)),
	}
	return jsonpersist.WriteJSONAtomic(latticePath, obj)
}

// FormatLatticeJSON renders points as a plain JSON integer.
func FormatLatticeJSON(points int32) []byte {
	b, err := json.Marshal(points)
	if err != nil {
		b = []byte("24")
	}
	return b
}

// sceneLatticeFile is the on-disk shape of lattice.json.
type sceneLatticeFile struct {
	Points *int32 `json:"points"`
}

// LoadSceneLattice reads the persisted lattice point count from latticePath
// (lattice.json). The bool return is false when the file yields no points key (fresh
// topology, or a missing/malformed file) — the caller then keeps DefaultLatticePoints.
// This is a PREFERENCE, not a structural invariant, so a missing/malformed file falls
// back quietly rather than failing loudly — see jsonpersist.ReadJSONBestEffort.
func LoadSceneLattice(latticePath string) (int32, bool) {
	var lf sceneLatticeFile
	jsonpersist.ReadJSONBestEffort(latticePath, &lf)
	if lf.Points == nil {
		return DefaultLatticePoints, false
	}
	return *lf.Points, true
}

// LoadLatticePoints reads the persisted lattice point count from lattice.json (FILE DATA)
// into ui.LatticePoints. Unlike LoadSpeed, there is no broadcast-to-sinks step here: this
// runs during the build (buildMoveDispatch, before any node's own build func has claimed
// BuildArgs.LatticeIn), so a node reads the seeded value directly via
// BuildArgs.LatticePointsSeed() rather than receiving it on a channel that does not exist
// yet. A live in-process change (the scene/latticePoints edit) goes through
// MoveDispatch.BroadcastLatticePoints instead.
func LoadLatticePoints(ui *viewstate.UIState, topologyPath string) {
	points, _ := LoadSceneLattice(scenepaths.LatticeFilePath(topologyPath))
	ui.LatticePoints = points
}

// SendLatticePointsNonBlocking delivers points to one node's buffered-1 LatticeIn
// channel WITHOUT blocking on a goroutine that may be asleep or mid-cycle. If the buffer
// already holds a stale pending value, that stale value is dropped and replaced — LATEST
// WINS is correct because the point count is absolute state, not an event stream (same
// reasoning as wire.SendSpeedNonBlocking's own doc comment). ch must be a channel this
// call's caller alone sends on (the stdin-reader goroutine, sole writer of every
// registered latticeIns channel).
func SendLatticePointsNonBlocking(ch chan int32, points int32) {
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
