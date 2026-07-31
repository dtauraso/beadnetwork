// Buffer/layout.go — single source of truth for the agnostic content-buffer
// column layout schema.
//
// tools/gen-node-defs reads this file and emits:
//   - Buffer/buffer_layout_gen.go  (Go offset constants + typed writer helpers)
//   - tools/topology-vscode/src/schema/buffer-layout.ts  (TS constants + DataView readers)
//
// Regenerate with: cd tools/topology-vscode && npm run gen:node-defs
//
// Field tags: buf:"type" where type is one of f32 | i32 | u32 | u8.
// Offsets and strides are computed by the generator in field-declaration order
// (packed; no implicit padding — DataView handles unaligned reads on both sides).
// Struct names beginning with "bufLayout" are recognised by the generator as
// column-block definitions; the suffix becomes the block name (e.g. bufLayoutBead
// → block "Bead").
//
// BUF_LAYOUT_VERSION is bumped whenever any column definition changes; the
// generated files carry the same version so a stale regeneration is immediately
// visible.

package Buffer

// BufLayoutVersion is the schema version. Bump when any column changes.
const BufLayoutVersion = 36

// BufInteriorSlotsPerNode is the fixed number of interior grid slots reserved per
// node in the Interior block (a 2x2 held/interior-bead grid: slot = row*2 + col).
// The Interior block carries exactly nodeCount*BufInteriorSlotsPerNode rows in
// stable node order, so it needs no separate count in the header — the decoder
// derives its length from nodeCount. Not a per-column generated field (there is no
// bufLayoutInterior column for it), but gen-node-defs DOES read this const directly
// (parseBufferLayout) and emits it as generated TS (INTERIOR_SLOTS_PER_NODE in
// buffer-layout.ts) and Go (BufInteriorSlotsPerNodeGenerated) constants, folded into
// BUF_LAYOUT_FINGERPRINT — so a drift here fails check-buffer-layout-parity.sh, not
// just a same-symbolic-constant-on-both-sides test that could never catch a value change.
const BufInteriorSlotsPerNode = 4

// NOTE: this file used to also declare a "semantic event enum" (BufEventRecv,
// BufEventFire, BufEventSend, BufEventArrive, BufEventDone), generated into
// BufEvent*ID (Go) / BUF_EVENT_* (TS) constants. It was deleted (not
// corrected): it was STALE and UNUSED everywhere except its own generated
// files and a tautological test asserting the constants equal themselves.
// The REAL per-tick event kind byte written into the buffer (EVENT block Kind
// column, and the Node block's transient per-tick flags) is the INDEX into
// T.TraceEventKinds (Buffer/stream_events.go buildKindIDMap: Recv=0, Fire=1,
// Send=2, EdgeBead=3, Geometry=4, NodeGeometry=5, Arrive=6, NodeBead=7, …) —
// an entirely different, already-correct numbering that has nothing to do
// with the deleted enum. Do not reintroduce a parallel BufEvent* enum; if a
// column ever needs a fixed-kind lookup, it should read T.TraceEventKinds'
// index directly, the way production code already does.

// --- Column block schemas -----------------------------------------------
// Each struct defines one column block. Fields are packed in declaration order.
// The generator computes byte offsets and stride from buf: tags.

