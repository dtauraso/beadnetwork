package main

import (
	T "github.com/dtauraso/wirefold/Trace"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
)

// loadSceneState installs the saved camera / distance groups / overlays / speed / scene
// sphere, arming each persist writer AFTER its own seed. The ordering inside is the whole
// content of this phase — do not reorder it.
//
// Initial camera viewpoint = FILE DATA. Go reads the saved camera from
// <topologyPath>/view/camera.json itself and installs it into the gesture-FSM viewpoint,
// so the buffer camera columns carry a real, non-degenerate saved pose from the first
// frame (pan works immediately). Absent/malformed file → a fixed non-degenerate default.
//
// The buffer's node/edge/port row-identity tables now live ON md itself (built once at
// load, in newMoveDispatch's RowTables.Build call, from the same spec-order nodeSeeds/
// edgeSeeds each per-owner stream frame uses below) — a node/edge/port hit (which
// carries only a numeric buffer row index) resolves back to its identity via
// md.RT.LookupNodeRow/md.RT.LookupEdgeRow/md.LookupPortRow with no separate resolver wiring.
// Initial camera viewpoint = FILE DATA: Go reads the saved camera from
// <topologyPath>/view/camera.json and installs it into the gesture-FSM viewpoint.
func loadSceneState(scenePath string, md *W.MoveDispatch, tr *T.Trace, speedSinks []chan float64) {
	W.SeedInitialViewpoint(scenePath, md, tr)
	// Does THIS scene own the three named distance groups? Resolve before anything emits a
	// VIEW frame, since that frame carries the three group lengths. Not file data — it is a
	// property of which scene was loaded (Wiring.SceneTab.DistanceGroups).
	md.ResolveSceneDistanceGroups(scenePath)
	// Restore persisted overlay visibility: seed md.ov from overlays.json and emit each flag
	// so the buffer streams the saved overlay state from the first frame. Seed BEFORE
	// EnableEditPersist so the seed's own emit does not write the loaded state back.
	md.LoadOverlays(scenePath, tr)
	// Restore the persisted playback speed: seed md.ui.speed from view/speed.json, broadcast
	// it to every clock-owning goroutine's own channel (same Delivery path a live slider
	// edit uses), and emit it so the buffer reflects it from the first frame. Seed BEFORE
	// EnableEditPersist so the seed's own emit does not write the loaded/default speed back.
	md.LoadSpeed(scenePath, speedSinks, tr)
	// Arm the WRITE side AFTER the seeds: from here, every gesture that changes the FSM
	// viewpoint (orbit/zoom/pan/home) debounces a write of the current pose back to
	// <topologyPath>/view/camera.json, so navigate-then-reload round-trips.
	// Arming after the seed keeps the seed's own emit from persisting the loaded/default pose.
	md.EnableViewpointPersist(scenePath)
	// Arm disk persistence for the FSM-applied edits (node-drag position, ring-move
	// anchor) — debounced Go-side read-modify-writes, armed after the seeds so their
	// own emits do not write loaded state back.
	md.EnableEditPersist(scenePath)

	// Install the scene sphere (persisted, or a content-fit centroid for a fresh
	// scene) BEFORE launching the movers and the stdin reader. It only needs the
	// movers to be BUILT (their seeded centers, available since LoadTopology), not
	// running; installing it after Start left md.sceneSphere written unsynchronized
	// while the mover/gesture goroutines could already read it on the drag path.
	md.LoadSceneSphere(scenePath)
}
