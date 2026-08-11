// Package viewpersist is the view-owner goroutine's (RunStdinReader, stdin_reader.go) OWN
// state for the three SCENE-LEVEL files it writes — camera.json/overlays.json/sphere.json,
// each genuinely singular (there is only one camera pose, one overlay-flag set, one scene
// sphere), so each stays one file with this one goroutine as its named owner
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path"), rather
// than a per-entity split the way node files are.
//
// This is a NEW package, not folded into nodes/Wiring/scenepersist (which already owns the
// generic Persister[T] mechanism and the WriteScene*/LoadScene* functions), because
// Persisters is MoveDispatch's own named grouping of six ALREADY-lifted persister
// instances (camerapersist.ViewpointPersister plus five scenepersist.Persister[T]
// specializations) — folding the grouping into scenepersist would make the generic
// mechanism package import camerapersist/viewstate/geom/MoveDispatch-specific types it has
// no other reason to know about.
//
// It is NOT a shared bag other goroutines reach into: a Persisters value's fields are
// read/written exclusively from methods the view-owner goroutine calls (EmitViewpoint,
// applyUpdate, LoadSceneSphere/handleSaveMsg — see each field's own comment, still on
// MoveDispatch). Node-drag position/local-polars and port-anchor edits are NOT here —
// those are persisted by each node's OWN mover (nm.persistRoot, quant_offset_persist.go /
// scene_anchor_persist.go), not by the view-owner goroutine.
//
// Each field is nil until armed by ArmViewpoint / ArmEdit (called from
// MoveDispatch.EnableViewpointPersist / EnableEditPersist after the startup seed). Each
// persister writes synchronously the moment its value changes (see scene_persist.go's
// header comment for why the prior debounce was removed) — there is no
// pending-value/clean-shutdown-flush machinery to maintain.
package viewpersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// Persisters holds the view-owner goroutine's six debounced-write persisters — see the
// package doc comment.
type Persisters struct {
	// vp is the camera-viewpoint persister (nodes/Wiring/camerapersist), armed by
	// ArmViewpoint after the startup seed. nil until armed (old path / tests). Write-only
	// from outside this package (ArmViewpoint returns the constructed value directly to its
	// caller, which needs it to wire viewstate.UIState.VP.Persist = p.Schedule) — there is
	// no in-package reader either, matching MoveDispatch.TR's own write-only precedent
	// (§28).
	vp *camerapersist.ViewpointPersister
	// overlays is the overlay-flags persister, armed by ArmEdit. nil until armed (tests
	// that never arm). Each of overlays/sphere/speed/lattice is one instantiation of the
	// shared scenepersist.Persister[T] actor (persister.go) — same shape, different
	// payload type and bound write func.
	overlays *scenepersist.Persister[viewstate.OverlayState]
	// sphere is the disk persister for the scene sphere (md.UI.SceneSphere), armed by
	// ArmEdit. It is only ever flushed — by LoadSceneSphere on a content-fit, and by
	// HandleSaveMsg — never scheduled on a value-change, because the sphere is
	// "established once and never moves" (MODEL.md). nil until armed (tests that never
	// arm).
	sphere *scenepersist.Persister[geom.SceneSphere]
	// speed is the playback-speed persister, armed by ArmEdit. nil until armed (tests that
	// never arm).
	speed *scenepersist.Persister[float64]
	// lattice is the lattice-point-count persister, armed by ArmEdit. nil until armed
	// (tests that never arm).
	lattice *scenepersist.Persister[int32]
}

// ArmViewpoint constructs and stores the camera-viewpoint persister rooted at
// topologyPath, and returns it so the caller (MoveDispatch.EnableViewpointPersist) can
// wire its Schedule func onto viewstate.UIState.VP.Persist — the one read of vp anywhere,
// done once at construction, which is why vp itself has no accessor.
func (p *Persisters) ArmViewpoint(topologyPath string) *camerapersist.ViewpointPersister {
	vp := &camerapersist.ViewpointPersister{Path: scenepaths.CameraFilePath(topologyPath)}
	p.vp = vp
	return vp
}

// ArmEdit constructs and stores the four FSM-edit persisters rooted at topologyPath —
// overlays/sphere/speed/lattice — see MoveDispatch.EnableEditPersist's own doc comment for
// the full list of what each persists and why topologyPath is always the tree root.
func (p *Persisters) ArmEdit(topologyPath string) {
	p.overlays = &scenepersist.Persister[viewstate.OverlayState]{
		Path: scenepaths.OverlaysFilePath(topologyPath), Write: scenepersist.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	p.sphere = &scenepersist.Persister[geom.SceneSphere]{
		Path: scenepaths.SphereFilePath(topologyPath), Write: scenepersist.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	p.speed = &scenepersist.Persister[float64]{
		Path: scenepaths.SpeedFilePath(topologyPath), Write: scenepersist.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	p.lattice = &scenepersist.Persister[int32]{
		Path: scenepaths.LatticeFilePath(topologyPath), Write: scenepersist.WriteSceneLattice, Tag: "scene_lattice_persist",
	}
}

// Overlays returns the overlay-flags persister (nil until ArmEdit runs — Schedule is a
// nil-receiver-safe no-op, per scenepersist.Persister[T]'s own doc comment).
func (p *Persisters) Overlays() *scenepersist.Persister[viewstate.OverlayState] { return p.overlays }

// Sphere returns the scene-sphere persister (nil until ArmEdit runs).
func (p *Persisters) Sphere() *scenepersist.Persister[geom.SceneSphere] { return p.sphere }

// Speed returns the playback-speed persister (nil until ArmEdit runs).
func (p *Persisters) Speed() *scenepersist.Persister[float64] { return p.speed }

// Lattice returns the lattice-point-count persister (nil until ArmEdit runs).
func (p *Persisters) Lattice() *scenepersist.Persister[int32] { return p.lattice }