// The Bead block is GONE. It was one row per live in-flight bead (world X/Y/Z + Value),
// repacked every tick as the bead moved. Nothing draws a moving bead any more: a traversal
// is rendered as the LIT bead of a node-owned fixed chain (see bufLayoutChainBead's Lit
// column and docs/beads-are-the-edge.md), so there is no per-tick position to carry.
//
// bufLayoutNode defines one row of the nodes column block.
// One row per node. Persistent geometry. (Recv/Fire/Send/Arrive/Done events are
// carried by the EVENT block only — see bufLayoutEvent — not by per-node columns.)
type bufLayoutNode struct {
	CX      float32 `buf:"f32"` // node center x (world)
	CY      float32 `buf:"f32"` // node center y (world)
	CZ      float32 `buf:"f32"` // node center z (world)
	Radius  float32 `buf:"f32"` // body/ring sphere radius
	SphereR float32 `buf:"f32"` // sphere-chain reach radius
	// VR*/FR* are the node's two great-circle ring-plane normals (vertical vr, flat fr),
	// the same orientation vectors the pre-branch SphereRing read from nodeGeometryStore.
	// SphereRing draws two tori at the owner's center oriented by these; they arrive on the
	// node-geometry trace event (Trace VRX.., FRX..).
	VRX      float32 `buf:"f32"` // vertical ring-plane normal x
	VRY      float32 `buf:"f32"` // vertical ring-plane normal y
	VRZ      float32 `buf:"f32"` // vertical ring-plane normal z
	FRX      float32 `buf:"f32"` // flat (equatorial) ring-plane normal x
	FRY      float32 `buf:"f32"` // flat ring-plane normal y
	FRZ      float32 `buf:"f32"` // flat ring-plane normal z
	Selected uint8   `buf:"u8"`  // persistent: 1 = this node is the click-selected node
	// KindId is the node's kind as a STABLE id, assigned once per kind in
	// nodes/<Kind>/SPEC.md (| kindId | N |) and never renumbered (the generator emits
	// kindIDMap/NODE_DEFS_ARRAY keyed by it; a removed kind leaves an undefined gap, not
	// a shift). Populated once on first KindNodeGeometry; 0xFF = unknown kind.
	KindId uint8 `buf:"u8"` // kind index into NODE_DEFS_ARRAY; 0xFF = unknown
	// LabelOff/LabelLen are this node's slice into the snapshot's trailing LABEL BYTES
	// section: LabelOff is the byte offset into that section, LabelLen the UTF-8 byte
	// length. The label bytes for all nodes are concatenated in node-row order (self-
	// sizing via the header labelBytesCount field, like portCount), so the numeric node
	// row carries its human label with no sidecar: the renderer slices
	// labelBytes[LabelOff : LabelOff+LabelLen) and TextDecoder-decodes it. LabelLen=0 = no
	// label (fall back to nothing / row index on the render side).
	LabelOff uint32 `buf:"u32"` // byte offset into the label-bytes section
	LabelLen uint32 `buf:"u32"` // label UTF-8 byte length
	// Hovered is the Go-owned pointer-hover flag: 1 marks the node currently under the
	// pointer (the gesture FSM tracks it from the raycast hit on each pointer-move and
	// emits KindHover). The renderer thickens+recolors this node's border ring (pre-branch
	// hover style: #aaddff, r*0.14). Persistent-until-next-move; NOT a transient event flag.
	// Selection styling takes precedence over hover where both apply (renderer-side).
	Hovered uint8 `buf:"u8"` // 1 = node is pointer-hovered
	// LatchedSel is Go-owned: 1 marks the LAST node that was click-selected, and stays 1
	// through a deselect (clicking empty space clears Selected but NOT LatchedSel; selecting
	// a DIFFERENT node moves LatchedSel to it). Set alongside Selected by the affected
	// node's own nodeMover (nodes/Wiring/node_move.go). Replaces the old TS-owned
	// `latchedSel` React state in NavGuides.tsx (that was a second, TS-invented selection
	// concept unreachable from Go); the render path now just reads this column.
	LatchedSel uint8 `buf:"u8"` // 1 = this is the last-selected node (persists through deselect)
	// GotDragMsg is Go-owned and DRAG-SCOPED: 1 marks a node that has received a
	// time.abc-drag message during the CURRENT drag (set on the recipient's own
	// nodeMover goroutine by quantized_move.go's neighborSetCRequantize, and cleared
	// at the start of each new drag via a moveMsgKindAbcReset broadcast). This is the recipient SET
	// the AbcDragLabel overlay lists by name — it replaces the old single-recipient
	// Overlay.LastAbcDragNodeRow column.
	GotDragMsg uint8 `buf:"u8"` // 1 = this node has received at least one time.abc-drag message
	// DragDeltaA/B/C carry the DRAGGED node's OWN quantized-triple change (newTriple -
	// oldTriple, integer indices) that THIS node received on the CURRENT drag's
	// neighborSetC message (see Trace.Event.DeltaA/B/C, nodes/Wiring/node_move.go
	// requantizeLocalPolars). DRAG-SCOPED like GotDragMsg: set from KindAbcDrag's
	// Event.DeltaA/B/C, cleared to 0 on KindAbcDragReset. Zero for a node that has not
	// (yet, this drag) received a message, and legitimately zero for a node whose
	// delta happened to be (0,0,0) — GotDragMsg is what distinguishes the two.
	DragDeltaA int32 `buf:"i32"` // dragged node's own theta-index delta, this drag
	DragDeltaB int32 `buf:"i32"` // dragged node's own phi-index delta, this drag
	DragDeltaC int32 `buf:"i32"` // dragged node's own r-index delta, this drag
	// DragRequantCount is Go-owned and DRAG-SCOPED, mirroring GotDragMsg/DragDeltaA-C
	// exactly: a per-RECIPIENT cumulative count of time.abc-drag re-quantize messages
	// THIS node has received during the CURRENT drag (incremented on the recipient's
	// own nodeMover goroutine by quantized_move.go's neighborSetCRequantize, alongside
	// GotDragMsg, and cleared to 0 at the start of each new drag via a
	// moveMsgKindAbcReset broadcast). Replaces the old central Overlay.AbcDragCount,
	// which summed ticks on a cross-goroutine channel that could drop them under
	// pointer-input load; this is state on the node's own reliable stream, so nothing
	// can drop it. The editor's "drag received ×N" total sums this column across all
	// node rows (TS-side; see overlay-flags.ts readDragReceivedCount).
	DragRequantCount int32 `buf:"i32"` // this node's own cumulative recipient count, this drag
	// GotForwardMsg/ForwardDeltaA-C/ForwardFromRow mirror GotDragMsg/DragDeltaA-C exactly,
	// but for the delta-forward full-graph-propagation observability feature
	// (nodes/Wiring/node_mover.go's forwardDeltaOnce): every node that picks up a delta
	// triple — as the direct drag-recipient (neighborSetCRequantize) or as a forward
	// recipient (moveMsgKindDeltaForward) — records it here on the FIRST delta it sees
	// this drag, then relays the SAME triple onward to its own OTHER neighbors exactly
	// once (forwardedThisDrag guards it — a later delta reaching an already-forwarded
	// node does not update these columns again). DRAG-SCOPED like GotDragMsg: cleared to
	// 0/-1 in the same moveMsgKindAbcReset handler (nodes/Wiring/node_mover.go) that
	// resets GotDragMsg/DragDeltaA-C.
	GotForwardMsg uint8 `buf:"u8"`  // 1 = this node has received a delta-forward message, this drag
	ForwardDeltaA int32 `buf:"i32"` // forwarded theta-index delta (the ORIGINAL dragged node's own delta)
	ForwardDeltaB int32 `buf:"i32"` // forwarded phi-index delta
	ForwardDeltaC int32 `buf:"i32"` // forwarded r-index delta
	// ForwardFromRow is the buffer ROW index of whichever neighbor's own hop reached this
	// node FIRST this drag, -1 when none (reset state). Resolved via
	// MoveDispatch.NodeRowFor at the forward-recipient's handler.
	ForwardFromRow int32 `buf:"i32"` // forwarder's buffer node row, -1 = none
	// CascadeRelay is Go-owned and STATIC (a pure function of this node's own kind, set
	// once at construction like KindId — never drag-scoped): which branch of
	// nodes/Wiring/node_mover.go's forwardDelta this node's kind takes when it picks up a
	// delta triple. 0 = FAN (relay to every cascade neighbor except the sender),
	// 1 = ROUTED (relay to a single target kind chosen by the SENDER's kind, or drop —
	// Pulse and TimeStart), 2 = TERMINUS (never relays onward — TimeEnd, PulseLeft,
	// PulseRight). FAN, not "flood": every one of these three is bounded, and "flood" is
	// reserved for the unbounded run the per-kind rules exist to stop (forwardDelta's doc). Read by the editor's drag log (AbcDragLabel.tsx) to name the DRAGGED
	// node's relay behavior beside its name. The classification lives in Go beside the
	// rules it summarizes (cascadeRelayClass) precisely so TS does not re-derive it from
	// KindId and drift when a rule changes.
	CascadeRelay uint8 `buf:"u8"` // 0 = fan, 1 = routed, 2 = terminus
}

