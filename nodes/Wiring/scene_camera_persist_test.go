package Wiring

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// scene_camera_persist_test.go — the WRITE side of camera-viewpoint-as-file-data. These
// pin: a full FSM→file→loadSceneViewpoint round-trip, debounce coalescing of a rapid
// gesture burst, and that camera.json and overlays.json — separate files, one writer each
// — never clobber one another.
//
// [allow-test-weakening] TestPersistLoadsLegacySceneJSONThenWritesNewFile was DELETED: it
// pinned the pre-split view/scene.json fallback (loadSceneViewpoint reading a shared
// scene.json's cameraPolar key when camera.json is absent), a capability removed along with
// sceneJSONPath/sceneCameraPath — no such file exists anywhere in this repo's tree and
// nothing ever wrote it once the one-file-per-writer split landed. Deleting the test is
// honest: the behavior it asserted no longer exists. A topology directory holding only that
// legacy sidecar now loses its camera pose on load and falls back to defaultViewpoint.

// vpEqual asserts two FSM viewpoint tuples match within tolerance.
func vpEqual(t *testing.T, gotPivot vec3, gotR float64, gotPos, gotUp dir, wantPivot vec3, wantR float64, wantPos, wantUp dir) {
	t.Helper()
	if !vecClose(gotPivot, wantPivot, 1e-9) {
		t.Fatalf("pivot=%v want %v", gotPivot, wantPivot)
	}
	if math.Abs(gotR-wantR) > 1e-9 {
		t.Fatalf("r=%v want %v", gotR, wantR)
	}
	if math.Abs(gotPos.Theta-wantPos.Theta) > 1e-9 || math.Abs(gotPos.Phi-wantPos.Phi) > 1e-9 {
		t.Fatalf("pos=%v want %v", gotPos, wantPos)
	}
	if math.Abs(gotUp.Theta-wantUp.Theta) > 1e-9 || math.Abs(gotUp.Phi-wantUp.Phi) > 1e-9 {
		t.Fatalf("up=%v want %v", gotUp, wantUp)
	}
}

// TestPersistViewpointRoundTrips sets a viewpoint via the FSM, triggers the persist, and
// reads the file back with loadSceneViewpoint — the exact inverse — asserting equality.
func TestPersistViewpointRoundTrips(t *testing.T) {
	td := t.TempDir()
	md := &MoveDispatch{}
	md.EnableViewpointPersist(td)

	wantPivot := vec3{X: 10, Y: 20, Z: 30}
	wantR := 250.0
	wantPos := dir{Theta: 1.1, Phi: 2.2}
	wantUp := dir{Theta: 0.3, Phi: 0.4}

	md.SetViewpoint(wantPivot, wantR, wantPos, wantUp)
	md.EmitViewpoint(nil) // synchronously writes camera.json (gesture path is EmitViewpoint)

	pivot, r, pos, up, ok := loadSceneViewpoint(td)
	if !ok {
		t.Fatalf("loadSceneViewpoint: ok=false after persist")
	}
	vpEqual(t, pivot, r, pos, up, wantPivot, wantR, wantPos, wantUp)
}

// TestPersistWriteBurstLandsFinalValue schedules many rapid viewpoint changes (a drag
// burst) and asserts the FINAL value is what's on disk. Each schedule() call now writes
// synchronously (the debounce that used to coalesce a burst into one write was removed —
// see scene_persist.go's header comment: unmeasured, and writeJSONAtomic does no fsync so
// the OS already coalesces at the page-cache level). What matters, and what this asserts,
// is that the on-disk state after the burst is correct — not how many writes it took to
// get there.
func TestPersistWriteBurstLandsFinalValue(t *testing.T) {
	td := t.TempDir()
	md := &MoveDispatch{}
	md.EnableViewpointPersist(td)
	md.SetViewpoint(vec3{}, 1, dir{}, dir{})
	for i := 0; i < 50; i++ {
		md.SetViewpoint(vec3{X: float64(i)}, float64(100+i), dir{Theta: float64(i) * 0.01}, dir{Phi: float64(i) * 0.02})
		md.EmitViewpoint(nil) // synchronous write, every call
	}

	// Final value is the last scheduled viewpoint (i=49).
	pivot, r, pos, up, ok := loadSceneViewpoint(td)
	if !ok {
		t.Fatalf("loadSceneViewpoint: ok=false")
	}
	vpEqual(t, pivot, r, pos, up,
		vec3{X: 49}, 149, dir{Theta: 49 * 0.01}, dir{Phi: 49 * 0.02})
}

// TestCameraAndOverlaysFilesDoNotClobber pins the point of this split: camera.json and
// overlays.json are DIFFERENT files, each with exactly one writer, so writing one can never
// touch the other's content (the old scene.json required a shared-document lock to make
// this true; the split makes it true by construction).
func TestCameraAndOverlaysFilesDoNotClobber(t *testing.T) {
	td := t.TempDir()

	// Go persists a camera first.
	md := &MoveDispatch{}
	md.EnableViewpointPersist(td)
	md.SetViewpoint(vec3{X: 11, Y: 22, Z: 33}, 321, dir{Theta: 0.5, Phi: 1.5}, dir{Theta: 0.05, Phi: 0.15})
	md.EmitViewpoint(nil)

	// A save persists Go's OWN overlay state into a SEPARATE file (overlays.json).
	ov := defaultOverlayState()
	ov.labelsGlobalVisible = false
	if err := writeSceneOverlays(overlaysFilePath(td), ov); err != nil {
		t.Fatalf("writeSceneOverlays: %v", err)
	}

	// The Go-owned camera survives — untouched, because overlays.json is a different file.
	pivot, r, pos, up, ok := loadSceneViewpoint(td)
	if !ok {
		t.Fatalf("loadSceneViewpoint: ok=false")
	}
	vpEqual(t, pivot, r, pos, up, vec3{X: 11, Y: 22, Z: 33}, 321, dir{Theta: 0.5, Phi: 1.5}, dir{Theta: 0.05, Phi: 0.15})

	// camera.json holds ONLY the camera pose — no overlay keys leaked in.
	raw, err := os.ReadFile(filepath.Join(td, "view", "camera.json"))
	if err != nil {
		t.Fatalf("read camera.json: %v", err)
	}
	var camObj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &camObj); err != nil {
		t.Fatalf("camera.json invalid: %v", err)
	}
	if _, ok := camObj["labelsGlobalHidden"]; ok {
		t.Fatalf("camera.json unexpectedly carries an overlay key")
	}

	// overlays.json holds the overlay field.
	raw, err = os.ReadFile(filepath.Join(td, "view", "overlays.json"))
	if err != nil {
		t.Fatalf("read overlays.json: %v", err)
	}
	var ovObj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &ovObj); err != nil {
		t.Fatalf("overlays.json invalid: %v", err)
	}
	if string(ovObj["labelsGlobalHidden"]) != "true" {
		t.Fatalf("labelsGlobalHidden=%s want true (overlay field should persist)", ovObj["labelsGlobalHidden"])
	}
}
