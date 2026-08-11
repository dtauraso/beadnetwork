// scene_sphere_persist.go — pure read/write helpers for the first-class SCENE SPHERE
// (sphere_layout.go sceneSphere) in view/sphere.json, plus InstallSceneSphere (the
// view-owner-goroutine-facing entry point that seeds ui.SceneSphere from disk or
// content-fits it and emits it) — the former dispatch.MoveDispatch.LoadSceneSphere method,
// moved here since it read/wrote nothing but *viewstate.UIState and *geomseeds.GeomSeeds
// (docs/planning/movedispatch-decomposition.md §34). sphere.json has exactly one writer
// (WriteSceneSphere), so each write is a fresh whole-file marshal — no read-modify-write.
// The sphere Persister instance itself still lives in nodes/Wiring/viewpersist
// (Persisters.sphere, armed by EnableEditPersist).
//
// On-disk shape:
//
//	{ "sceneSphere": { "center": [x,y,z], "radius": n } }
//
// Pointer fields distinguish "absent" from a legitimate zero so a partial object is
// rejected (→ content-fit default) rather than silently read as a degenerate sphere.
package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

type sceneSphereJSON struct {
	Center *[3]float64 `json:"center"`
	Radius *float64    `json:"radius"`
}

// LoadSceneSphere reads the persisted scene sphere from sphere.json. ok is false when it
// yields no complete sphere — callers then content-fit.
func LoadSceneSphere(topologyPath string) (geom.SceneSphere, bool) {
	var sj sceneSphereJSON
	jsonpersist.ReadJSONBestEffort(scenepaths.SphereFilePath(topologyPath), &sj)
	if sj.Center == nil || sj.Radius == nil {
		return geom.SceneSphere{}, false
	}
	return geom.SceneSphere{
		Center: wire.Vec3{X: sj.Center[0], Y: sj.Center[1], Z: sj.Center[2]},
		Radius: *sj.Radius,
	}, true
}

// WriteSceneSphere writes the scene sphere as the whole content of sphereJSONPath
// (sphere.json) — the sole writer of that file, so no read-modify-write is needed.
func WriteSceneSphere(sphereJSONPath string, s geom.SceneSphere) error {
	center := [3]float64{s.Center.X, s.Center.Y, s.Center.Z}
	radius := s.Radius
	return jsonpersist.WriteJSONAtomic(sphereJSONPath, sceneSphereJSON{Center: &center, Radius: &radius})
}

// InstallSceneSphere installs ui.SceneSphere from FILE DATA, or — when sphere.json has no
// persisted sphere — from a one-time content-fit of the current node centers (so an
// existing scene gets a sane reference without any authored value). Call after LoadTopology
// (node centers are loaded) and before the sphere is used to derive positions.
//
// A content-fit fallback is PERSISTED IMMEDIATELY, and that is load-bearing. Every node's
// position is a scene polar measured ABOUT THIS CENTER, so the center must be the same on
// the next load or every position is silently reinterpreted:
//
//	load 1: no sphere -> content-fit S1 -> user drags -> scenePolars persisted about S1
//	load 2: still no sphere -> content-fit over the NEW centers -> S2 != S1
//	        -> every scenePolar now read about S2 -> the whole diagram drifts
//
// Go owns the authoritative scene state (MODEL.md), so its durability must not depend on the
// webview sending a command. Persisting here removes the last reason for TS to trigger a
// save at all.
func InstallSceneSphere(ui *viewstate.UIState, gs *geomseeds.GeomSeeds, topologyPath string) {
	if s, ok := LoadSceneSphere(topologyPath); ok {
		ui.SceneSphere = s
	} else {
		// InstallSceneSphere runs on the main goroutine BEFORE Start launches any mover
		// goroutine and before RunStdinReader's dispatch loop begins, so gs's positions
		// (which heldCenters reads) are still empty here — use the load-time geom sweep
		// instead (safe: no mover goroutine is mutating geom yet).
		ui.SceneSphere = geom.ContentFitSceneSphere(gs.LoadTimeCenters())
		// Best-effort: a read-only or absent scene dir must not stop the sim from running.
		// The in-memory sphere is correct either way; only cross-run stability is at stake.
		// Path via scenepaths.SphereFilePath (scenepaths/scene_paths.go) — the authoritative
		// resolver, per check-scene-path-resolution.sh; never hand-rolled.
		if topologyPath != "" {
			_ = WriteSceneSphere(scenepaths.SphereFilePath(topologyPath), ui.SceneSphere)
		}
	}
	// The scene sphere is established here and never moves again (MODEL.md), so the VIEW
	// frame emit below is the single source-of-truth broadcast the renderer uses in place
	// of deriving a content-sphere centroid from live node positions.
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): the gesture/stdin-reader goroutine
	// (this one — InstallSceneSphere runs before any other goroutine launches) also writes its
	// own VIEW frame directly, carrying this one-time scene-sphere event. SceneSphere
	// decodes entirely from the VIEW frame's own Scene block (buffer-log.ts's
	// decodeEventLine "scene-sphere" case) — no row identity to resolve.
	ui.EmitViewFrame([]wire.RowEvent{{Kind: T.KindSceneSphere, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
}
