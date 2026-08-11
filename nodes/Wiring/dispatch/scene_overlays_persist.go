// scene_overlays_persist.go — MoveDispatch-facing side of Go's OWN overlay-visibility
// persister (view/overlays.json): LoadOverlays. The persister itself is one instantiation
// of scenepersist.Persister[viewstate.OverlayState] (the shared debounce-then-write actor
// shape used by every scene-level file writer, see that type's own doc comment), constructed
// in move_persist.go's EnableEditPersist. Pure read/write helpers (WriteSceneOverlays/
// LoadSceneOverlays) live in nodes/Wiring/scenepersist/scene_overlays_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// the persister's Schedule() — both triggers (the bare `save` command and the on-change
// write) are dispatched from its own message loop. overlays.json is scene-level and
// genuinely singular, so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") rather than a per-entity split.
//
// LOAD side: LoadOverlays installs the persisted overlay state into md.ui.ov on startup +
// emits it so the first snapshot reflects the saved state — closing the
// toggle→reload→still-toggled round trip.
package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// LoadOverlays reads the overlay-visibility state from overlays.json (FILE DATA) into
// md.ui.ov and streams it so the buffer reflects the current overlay state from the first
// frame. A file with no overlay keys resolves to the code defaults (LoadSceneOverlays
// starts from viewstate.DefaultOverlayState and applies any present keys) — and those
// defaults are STILL emitted, so the UI shows the default-visible overlays instead of an
// all-off buffer. Call after LoadTopology (which builds MoveDispatch) and BEFORE
// EnableEditPersist so this emit does not write the loaded/default state back.
func (md *MoveDispatch) LoadOverlays(topologyPath string, tr *T.Trace) {
	ov, _ := scenepersist.LoadSceneOverlays(scenepaths.OverlaysFilePath(topologyPath)) // ov = defaults with any persisted keys applied
	md.UI.OV.SetGuideVisibility(ov)
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): the gesture/stdin-reader goroutine
	// (this one) writes its own VIEW frame directly, carrying the one-time overlay-flag
	// events this load implies — one RowEvent per flag kind. Every overlay kind decodes
	// entirely from the VIEW frame's own Overlay block (buffer-log.ts's decodeEventLine
	// OVERLAY_KINDS branch) — no row identity to resolve. tr is unused now (kept in the
	// signature to avoid rippling a call-site signature change through main.go).
	md.UI.EmitViewFrame([]wire.RowEvent{
		{Kind: T.KindSceneTori, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindScenePoles, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodePoles, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindSelSpherePoles, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindHandholds, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindLabelsGlobal, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindOverlaysVis, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodeBody, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodeRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindRingPick, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindSelectionRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindHoverRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindReachSphere, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
	})
}