// bufLayoutChainBead defines one row of the chain-bead column block — the node-owned
// placeholder sequence that IS the visual of an edge (docs/beads-are-the-edge.md). One row
// per placeholder bead on one of this node's OUTGOING edges, in that edge's order outward
// from this node.
//
// OX/OY/OZ are NODE-LOCAL, exactly like the Interior block's: the offset from this node's
// own center, with the renderer adding the center to get the world position. That is not the
// renderer owning positions (Go owns the offsets); it is what makes moving a node constant
// time, because the offsets do not change when the node's center does.
//
// A chain bead has no absolute position column ON PURPOSE. The old moving bead (bufLayoutBead
// below) carried recomputed world X/Y/Z every tick; a chain does not move, so there is
// nothing to recompute.
//
// NOTE: no bead position here depends on another bead's position. That is the line separating
// this from the reverted bead-chain wire (memory/project_wire_is_straight_line_not_chain.md),
// whose spacing came from neighbour midpoints and therefore followed a drag in O(N²). Each
// offset is index × spacing along this node's own aim at the target — dependency depth 1.
type bufLayoutChainBead struct {
	OX float32 `buf:"f32"` // node-local offset x from this node's center
	OY float32 `buf:"f32"` // node-local offset y
	OZ float32 `buf:"f32"` // node-local offset z
	// Lit is the ANIMATION: 1 marks the bead a traversal has currently reached on this
	// chain, 0 every other bead. This replaces a bead MOVING along a wire — the chain is
	// fixed and the lighting is what advances (docs/beads-are-the-edge.md).
	//
	// Go owns it: the source node drives its own outgoing wires and reads their in-flight
	// fractional progress t on its own goroutine, then lights index = t × count. The
	// renderer only colours what this column says.
	Lit uint8 `buf:"u8"` // 1 = a traversal has reached this bead
	// LitValue is the VALUE (0|1) of the traversal occupying this bead, meaningful only when
	// Lit==1. The lit bead is drawn in bead 0's or bead 1's own fill, so the renderer needs
	// the value, not just the fact of being lit — the whole animation is ONE fill-colour
	// change against the grey resting chain.
	LitValue int32 `buf:"i32"` // traversing bead's value (0|1); meaningful when Lit==1
	// Radius is this bead's own SPHERE radius (not the outer/ring radius). Beads on
	// different edges differ: each edge picks the one size that makes its own beads touch
	// their neighbours exactly on a STRAIGHT chain (nodes/Wiring/chain_beads.go), so this
	// can no longer be a fixed constant shared by every bead the way BeadRadius once was.
	Radius float32 `buf:"f32"` // per-edge bead sphere radius
}

