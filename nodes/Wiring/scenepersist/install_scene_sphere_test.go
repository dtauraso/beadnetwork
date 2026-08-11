package scenepersist

// install_scene_sphere_test.go — round-trip/content-fit coverage for InstallSceneSphere and
// the sphere Persister's flush, moved from nodes/Wiring/dispatch/scene_sphere_persist_test.go
// (docs/planning/movedispatch-decomposition.md §34): none of it drove anything beyond
// viewstate.UIState/geomseeds.GeomSeeds fields and this package's own functions.

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// seedsFromCenters builds a nodeSeeds slice (the frozen load-time set loadTimeCenters
// rebuilds from) directly from world centers, for tests that don't go through a real
// newMoveDispatch/LoadTopology pass.
func seedsFromCenters(centers map[string]wire.Vec3) []geomseeds.NodeGeomSeed {
	out := make([]geomseeds.NodeGeomSeed, 0, len(centers))
	for id, c := range centers {
		out = append(out, geomseeds.NodeGeomSeed{ID: id, CX: c.X, CY: c.Y, CZ: c.Z})
	}
	return out
}

// TestSceneSphereDefaultsFromContentFit: with no persisted sphere, InstallSceneSphere falls
// back to a content-fit of the node centers rather than a zero sphere.
func TestSceneSphereDefaultsFromContentFit(t *testing.T) {
	ui := &viewstate.UIState{}
	gs := &geomseeds.GeomSeeds{}
	gs.NodeSeeds = seedsFromCenters(map[string]wire.Vec3{
		"a": {X: 0, Y: 0, Z: 0},
		"b": {X: 100, Y: 0, Z: 0},
	})
	// InstallSceneSphere's content-fit path reads gs.LoadTimeCenters() (rebuilt from the
	// frozen gs.NodeSeeds set above), not an atomic snap.
	InstallSceneSphere(ui, gs, t.TempDir()) // no scene.json → content-fit
	if ui.SceneSphere.Radius <= 0 {
		t.Fatalf("content-fit sphere has non-positive radius: %+v", ui.SceneSphere)
	}
	// Center should be the bbox midpoint (≈ (50,0,0)), not the origin default.
	if ui.SceneSphere.Center.X < 40 || ui.SceneSphere.Center.X > 60 {
		t.Fatalf("content-fit center X=%v, want ≈50", ui.SceneSphere.Center.X)
	}
}

// TestSceneSphereContentFitSurvivesReloadAfterMove pins the invariant the content-fit
// persist exists for: the scene center must NOT be re-derived from moved nodes.
//
// Every node position is a scene polar measured about this center, so a center that shifts
// between runs silently reinterprets the whole diagram. Load 1 content-fits S1 and the user
// drags; load 2 must still see S1. If the fallback is not persisted, load 2 content-fits
// over the NEW centers, gets S2 != S1, and every position drifts.
//
// A plain "does it write the file" assertion would NOT catch that — it passes while the
// second load still recomputes. Drive two real loads across a move instead.
func TestSceneSphereContentFitSurvivesReloadAfterMove(t *testing.T) {
	dir := t.TempDir()

	newState := func(bx float64) (*viewstate.UIState, *geomseeds.GeomSeeds) {
		ui := &viewstate.UIState{}
		gs := &geomseeds.GeomSeeds{}
		gs.NodeSeeds = seedsFromCenters(map[string]wire.Vec3{
			"a": {X: 0, Y: 0, Z: 0},
			"b": {X: bx, Y: 0, Z: 0},
		})
		return ui, gs
	}

	// Load 1: no scene.json → content-fit S1, which must be persisted.
	ui1, gs1 := newState(100)
	InstallSceneSphere(ui1, gs1, dir)
	s1 := ui1.SceneSphere
	if s1.Radius <= 0 {
		t.Fatalf("load 1: content-fit sphere has non-positive radius: %+v", s1)
	}

	// The user drags node b far away. Its scene polar was measured about S1.
	// Load 2: a NEW process over the MOVED tree. It must read S1 back, not re-fit.
	ui2, gs2 := newState(900)
	InstallSceneSphere(ui2, gs2, dir)
	s2 := ui2.SceneSphere

	if s2.Center != s1.Center || s2.Radius != s1.Radius {
		t.Fatalf("scene sphere drifted across reload after a move:\n  load 1: %+v\n  load 2: %+v\n"+
			"every node's scenePolar is measured about this center, so the diagram would shift.", s1, s2)
	}
}

func TestSceneSpherePersisterFlushNow(t *testing.T) {
	dir := t.TempDir()
	p := &Persister[geom.SceneSphere]{
		Path: scenepaths.SphereFilePath(dir), Write: WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	s := geom.SceneSphere{Center: wire.Vec3{X: 1, Y: 2, Z: 3}, Radius: 40}
	p.Schedule(s)

	got, ok := LoadSceneSphere(dir)
	if !ok {
		t.Fatal("LoadSceneSphere: ok=false after flushNow")
	}
	if got != s {
		t.Fatalf("flushNow round-trip: got %+v want %+v", got, s)
	}
}
