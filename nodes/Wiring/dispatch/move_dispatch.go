// move_dispatch.go — decentralized node-move handling.
//
// A node-move is NOT handled by a central coordinator. Instead the loader wires each
// mover's dedicated inbound channels: there is no shared many-to-one inbox — every pair of movers that talk (a node and its
// incident edge, or two nodes joined by an edge) gets its OWN directed channel, plus every
// mover gets one dedicated "external" channel for the stdin/gesture goroutine's rare direct
// entries. The stdin reader's whole job for a move is to write each entry onto the ONE
// channel that entry addresses. No recompute, no topology logic lives in the reader.
//
// Two kinds of mover own the work, each in its own goroutine (MODEL.md: each node
// and each wire is a goroutine; geometry emission is per-goroutine —
// memory/feedback_per_goroutine_bridge.md):
//
//   - nodeMover: owns ONE node's geometry. On a move for itself it updates its own
//     held position and re-emits its own node-geometry (emitNodeGeometry).
//   - edgeMover: owns ONE edge. It holds BOTH endpoint nodeGeoms (set at load). On
//     a move of either endpoint it updates the matching endpoint, recomputes its
//     OWN segment + arc (segmentBetweenPorts/arcLengthBetweenPorts), writes them
//     onto the source Out (next placement reads them), revises any in-flight bead's
//     remaining travel on the dest wire (ReviseInFlightGeometry, fraction-preserving),
//     and emits its OWN edge geometry (tr.Geometry).
//
// This reproduces, per-goroutine, exactly what the old central applyNodeMove did in
// one stdin goroutine: same node-geometry emit, same per-edge segment/arc recompute,
// same in-flight revision, same edge-geometry emit.
//
// This file holds the MoveDispatch composer struct itself and its nodeInboxes
// sub-owner; construction lives in move_dispatch_construct.go (newMoveDispatch) and the
// public delegator API lives in move_dispatch_api.go.

package dispatch

import (
	geomseeds "github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/layoutquant"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeinbox"
	rowtables "github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	sceneswitch "github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/streamwire"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/Trace"
)

// MoveDispatch is the pure registry built at load that owns every mover and wires their
// dedicated channels together — there
// is no shared dispatch map anymore; md.MR.NodeGeoms()/md.MR.EdgeMovers() themselves are the
// directories a mover's resolveDest closure and the external-entry helpers below look up.
type MoveDispatch struct {
	// MR owns the nodeMover/edgeMover directories (nodes/Wiring/moverreg, lifted out of this
	// package in docs/planning/movedispatch-decomposition.md §25). Bind/CenterOfNode/
	// EnqueueFor/FinalizeActors are called directly as md.MR.X by in-package callers; only
	// Start stays a MoveDispatch method. Exported (§28): its own
	// field surface is already reached through moverreg's own exported accessors
	// (NodeGeoms()/EdgeMovers()/...), so this is the same "already-exported sub-object"
	// shape as GS/UI/Scenes/RT below, applied to the one field that predated the pattern.
	MR moverreg.MoverRegistry
	// GS owns every node/edge's load-time seed geometry (nodes/Wiring/geomseeds). Exported:
	// external callers reach it directly (md.GS.NodeSeedsFn()) instead of through
	// MoveDispatch delegator methods.
	GS geomseeds.GeomSeeds
	// TR is the trace sink (retained for trace emission; diagnostic breadcrumbs removed).
	// Exported (§28): write-only from outside this file already (newMoveDispatch's own
	// struct literal); no in-package reader of md.TR exists, so exporting it costs nothing
	// and matches GS/UI/Scenes/RT's pattern rather than staying the one unexplained holdout.
	TR *T.Trace
	// Persist groups the six debounced disk persisters (nodes/Wiring/viewpersist, lifted
	// out of this package in docs/planning/movedispatch-decomposition.md §29), each nil
	// until armed by EnableViewpointPersist / EnableEditPersist after the startup seed.
	// Grouped the same way ui.vp/ui.ov/ui.gest are, so a bare test-constructed
	// MoveDispatch only has to reason about one zero-value sub-struct instead of six loose
	// nilable fields. Exported (§29): same already-exported-sub-object shape as MR/LQ
	// above — its own field surface is reached only through viewpersist's own exported
	// methods (ArmViewpoint/ArmEdit/Overlays()/Sphere()/Speed()/Lattice()).
	Persist viewpersist.Persisters
	// Sw owns the fd-wiring for the per-node interior stream (nodes/Wiring/streamwire,
	// lifted out of this package in docs/planning/movedispatch-decomposition.md §29); the
	// per-edge/per-node streams only — the dedicated VIEW stream's own fd/frame-builder/
	// tick now live on UI (nodes/Wiring/viewstate), lifted out per
	// docs/planning/gesture-actor.md. MoveDispatch's public SetEdgeStreams/SetNodeStreams
	// methods stay as thin delegators so the external API is unchanged. Exported (§29):
	// same already-exported-sub-object shape as MR/LQ above.
	Sw streamwire.StreamWiring
	// UI owns the camera/overlay/gesture/selection/abc-drag UI state AND the VIEW stream's
	// own write side (nodes/Wiring/viewstate — lifted out of this package per
	// docs/planning/gesture-actor.md; the earlier "uiState declined" probes,
	// docs/planning/movedispatch-decomposition.md sections 5/6b, were blocked specifically
	// by the VIEW emitter's direct field access, which moved here too). Exported: the
	// type's own surface, matching GS/RT/Scenes' existing pattern — MoveDispatch's
	// setSelectionUI/setHoverUI/sendEdgeSelect-shaped helpers still route through it via
	// bound func values (move_dispatch_api.go), since moverRegistry/layoutQuantizer/
	// edgeMover are unexported Wiring types viewstate cannot name.
	UI viewstate.UIState
	// LQ owns the quantized scene-polar move math (nodes/Wiring/layoutquant, lifted out
	// per docs/planning/movedispatch-decomposition.md §24 — quantizedLayout gates the
	// quantized absolute-scene-polar snap — every node is a root, measured/derived
	// about the scene center only, with no per-neighbour stored coordinate (MODEL.md
	// "the polar model"). Its methods take moverreg.MoverRegistry's own directories
	// (md.MR.NodeGeoms()/md.MR.EdgeMovers()) as explicit parameters rather than a
	// *moverreg.MoverRegistry back-reference, which is what let the type move to its own
	// package with nothing in MoverRegistry exported. Exported (§28): same
	// already-exported-sub-object shape as MR above.
	LQ layoutquant.LayoutQuantizer
	// Scenes owns tab switching: the anchor to persist the selection against, and the
	// quit func whose call the extension host's looping respawn follows (scene_switch.go).
	// Zero until the run's setup code arms it (Scenes.AnchorPath/Scenes.Quit, set directly
	// by runtopology/topology_run.go), so a bare test-constructed MoveDispatch can never
	// end a process. Exported: external callers reach it directly.
	Scenes sceneswitch.SceneSwitch

	// RT owns the three row-identity tables (nodes/Wiring/rowtables). Exported: external
	// callers (root package) reach it directly (md.RT.NodeRowFor(...)) instead of through
	// MoveDispatch delegator methods.
	RT rowtables.RowTables

	// Inboxes owns the directories of per-node dedicated channels a kind can claim for
	// itself at build time (nodes/Wiring/nodeinbox, lifted out of this package in
	// docs/planning/movedispatch-decomposition.md §29 — see its own package doc comment).
	// Exported (§29): same already-exported-sub-object shape as MR/LQ above.
	Inboxes nodeinbox.NodeInboxes
}