// bufLayoutInterior defines one row of the interior-bead column block.
// The block carries a FIXED BufInteriorSlotsPerNode (4) rows per node, in stable
// node order: row = nodeRow*BufInteriorSlotsPerNode + slot, slot = gridRow*2 + gridCol.
// Matched from KindNodeBead trace events (node's 2x2 held/interior grid). OX/OY/OZ
// are the Go-owned NODE-LOCAL slot offset (relative to the node center — the renderer
// adds the node center to get the world position); Present=0 hides the slot even when
// Value/offset are present (a popped/empty slot is streamed explicitly so it clears).
type bufLayoutInterior struct {
	Present uint8   `buf:"u8"`  // 1 = slot filled (draw); 0 = empty (hide)
	Value   int32   `buf:"i32"` // bead value (0|1); colored via bead-style
	OX      float32 `buf:"f32"` // node-local slot offset x
	OY      float32 `buf:"f32"` // node-local slot offset y
	OZ      float32 `buf:"f32"` // node-local slot offset z
}

// bufLayoutEdge defines one row of the edges column block.
// One row per edge (wire). Matched from KindGeometry trace events.
//
// SX..EZ is the edge's own straight SEGMENT (docs/bead-lattice.md "Ownership": the
// edgeMover publishes the segment only) — NODE SURFACE TO NODE SURFACE along the
// centre-to-centre line (edgeSegment, nodes/Wiring/port_geometry.go), the same two
// points chain_beads.go anchors bead 0 and the last bead to. There is no port row to
// reference any more: a port stopped being a place (docs/channels-not-ports.md), so this
// column pair is the edge's own emitted endpoints, not an index into a Port block that no
// longer exists. This DOES reintroduce a second copy of a world position (the node's own
// center + torus radius, computed once here rather than read live) — the prior port-row
// indirection existed specifically to dodge that tear, but the tear it was dodging was
// itself the bug this rewrite removes (a port's PLACE floating apart from the node's own
// surface): the edgeMover recomputes this segment on every endpoint move
// (recomputeGeometry), same as it always has, so it is never more than one move-event
// stale — no different from the Node block's own center staying fresh.
type bufLayoutEdge struct {
	SX       float32 `buf:"f32"` // segment start x (source node surface, world)
	SY       float32 `buf:"f32"` // segment start y
	SZ       float32 `buf:"f32"` // segment start z
	EX       float32 `buf:"f32"` // segment end x (target node surface, world)
	EY       float32 `buf:"f32"` // segment end y
	EZ       float32 `buf:"f32"` // segment end z
	Selected uint8   `buf:"u8"`  // persistent: 1 = this edge is the click-selected edge
	// EdgeLabelOff/EdgeLabelLen are this edge's slice into the snapshot's trailing EDGE-LABEL
	// BYTES section (the label-section analogue for edges): EdgeLabelOff is the byte offset,
	// EdgeLabelLen the UTF-8 byte length. Edge labels are carried ONLY for the .probe buffer-
	// decoded log (geometry `edge`, select-edge) — the render/bridge path
	// still resolves an edge hit by row index (LookupEdgeRow), never by this string.
	// Concatenated in the same stable edge-row order as the Edge block.
	EdgeLabelOff uint32 `buf:"u32"` // byte offset into the edge-label-bytes section
	EdgeLabelLen uint32 `buf:"u32"` // edge-label UTF-8 byte length
}

