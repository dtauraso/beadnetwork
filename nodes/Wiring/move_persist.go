package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// persisters is the view-owner goroutine's (RunStdinReader, stdin_reader.go) OWN state for
// the three SCENE-LEVEL files it writes — camera.json/overlays.json/sphere.json, each
// genuinely singular (there is only one camera pose, one overlay-flag set, one scene
// sphere), so each stays one file with this one goroutine as its named owner
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path"), rather than a per-entity split
// the way node files are. It is NOT a shared bag other goroutines reach into: md.persist's
// three fields are read/written exclusively from methods this same view-owner goroutine
// calls (EmitViewpoint, applyUpdate, LoadSceneSphere/handleSaveMsg — see each field's own
// comment). Node-drag position/local-polars and port-anchor edits are NOT here — those are
// persisted by each node's OWN mover (nm.persistRoot, quant_offset_persist.go /
// scene_anchor_persist.go), not by the view-owner goroutine.
//
// Each nil until armed by EnableViewpointPersist / EnableEditPersist after the startup
// seed. Each persister writes synchronously the moment its value changes (see
// scene_persist.go's header comment for why the prior debounce was removed) — there is no
// pending-value/clean-shutdown-flush machinery to maintain.
type persisters struct {
	// vp is the camera-viewpoint persister (nodes/Wiring/camerapersist), armed by
	// EnableViewpointPersist after the startup seed. nil until armed (old path / tests).
	vp *camerapersist.ViewpointPersister
	// overlays is the overlay-flags persister (scene_overlays_persist.go), armed by
	// EnableEditPersist after the startup seed. nil until armed (tests that never arm). Each
	// of overlays/sphere/speed/lattice is one instantiation of the shared
	// scenepersist.Persister[T] actor (persister.go) — same shape, different payload type
	// and bound write func.
	overlays *scenepersist.Persister[viewstate.OverlayState]
	// sphere is the disk persister for the scene sphere (sphere_layout.go md.ui.sceneSphere),
	// armed by EnableEditPersist. It is only ever flushed — by LoadSceneSphere on a
	// content-fit, and by handleSaveMsg — never scheduled on a value-change, because the
	// sphere is "established once and never moves" (MODEL.md). nil until armed (tests that
	// never arm).
	sphere *scenepersist.Persister[geom.SceneSphere]
	// speed is the playback-speed persister (scene_speed_persist.go), armed by
	// EnableEditPersist. nil until armed (tests that never arm).
	speed *scenepersist.Persister[float64]
	// lattice is the lattice-point-count persister (scene_lattice_persist.go), armed by
	// EnableEditPersist. nil until armed (tests that never arm).
	lattice *scenepersist.Persister[int32]
}

// EnableViewpointPersist arms gesture-driven camera persistence: every subsequent
// EmitViewpoint (orbit/zoom/pan/home) writes the current viewpoint to
// `<topologyPath>/view/camera.json` (nodes/Wiring/camerapersist). Call AFTER
// SeedInitialViewpoint so the seed's own emit does not write the loaded/default pose back.
// Go owns this write (MODEL.md); the old path persists the camera via its own TS scene-save.
func (md *MoveDispatch) EnableViewpointPersist(topologyPath string) {
	p := &camerapersist.ViewpointPersister{Path: scenepaths.CameraFilePath(topologyPath)}
	md.persist.vp = p
	md.UI.VP.Persist = p.Schedule
}

// EnableEditPersist arms disk persistence for the FSM-applied topology edits:
//   - node-drag (RootMove) → the moved node's own position.json + local-polars.json,
//     written by that node's OWN mover (nm.persistRoot, set below on every mover — see
//     quant_offset_persist.go's persistQuantOffset/persistLocalPolars)
//   - ring-move (applyRingAnchor) → the moved port's own node's OWN mover writes the
//     port's anchorId back to the port json file (scene_anchor_persist.go's
//     persistPortAnchor), same nm.persistRoot as above
//   - overlays (applyUpdate toggle/set) → overlay-visibility keys in view/overlays.json
//   - clock speed (applyUpdate clock/speed) → view/speed.json (scene_speed_persist.go)
//   - pair lattice point count (applyUpdate scene/latticePoints) → view/lattice.json
//     (scene_lattice_persist.go)
//
// topologyPath is always the tree root directory — LoadTopology rejects anything else
// (topo_spec.go) — so it is used directly as root for the per-node/per-port persisters.
// Call AFTER SeedInitialViewpoint so the seed emits do not write the loaded state back.
func (md *MoveDispatch) EnableEditPersist(topologyPath string) {
	root := topologyPath
	// The LOADED scene's own root, kept so a structural edit (scene_structure.go's node
	// create/delete) can write into the tree that is actually showing. Every other persister
	// here already closes over a path derived from it; this is the one operation that needs
	// the root itself, because it creates and removes whole node directories rather than
	// rewriting one known file.
	md.Scenes.TreeRoot = root
	md.persist.overlays = &scenepersist.Persister[viewstate.OverlayState]{
		Path: scenepaths.OverlaysFilePath(topologyPath), Write: scenepersist.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	md.persist.sphere = &scenepersist.Persister[geom.SceneSphere]{
		Path: scenepaths.SphereFilePath(topologyPath), Write: scenepersist.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	md.persist.speed = &scenepersist.Persister[float64]{
		Path: scenepaths.SpeedFilePath(topologyPath), Write: scenepersist.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	md.persist.lattice = &scenepersist.Persister[int32]{
		Path: scenepaths.LatticeFilePath(topologyPath), Write: scenepersist.WriteSceneLattice, Tag: "scene_lattice_persist",
	}
	// Every node's own mover writes its own position.json/local-polars.json/port-anchor
	// files — set the tree root on each nodeMover directly rather than routing writes
	// through a shared MoveDispatch-owned persister (docs/planning/decentralized-
	// persistence.md "The model"). A plain field write on each mover, done here before
	// any mover goroutine starts (Start runs after EnableEditPersist in every real call
	// path), so no synchronization is needed.
	for _, nm := range md.mr.nodeGeoms {
		nm.setPersistRoot(root)
	}
}
