package dispatch_test

// scene_lattice_persist_test.go — round-trip test for the pair-lattice point-count
// persister (view/lattice.json): a scene/latticePoints edit -> the synchronous writer
// persists the count to disk -> a reload reads it back into md.UI.LatticePoints. Same
// real-on-disk-tree shape as scene_speed_persist_test.go
// (memory/feedback_headless_repro_verifies_persistence.md).

import (
	"os"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// TestPersistLatticePointsRoundTrips: schedule a lattice write -> lattice.json carries the
// exact count -> a fresh LoadLatticePoints call restores md.UI.LatticePoints from disk.
func TestPersistLatticePointsRoundTrips(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	md.Persist.Lattice().Schedule(12)

	got, found := scenepersist.LoadSceneLattice(scenepaths.LatticeFilePath(root))
	if !found {
		t.Fatalf("loadSceneLattice found no points key after flush")
	}
	if got != 12 {
		t.Fatalf("lattice.json round-trip: got %v, want 12", got)
	}

	fresh := loadTreeMD(t, root)
	scenepersist.LoadLatticePoints(&fresh.UI, root)
	if fresh.UI.LatticePoints != 12 {
		t.Fatalf("LoadLatticePoints did not restore ui.latticePoints=12, got %v", fresh.UI.LatticePoints)
	}
}

// TestLoadLatticePointsFallsBackQuietlyWhenMissing: a fresh topology with no lattice.json
// falls back to the default count (24) rather than failing loudly — this is a PREFERENCE,
// not a structural invariant like counts.json.
func TestLoadLatticePointsFallsBackQuietlyWhenMissing(t *testing.T) {
	root := writeTree(t) // no view/lattice.json
	md := loadTreeMD(t, root)

	scenepersist.LoadLatticePoints(&md.UI, root)

	if md.UI.LatticePoints != scenepersist.DefaultLatticePoints {
		t.Fatalf("LoadLatticePoints with no file: got ui.latticePoints=%v, want default %v", md.UI.LatticePoints, scenepersist.DefaultLatticePoints)
	}
}

// TestLoadLatticePointsFallsBackQuietlyWhenMalformed: a lattice.json that isn't valid JSON
// falls back to the default count too — same PREFERENCE behavior as a missing file, never a
// failed load (readJSONBestEffort).
func TestLoadLatticePointsFallsBackQuietlyWhenMalformed(t *testing.T) {
	root := writeTree(t)
	if err := os.MkdirAll(scenepaths.LatticeFilePath(root)[:len(scenepaths.LatticeFilePath(root))-len("/lattice.json")], 0o755); err != nil {
		t.Fatalf("mkdir view dir: %v", err)
	}
	if err := os.WriteFile(scenepaths.LatticeFilePath(root), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write malformed lattice.json: %v", err)
	}
	md := loadTreeMD(t, root)

	scenepersist.LoadLatticePoints(&md.UI, root)

	if md.UI.LatticePoints != scenepersist.DefaultLatticePoints {
		t.Fatalf("LoadLatticePoints with malformed file: got ui.latticePoints=%v, want default %v", md.UI.LatticePoints, scenepersist.DefaultLatticePoints)
	}
}