// bufLayoutLayoutLink defines one row of the LAYOUT-link column block: the cascade-link
// relationship (nodes/<id>/cascade-edges.json, a STORED per-node list — see
// nodes/Wiring/node_mover.go's cascadeEdges doc comment), NOT the bead-edge graph and NOT
// derived from LocalPolars/domain-edge adjacency. The PAIR is streamed once at load
// (deduplicated: only the lexicographically-smaller endpoint emits it).
//
// SrcNodeRow/DstNodeRow are the buffer NODE-ROW indices (resolved against the Node block —
// see nodeRowIndex). The overlay is its OWN edge between the two NODES' CENTERS
// (CX/CY/CZ, re-streamed on every move) — it does NOT reference any bead-edge row, so it
// can never be coupled to (or dimmed/tinted by) the bead edge's own selection/opacity
// state. No EdgeRow column.
type bufLayoutLayoutLink struct {
	SrcNodeRow int32 `buf:"i32"` // one endpoint's buffer node-row index (lexicographically smaller)
	DstNodeRow int32 `buf:"i32"` // the other endpoint's buffer node-row index
}

// The Port block is GONE (docs/channels-not-ports.md): a port is a load-time
// channel-binding ROLE (PortSpec, a.In()/a.Out()), never a place, so it has no ring
// anchor, no world position, and no buffer row of its own any more. An edge's
// endpoints now ride the Edge block's own SX..EZ (bufLayoutEdge, above); hover/select
// address the NODE, not a per-port row.

