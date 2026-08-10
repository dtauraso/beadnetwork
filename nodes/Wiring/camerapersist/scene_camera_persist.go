// Package camerapersist is the WRITE (and shared JSON shape) side of
// camera-viewpoint-as-file-data.
//
// The read side (Wiring's scene_camera.go) loads the saved polar camera from
// `<topologyPath>/view/camera.json` into the gesture-FSM viewpoint on startup, using the
// PolarCamera shape defined here. This package is the mirror: whenever a GESTURE changes
// the FSM viewpoint (orbit/zoom/pan/home), Go persists the current viewpoint back to that
// same file, in the EXACT schema the read side reads, so navigate-then-reload round-trips.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go — the single goroutine
// that reads stdin and dispatches gestures/edits) is the SOLE caller of Schedule below, via
// EmitViewpoint (viewpoint_state.go), directly or through the gesture FSM (gesture.go).
// camera.json is scene-level and genuinely singular (there is only ever one camera pose),
// so — unlike a node's own files, which each node's own mover writes — this stays one file
// with one named owning goroutine rather than a per-entity split
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path").
//
// Go owns persistence (MODEL.md): there is no TS→Go camera-save on the new path. The
// write is:
//   - SYNCHRONOUS: Schedule writes camera.json immediately, inline on the view-owner
//     goroutine (every gesture path serializes through it, so there is only ever one
//     writer). No debounce: see scene_persist.go's header comment for why the prior 250ms
//     coalescing window was removed.
//   - WHOLE-FILE: camera.json holds ONLY the camera pose (one-file-per-writer,
//     the one-file-per-writer split) — no other writer touches it,
//     so each write marshals the pose fresh and overwrites the file, no read-modify-write.
//   - FIRE-AND-FORGET: errors are logged, not returned; it never blocks the gesture.
//
// Before this split, camera.json's content lived at a shared view/scene.json's
// `cameraPolar` key, alongside the overlays and sphere writers. That shared sidecar and its
// best-effort read fallback were REMOVED (scene_paths.go's header) — a topology holding
// only the old sidecar now falls straight to loadSceneViewpoint's defaultViewpoint.
//
// The crash-safe (tmp-then-rename) write plumbing is shared machinery from
// nodes/Wiring/jsonpersist (WriteJSONAtomic) — this package holds only the camera-specific
// shape.
package camerapersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

// PolarCamera mirrors TS's persisted `cameraPolar` JSON shape exactly (camera-store.ts
// PolarCamera / parse.ts parsePolarCamera):
//
//	{ "pivot": [x,y,z], "r": n, "pos": [theta,phi], "up": [theta,phi] }
//
// Pointer fields distinguish "absent" from a legitimate zero so a partial object is
// rejected rather than silently read as a degenerate pose.
type PolarCamera struct {
	Pivot *[3]float64 `json:"pivot"`
	R     *float64    `json:"r"`
	Pos   *[2]float64 `json:"pos"`
	Up    *[2]float64 `json:"up"`
}

// ViewpointPersister writes viewpoint changes to camera.json as they happen. Armed after
// the startup seed (EnableViewpointPersist), then called exclusively by the view-owner
// goroutine (RunStdinReader) — see the OWNER note above.
type ViewpointPersister struct {
	Path string // camera.json path (scenepaths.CameraFilePath(topologyPath))
}

// Schedule writes the given viewpoint to camera.json synchronously. Fire-and-forget:
// errors are logged, not returned.
func (p *ViewpointPersister) Schedule(v geom.Viewpoint) {
	if p == nil || p.Path == "" {
		return
	}
	cam := ViewpointToPolar(v)
	if err := WriteSceneCameraPolar(p.Path, cam); err != nil {
		jsonpersist.LogPersistErr("scene_camera_persist", p.Path, err)
		return
	}
}

// ViewpointToPolar converts an FSM viewpoint to the persisted cameraPolar shape. It is the
// exact inverse of the read side's mapping (Wiring's loadSceneViewpoint), so a
// load→persist→load round-trips.
func ViewpointToPolar(v geom.Viewpoint) *PolarCamera {
	pivot := [3]float64{v.Pivot.X, v.Pivot.Y, v.Pivot.Z}
	r := v.R
	pos := [2]float64{v.Pos.Theta, v.Pos.Phi}
	up := [2]float64{v.Up.Theta, v.Up.Phi}
	return &PolarCamera{Pivot: &pivot, R: &r, Pos: &pos, Up: &up}
}

// WriteSceneCameraPolar writes cam as the whole content of path (camera.json) — the sole
// writer of that file, so no read-modify-write is needed.
func WriteSceneCameraPolar(path string, cam *PolarCamera) error {
	return jsonpersist.WriteJSONAtomic(path, cam)
}
