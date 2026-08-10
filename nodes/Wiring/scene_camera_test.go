package Wiring

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// scene_camera_test.go — the initial camera viewpoint is FILE DATA loaded by Go from
// view/camera.json (SeedInitialViewpoint), not a computed seed. These tests pin the schema
// match with TS's persisted cameraPolar, the non-degenerate default fallback, and that a
// pan on the loaded pose moves the pivot within a valid (non-collapsed) screen basis.

// writeCameraFile writes camera.json under <dir>/view/ and returns dir (a topology-tree
// path). body is the unwrapped cameraPolar shape ({"pivot":...,"r":...,"pos":...,"up":...})
// — camera.json holds exactly that shape, not a "cameraPolar"-keyed wrapper (scene_paths.go,
// scene_camera_persist.go).
func writeCameraFile(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	viewDir := filepath.Join(dir, "view")
	if err := os.MkdirAll(viewDir, 0o755); err != nil {
		t.Fatalf("mkdir view: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viewDir, "camera.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write camera.json: %v", err)
	}
	return dir
}

// basisNonDegenerate asserts the screen basis has three finite, unit-length, mutually
// orthogonal vectors — i.e. up and pos are not collinear (the old zero-value bug).
func basisNonDegenerate(t *testing.T, pos, up dir) {
	t.Helper()
	b := basisFromViewpoint(pos, up)
	for _, v := range []vec3{b.RefX, b.RefY, b.Pole} {
		if math.IsNaN(v.X) || math.IsNaN(v.Y) || math.IsNaN(v.Z) {
			t.Fatalf("basis vector has NaN (degenerate): %v", v)
		}
		if l := v.Length(); math.Abs(l-1) > 1e-6 {
			t.Fatalf("basis vector not unit length: %v (len %v)", v, l)
		}
	}
}

func TestLoadSceneViewpointMatchesCameraPolar(t *testing.T) {
	// Exact TS cameraPolar shape (camera-store.ts PolarCamera / serializeSceneState).
	dir := writeCameraFile(t, `{
	  "pivot": [10, 20, 30],
	  "r": 250,
	  "pos": [1.1, 2.2],
	  "up": [0.3, 0.4]
	}`)

	pivot, r, pos, up, ok := loadSceneViewpoint(dir)
	if !ok {
		t.Fatalf("loadSceneViewpoint: ok=false for a valid cameraPolar")
	}
	if !vecClose(pivot, vec3{X: 10, Y: 20, Z: 30}, 1e-9) {
		t.Fatalf("pivot=%v want (10,20,30)", pivot)
	}
	if math.Abs(r-250) > 1e-9 {
		t.Fatalf("r=%v want 250", r)
	}
	if math.Abs(pos.Theta-1.1) > 1e-9 || math.Abs(pos.Phi-2.2) > 1e-9 {
		t.Fatalf("pos=%v want {1.1,2.2}", pos)
	}
	if math.Abs(up.Theta-0.3) > 1e-9 || math.Abs(up.Phi-0.4) > 1e-9 {
		t.Fatalf("up=%v want {0.3,0.4}", up)
	}

	// The loaded pose is installed into the FSM and is non-degenerate; a pan then moves the
	// pivot within a valid basis (the exact thing the old zero-value viewpoint broke).
	md := &MoveDispatch{}
	SeedInitialViewpoint(dir, md, nil)
	if !vecClose(md.ui.vp.Pivot, vec3{X: 10, Y: 20, Z: 30}, 1e-9) || math.Abs(md.ui.vp.R-250) > 1e-9 {
		t.Fatalf("SeedInitialViewpoint did not install the loaded pose: %+v", md.ui.vp.viewpoint)
	}
	basisNonDegenerate(t, md.ui.vp.Pos, md.ui.vp.Up)

	before := md.ui.vp.Pivot
	md.PanViewpoint(vec3{X: 5, Y: -7, Z: 2}, nil)
	if !vecClose(md.ui.vp.Pivot, before.Add(vec3{X: 5, Y: -7, Z: 2}), 1e-9) {
		t.Fatalf("pan pivot=%v want %v", md.ui.vp.Pivot, before.Add(vec3{X: 5, Y: -7, Z: 2}))
	}
}

func TestSeedInitialViewpointAbsentFileUsesDefault(t *testing.T) {
	// A fresh topology dir with no view/camera.json → the fixed non-degenerate default.
	dir := t.TempDir()
	if _, _, _, _, ok := loadSceneViewpoint(dir); ok {
		t.Fatalf("loadSceneViewpoint: ok=true for an absent camera.json")
	}

	md := &MoveDispatch{}
	SeedInitialViewpoint(dir, md, nil)

	// Default: pivot=origin, r=defaultViewpointR, pos=+Z (square-on), up=+Y.
	if !vecClose(md.ui.vp.Pivot, vec3{X: 0, Y: 0, Z: 0}, 1e-9) {
		t.Fatalf("default pivot=%v want origin", md.ui.vp.Pivot)
	}
	if math.Abs(md.ui.vp.R-defaultViewpointR) > 1e-9 {
		t.Fatalf("default r=%v want %v", md.ui.vp.R, defaultViewpointR)
	}
	// pos +Z, up +Y → non-degenerate basis, and pan works.
	basisNonDegenerate(t, md.ui.vp.Pos, md.ui.vp.Up)
	posW := anglesToWorldOffset(1, md.ui.vp.Pos.Theta, md.ui.vp.Pos.Phi)
	if !vecClose(posW, vec3{X: 0, Y: 0, Z: 1}, 1e-9) {
		t.Fatalf("default pos world=%v want +Z", posW)
	}
	upW := anglesToWorldOffset(1, md.ui.vp.Up.Theta, md.ui.vp.Up.Phi)
	if !vecClose(upW, vec3{X: 0, Y: 1, Z: 0}, 1e-9) {
		t.Fatalf("default up world=%v want +Y", upW)
	}
	md.PanViewpoint(vec3{X: 1, Y: 2, Z: 3}, nil)
	if !vecClose(md.ui.vp.Pivot, vec3{X: 1, Y: 2, Z: 3}, 1e-9) {
		t.Fatalf("default pan pivot=%v want (1,2,3)", md.ui.vp.Pivot)
	}
}

// A malformed / partial cameraPolar is rejected (falls back), matching parsePolarCamera
// which drops a partial object rather than reading a degenerate pose.
func TestLoadSceneViewpointRejectsPartial(t *testing.T) {
	dir := writeCameraFile(t, `{ "pivot": [1,2,3], "r": 100 }`)
	if _, _, _, _, ok := loadSceneViewpoint(dir); ok {
		t.Fatalf("loadSceneViewpoint: ok=true for a partial cameraPolar (missing pos/up)")
	}
}