// bufLayoutCamera defines the camera column block (always 1 row).
// Matched from KindCamera trace events.
type bufLayoutCamera struct {
	PX       float32 `buf:"f32"` // pivot world x
	PY       float32 `buf:"f32"` // pivot world y
	PZ       float32 `buf:"f32"` // pivot world z
	R        float32 `buf:"f32"` // orbit radius
	PosTheta float32 `buf:"f32"` // pivot→camera polar θ
	PosPhi   float32 `buf:"f32"` // pivot→camera polar φ
	UpTheta  float32 `buf:"f32"` // up-hint polar θ
	UpPhi    float32 `buf:"f32"` // up-hint polar φ
}

// bufLayoutOverlay defines the overlay visibility column block (always 1 row).
// Matched from KindSceneTori/ScenePoles/…/OverlaysVis trace events.
// Field order matches overlayState in overlay_gen.go.
type bufLayoutOverlay struct {
	SceneTori      uint8 `buf:"u8"` // 1 = polar-guide tori visible
	ScenePoles     uint8 `buf:"u8"` // 1 = scene-center pole frame visible
	NodePoles      uint8 `buf:"u8"` // 1 = per-node pole frames visible
	SelSpherePoles uint8 `buf:"u8"` // 1 = selection-sphere pole axes visible
	Handholds      uint8 `buf:"u8"` // 1 = rotation grab-sphere handholds visible
	LabelsGlobal   uint8 `buf:"u8"` // 1 = all node labels visible
	OverlaysVis    uint8 `buf:"u8"` // 1 = master overlays toggle on
	// CascadeLinks mirrors the LAYOUT-link overlay's own visibility (default OFF, unlike the
	// other flags which default on) — the cyan second-tube overlay reads the LayoutLink block
	// only when this is set. NOT the same thing as the LayoutLink block existing: the data
	// streams every snapshot regardless, this just gates the render.
	CascadeLinks uint8 `buf:"u8"` // 1 = layout-link (cascade-link) overlay visible
	// DragNodeRow is the row index (into the Node block) of the node currently
	// being dragged by the gesture FSM (nodes/Wiring/gesture.go g.dragNode),
	// or -1 when no drag is in progress. Identity rides row index, not a
	// name/id sidecar; TS resolves the human name from that row's Label.
	DragNodeRow int32 `buf:"i32"` // dragged node's row index, -1 = not dragging
	// GroupLenTime/GroupLenInput/GroupLenGate are the "distance home button" toolbar
	// panel's 3 read-only group max-pair-lengths (nodes/Wiring/distance_groups.go's
	// distanceGroupOrder: time, input, gate). Computed fresh every VIEW-frame emit from
	// the live node centers (max over the group's pairs of |center(target)-center(source)|,
	// mirroring reachRFromPolar's max-over-edges loop) — Go owns the group definitions and
	// the math; the panel only reflects these 3 numbers and fires an edit-update on an
	// arrow click. 0 when a group's centers aren't resolvable yet.
	GroupLenTime  float32 `buf:"f32"` // time-nodes group's current max pair length
	GroupLenInput float32 `buf:"f32"` // input-node group's current max pair length
	GroupLenGate  float32 `buf:"f32"` // gate/pulse-nodes group's current max pair length
}

// bufLayoutScene defines the scene-sphere column block (always 1 row).
// Matched from KindSceneSphere trace events. The scene sphere is the persisted, first-class
// world anchor every node's scene polar is measured about (nodes/Wiring/sphere_layout.go
// sceneSphere) — established ONCE at load and never moved. Replaces the TS-side
// contentSphereFromCenters (a derived, non-authoritative content-sphere centroid recomputed
// from live node positions every frame) as the sphere NavGuides draws its polar tori around.
type bufLayoutScene struct {
	CX     float32 `buf:"f32"` // scene-sphere center x (world)
	CY     float32 `buf:"f32"` // scene-sphere center y (world)
	CZ     float32 `buf:"f32"` // scene-sphere center z (world)
	Radius float32 `buf:"f32"` // scene-sphere radius
}

