// kind_events.go — the closed EVENT-KIND vocabulary. Every per-owner stream's RowEvent.Kind
// (nodes/Wiring's owner_events.go and friends) is one of the Kind* constants below;
// Buffer.KindID resolves it to its TraceEventKinds index for the wire encoding. Node is
// always the emitting node — the one that received the value (recv) or sent it
// (send/fire) — Port distinguishes input vs output where applicable.
//
// This is data, not logic: split out of Trace.go (which keeps the writer type, event.go
// the payload types, breadcrumb_labels.go the breadcrumb sub-vocabulary) so each file names
// one concern. tools/gen-node-defs/trace_kinds.go's parseTraceKinds already scans every
// non-test file under this directory (not a hardcoded filename), so this split does not
// blind that generator.
package Trace

const (
	KindRecv = "recv"
	KindFire = "fire"
	KindSend = "send"
	// KindEdgeBead is the per-frame bead-position kind (wire value "edge-bead", paired with
	// KindNodeBead below). The wire's delivery goroutine resolves one every ~16 ms while a
	// bead is in flight, carrying the bead's evaluated 3-D position so the renderer plots it
	// without computing geometry itself.
	KindEdgeBead = "edge-bead"
	// KindGeometry carries an edge's authoritative straight-segment endpoints. The
	// edgeMover resolves one per edge on load and again whenever a node-move re-derives
	// that edge's segment, so the renderer draws the wire tube from Go's endpoints and
	// computes no geometry of its own. Keyed by edge label (== the TS edge id).
	KindGeometry = "geometry"
	// KindNodeGeometry carries one node's authoritative center + per-port world
	// positions/directions. Each nodeMover resolves this once on startup and again on
	// every node-move — the node owns its own geometry (wires own bead-position).
	// Keyed by node id.
	KindNodeGeometry = "node-geometry"
	// KindArrive marks a bead COMPLETING its traversal on a wire — the bead has reached
	// the destination port and is delivered into the slot. The wire resolves it from
	// deliverLocked (the single delivery path), keyed by the bead's SOURCE node+port —
	// the same routing key as send/position — so the renderer clears the transit pulse
	// the instant the bead arrives.
	KindArrive = "arrive"
	// KindNodeBead carries one INTERIOR slot's authoritative grid-slot state (node 1's
	// depleting/refilling buffer). Node 1's Update computes the 2x2 grid slot positions
	// coupled to the working/backup array mutation and resolves a 4-slot SNAPSHOT (one
	// node-bead kind per slot) whenever the array changes. Keyed by node id +
	// (row,col): row 0 = top/backup, row 1 = bottom/working; col is the index within
	// that row. Payload = present (filled?) + value (0|1) + world position (x,y,z). A
	// popped slot carries present=false so TS clears it (absence can't be rendered,
	// presence can).
	KindNodeBead = "node-bead"
	// KindCamera carries the polar camera viewpoint state. Go resolves it whenever the
	// camera is set, orbited, zoomed, or panned, so the renderer can reconstruct the
	// camera pose without computing any geometry itself.
	KindCamera = "camera"
	// KindSceneTori carries the polar-guide tori visibility state.
	KindSceneTori = "scene-tori"
	// KindScenePoles carries the scene-center pole frame visibility state.
	KindScenePoles = "scene-poles"
	// KindNodePoles carries the per-node pole frame visibility state.
	KindNodePoles = "node-poles"
	// KindSelSpherePoles carries the selection-sphere pole axis visibility state.
	KindSelSpherePoles = "sel-sphere-poles"
	// KindHandholds carries the rotation-handhold grab-sphere visibility state.
	KindHandholds = "handholds"
	// KindLabelsGlobal carries the global node-label visibility state.
	KindLabelsGlobal = "labels-global"
	// KindOverlaysVis carries the master overlays visibility state.
	KindOverlaysVis = "overlays-vis"
	// KindNodeBody carries the node-sphere visibility state.
	KindNodeBody = "node-body"
	// KindNodeRing carries the per-node border-ring visibility state.
	KindNodeRing = "node-ring"
	// KindRingPick carries the ring click-band's visibility state — the band a click lands
	// on to author a port∈torus lock, painted so its position is visible. The band takes
	// clicks either way; this flag only says whether it is drawn.
	KindRingPick = "ring-pick"
	// KindSelectionRing carries the selected node's ring+halo visibility state.
	KindSelectionRing = "selection-ring"
	// KindHoverRing carries the hovered node's ring visibility state.
	KindHoverRing = "hover-ring"
	// KindReachSphere carries the selected node's reach-sphere ring visibility state.
	KindReachSphere = "reach-sphere"
	// KindSelect carries the CURRENTLY-SELECTED node id (click-select), or an edge label
	// on Edge with Node empty (edge selection — selection is single + exclusive across
	// nodes and edges). Node="" clears the selection (empty-space click).
	KindSelect = "select"
	// KindHover carries the CURRENTLY-HOVERED entity (pointer hover). Port!="" hovers
	// that port (Node is its owning node, Value=1 for an input port); otherwise Node
	// hovers that node (Node="" clears all hover).
	KindHover = "hover"
	// KindSceneSphere carries the persisted scene sphere (center + radius) — the fixed
	// world anchor every node's scene polar is measured about. Established ONCE at load
	// and never moves.
	KindSceneSphere = "scene-sphere"
	// KindAbcDrag marks one time-node (Time) abc-drag re-quantize event — the
	// routed counterpart to the "time.abc-drag" debug breadcrumb emitted alongside it
	// (nodes/Wiring/quantized_move.go neighborSetCRequantize).
	KindAbcDrag = "abc-drag"
	// KindAbcDragReset marks the START of one drag operation — resolved exactly once at
	// the gesture FSM's pending→dragging transition, BEFORE the dragged node's
	// neighborSetC fan resolves any KindAbcDrag marks for that drag.
	KindAbcDragReset = "abc-drag-reset"
	// KindBreadcrumb carries a DEBUG BREADCRUMB (see .claude/rules/go-debugging.md's
	// "Debugging the Go layer (probe breadcrumbs)" section) as a structured buffer
	// EVENT row instead of a free-form JSON
	// stdout line. It rides the EMITTING goroutine's own per-owner stream (node/edge/
	// interior/VIEW) — main.go's own breadcrumbs (no per-node stream) ride the VIEW
	// stream. Label (a BreadcrumbLabel* index, breadcrumb_labels.go) names which of the
	// breadcrumb sites emitted it; the row's other columns (Value/X/Y/Z/NodeRow/PortRow/
	// TargetRow/TargetPortRow) are REUSED per label, with Label/TextOff/TextLen (the
	// bufLayoutEvent.Debug flag is always 1 on this Kind) as the two dedicated
	// breadcrumb-only columns.
	KindBreadcrumb = "breadcrumb"
)

// TraceEventKinds is the single source of truth for the closed kind vocabulary.
// gen-node-defs reads this slice to emit trace-kinds.ts (the TS decode side's kindId →
// name lookup), and Buffer.KindID indexes it to resolve a RowEvent's string Kind to its
// numeric id for the wire encoding. There is no tsc exhaustiveness check derived from
// it — adding a kind here does not force a TS branch anywhere; it only extends the
// lookup table.
var TraceEventKinds = []string{KindRecv, KindFire, KindSend, KindEdgeBead, KindGeometry, KindNodeGeometry, KindArrive, KindNodeBead, KindCamera, KindSceneTori, KindScenePoles, KindNodePoles, KindSelSpherePoles, KindHandholds, KindLabelsGlobal, KindOverlaysVis, KindNodeBody, KindNodeRing, KindRingPick, KindSelectionRing, KindHoverRing, KindReachSphere, KindSelect, KindHover, KindSceneSphere, KindAbcDrag, KindAbcDragReset, KindBreadcrumb}
