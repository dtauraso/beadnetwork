// scene_camera_test.go is an EXTERNAL test (package scenecamera_test): SeedInitialViewpoint
// and LoadSceneViewpoint need no unexported Wiring field, only the exported md.UI.VP field
// (nodes/Wiring/viewstate.UIState.VP) — a same-package Wiring test cannot import
// scenecamera, since scenecamera already imports Wiring (a real cycle).
package scenecamera_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	Wiring "github.com/dtauraso/wirefold/nodes/Wiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenecamera"
	"github.com/dtauraso/wirefold/nodes/wire"
)

// vec3 / vecClose: local test-only spellings, same pattern every other package's own test
// suite carries (nodes/Wiring/geom/gesture_camera_test.go's vecClose, etc.) — this package
// has no production vec3 alias for a test file to reuse.
type vec3 = wire.Vec3

func vecClose(a, b vec3, tol float64) bool {
	return math.Abs(a.X-b.X) < tol && math.Abs(a.Y-b.Y) < tol && math.Abs(a.Z-b.Z) < tol
}

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
func basisNonDegenerate(t *testing.T, pos, up geom.Dir) {
	t.Helper()
	b := geom.BasisFromViewpoint(pos, up)
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

	pivot, r, pos, up, ok := scenecamera.LoadSceneViewpoint(dir)
	if !ok {
		t.Fatalf("LoadSceneViewpoint: ok=false for a valid cameraPolar")
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
	md := &Wiring.MoveDispatch{}
	scenecamera.SeedInitialViewpoint(dir, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint, nil)
	vp := md.UI.VP.Viewpoint
	if !vecClose(vp.Pivot, vec3{X: 10, Y: 20, Z: 30}, 1e-9) || math.Abs(vp.R-250) > 1e-9 {
		t.Fatalf("SeedInitialViewpoint did not install the loaded pose: %+v", vp)
	}
	basisNonDegenerate(t, vp.Pos, vp.Up)

	before := vp.Pivot
	md.UI.VP.PanViewpoint(vec3{X: 5, Y: -7, Z: 2}, nil)
	if got := md.UI.VP.Viewpoint.Pivot; !vecClose(got, before.Add(vec3{X: 5, Y: -7, Z: 2}), 1e-9) {
		t.Fatalf("pan pivot=%v want %v", got, before.Add(vec3{X: 5, Y: -7, Z: 2}))
	}
}

func TestSeedInitialViewpointAbsentFileUsesDefault(t *testing.T) {
	// A fresh topology dir with no view/camera.json → the fixed non-degenerate default.
	dir := t.TempDir()
	if _, _, _, _, ok := scenecamera.LoadSceneViewpoint(dir); ok {
		t.Fatalf("LoadSceneViewpoint: ok=true for an absent camera.json")
	}

	md := &Wiring.MoveDispatch{}
	scenecamera.SeedInitialViewpoint(dir, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint, nil)
	vp := md.UI.VP.Viewpoint

	// Default: pivot=origin, r=DefaultViewpointR, pos=+Z (square-on), up=+Y.
	if !vecClose(vp.Pivot, vec3{X: 0, Y: 0, Z: 0}, 1e-9) {
		t.Fatalf("default pivot=%v want origin", vp.Pivot)
	}
	if math.Abs(vp.R-scenecamera.DefaultViewpointR) > 1e-9 {
		t.Fatalf("default r=%v want %v", vp.R, scenecamera.DefaultViewpointR)
	}
	// pos +Z, up +Y → non-degenerate basis, and pan works.
	basisNonDegenerate(t, vp.Pos, vp.Up)
	posW := geom.AnglesToWorldOffset(1, vp.Pos.Theta, vp.Pos.Phi)
	if !vecClose(posW, vec3{X: 0, Y: 0, Z: 1}, 1e-9) {
		t.Fatalf("default pos world=%v want +Z", posW)
	}
	upW := geom.AnglesToWorldOffset(1, vp.Up.Theta, vp.Up.Phi)
	if !vecClose(upW, vec3{X: 0, Y: 1, Z: 0}, 1e-9) {
		t.Fatalf("default up world=%v want +Y", upW)
	}
	md.UI.VP.PanViewpoint(vec3{X: 1, Y: 2, Z: 3}, nil)
	if got := md.UI.VP.Viewpoint.Pivot; !vecClose(got, vec3{X: 1, Y: 2, Z: 3}, 1e-9) {
		t.Fatalf("default pan pivot=%v want (1,2,3)", got)
	}
}

// A malformed / partial cameraPolar is rejected (falls back), matching parsePolarCamera
// which drops a partial object rather than reading a degenerate pose.
func TestLoadSceneViewpointRejectsPartial(t *testing.T) {
	dir := writeCameraFile(t, `{ "pivot": [1,2,3], "r": 100 }`)
	if _, _, _, _, ok := scenecamera.LoadSceneViewpoint(dir); ok {
		t.Fatalf("LoadSceneViewpoint: ok=true for a partial cameraPolar (missing pos/up)")
	}
}