// bufLayoutEvent defines one row of the per-tick EVENT column block.
// The block is self-sizing via an eventCount field in the snapshot header; it carries
// the causal trace events that occurred since the previous snapshot (recv/fire/send/
// arrive/position and the state-change kinds), cleared each emit like the
// transient node flags. It is consumed ONLY by the ext-host buffer-decoded .probe logger —
// the render path ignores it. Kind is the event's index into TRACE_EVENT_KINDS (shared
// Go/TS vocabulary); the row/label references resolve identities via the existing row
// tables + string sections, so no id/port/edge strings are duplicated per event.
// Sentinel: row/index fields are -1 when the event does not carry that reference.
type bufLayoutEvent struct {
	Kind          uint8  `buf:"u8"`  // index into TRACE_EVENT_KINDS
	NodeRow       int32  `buf:"i32"` // emitting node's buffer row (-1 = none)
	PortRow       int32  `buf:"i32"` // port's buffer row (-1 = none)
	TargetRow     int32  `buf:"i32"` // target node's buffer row (send; -1 = none)
	TargetPortRow int32  `buf:"i32"` // target handle's port row (send; -1 = none)
	EdgeRow       int32  `buf:"i32"` // edge's buffer row (geometry/select-edge; -1 = none)
	Slot          int32  `buf:"i32"` // node-bead interior slot = row*2+col (-1 = none)
	Value         int32  `buf:"i32"` // event value (recv/send/position/status/select mode/…)
	Bead          uint32 `buf:"u32"` // per-wire bead id (wire-bead events; 0 = none)
	// BeadSteps is a send event's edge bead-step count (docs/bead-lattice.md "The
	// count") — was ArcLength before the bead lattice replaced the arc-length model.
	BeadSteps    float32 `buf:"f32"` // send: edge bead-step count
	SimLatencyMs float32 `buf:"f32"` // send: wire traversal latency (ms), derived from BeadSteps
	X            float32 `buf:"f32"` // position/status world/marker x
	Y            float32 `buf:"f32"` // position/status world/marker y
	Z            float32 `buf:"f32"` // position/status world/marker z
	F            float32 `buf:"f32"` // position: fractional progress t
	// Label is the Breadcrumb sub-label enum (T.BreadcrumbLabels index) — only
	// meaningful when Kind == KindBreadcrumb; 0 otherwise (also a valid label id,
	// but other-kind rows never read Label). See Trace.go's BreadcrumbLabel* consts.
	Label uint8 `buf:"u8"` // breadcrumb sub-label index (Kind==breadcrumb only)
	// Debug flags this row as a DEBUG BREADCRUMB (vs. a structured domain trace
	// event) for the ext host's probe-merge.sh --debug filter. 1 = breadcrumb.
	Debug uint8 `buf:"u8"` // 1 = this row is a debug breadcrumb
	// TextOff/TextLen are this event's slice into the frame's trailing EVENT-TEXT
	// BYTES section (the sanctioned SINGLE free-form string escape hatch for the
	// event row — tools/check-event-string-section-singular.sh enforces at most
	// one such Off/Len pair on this struct). Used only for genuinely free-form
	// remainder text a breadcrumb payload doesn't fit into a typed column
	// (Value/X/Y/Z/NodeRow/PortRow/TargetRow/TargetPortRow/EdgeRow/Slot). TextLen=0
	// = no text.
	TextOff uint32 `buf:"u32"` // byte offset into the event-text-bytes section
	TextLen uint32 `buf:"u32"` // event text UTF-8 byte length
}

// schemaTypes prevents the bufLayout* types from being flagged as unused by
// staticcheck. They are schema sources: the generator reads them via AST at
// codegen time; they are not used at runtime.
var _ = [...]any{
	bufLayoutNode{},
	bufLayoutChainBead{},
	bufLayoutInterior{},
	bufLayoutEdge{},
	bufLayoutLayoutLink{},
	bufLayoutCamera{},
	bufLayoutOverlay{},
	bufLayoutScene{},
	bufLayoutEvent{},
}
