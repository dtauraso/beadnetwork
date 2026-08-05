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
const BufLayoutVersion = 40

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
	// NodeId is this node's stable, load-time numeric identity (1-based; ROW ID = NODE ID -
	// 1, so the loader/mover enforce Id == row+1 by construction — see
	// persistence-ownership.md and nodes/Wiring/node_mover.go). Before this column, a node's
	// identity existed ONLY as the row it arrived on: both Go and TS reconstructed it from
	// position by the SAME offline rule, so a frame could never contradict where it landed —
	// a permutation would render silently in the wrong place. This column makes that
	// contradiction detectable: the decoder compares Id against (arrival row + 1) and reports
	// a mismatch loudly (buffer-decode.ts's decodeNodeStreamFrame) instead of trusting the
	// row. This is NOT the removed id/label/kind sidecar message (bridge-surface.md's
	// no-sidecar clause) — it is a column inside the Node block, the same shape as KindId.
	NodeId  int32   `buf:"i32"` // this node's own numeric id (1-based); row+1 by construction
	CX      float32 `buf:"f32"` // node center x (world)
	CY      float32 `buf:"f32"` // node center y (world)
	CZ      float32 `buf:"f32"` // node center z (world)
	Radius  float32 `buf:"f32"` // body/ring sphere radius
	SphereR float32 `buf:"f32"` // sphere-chain reach radius
	// VR*/FR* are the node's two great-circle ring-plane normals (vertical vr, flat fr),
	// the same orientation vectors the pre-branch SphereRing read from nodeGeometryStore.
	// SphereRing draws two tori at the owner's center oriented by these; they arrive on the
	// node-geometry trace event (Trace VRX.., FRX..).
	VRX float32 `buf:"f32"` // vertical ring-plane normal x
	VRY float32 `buf:"f32"` // vertical ring-plane normal y
	VRZ float32 `buf:"f32"` // vertical ring-plane normal z
	FRX float32 `buf:"f32"` // flat (equatorial) ring-plane normal x
	FRY float32 `buf:"f32"` // flat ring-plane normal y
	FRZ float32 `buf:"f32"` // flat ring-plane normal z
	// PoleTheta/PolePhi are the direction this node's OWN local polar frame is poled at,
	// as an angle pair in the scene frame (same convention as everywhere else: θ = angle
	// from world +y, 0..π; φ = azimuth around +y, -π..π). It is the node's own scene-polar
	// direction REVERSED — the node's frame points back at the scene centre — derived by
	// that node's own mover from its own ScenePolar as (π−θ, φ+π), pure angle arithmetic on
	// a coordinate it already owns (nodes/Wiring/polar.go inwardPole). Angles, not a unit
	// vector like VR*/FR* above: those are fixed cartesian constants, this is the node's
	// polar coordinate and stays in the polar vocabulary until the renderer edge, where
	// NavGuides converts once to orient the frame. (0,0) = world +y, the value a node
	// without a position yet carries.
	PoleTheta float32 `buf:"f32"` // this node's local-frame pole: θ from world +y (radians)
	PolePhi   float32 `buf:"f32"` // this node's local-frame pole: φ azimuth around +y (radians)
	// RingAxisTheta/RingAxisPhi are the axis this node's drawn TORUS is poled at — the
	// normal of the ring's plane, same angle convention as PoleTheta/PolePhi above.
	//
	// SEPARATE from the pole on purpose. The pole is the node's own local polar frame,
	// consumed by navigation (buffer-nav.ts, NavGuides); the ring axis is what the ring is
	// DRAWN with. They coincide in a scene that wants its rings poled inward, and differ in
	// one that wants an edge to lie in the ring plane (nodes/Wiring's poleContainingEdge) —
	// and a scene that wants neither streams the torus's own +Z, which draws exactly as an
	// unrotated ring did. Reusing the pole for both would have made a rendering choice
	// change what navigation reads.
	RingAxisTheta float32 `buf:"f32"` // drawn ring's plane normal: θ from world +y (radians)
	RingAxisPhi   float32 `buf:"f32"` // drawn ring's plane normal: φ azimuth around +y (radians)
	// VectorLen is how long the node's own drawn VECTOR is, along the same axis as its
	// ring (RingAxisTheta/Phi above), starting at the node's centre. ZERO means this node
	// draws no vector — so one column says both whether and how far, and a scene that wants
	// no vectors needs no second flag anywhere. Go decides per scene; the renderer draws
	// what it is given.
	VectorLen float32 `buf:"f32"` // node vector length along the ring axis; 0 = no vector
	// VectorTheta/VectorPhi are the vector's OWN direction, same angle convention as
	// PoleTheta/PolePhi and RingAxisTheta/RingAxisPhi above (θ from world +y, φ azimuth
	// around +y) — SEPARATE from RingAxisTheta/Phi so a scene/user can point a node's
	// vector somewhere other than its ring axis. Each node's mover holds these as an
	// INTEGER index pair, not a free float (memory/feedback_abc_times_constant_not_
	// rederive.md): the streamed value is index * nodes/Wiring.VectorAngleStep, and the
	// index is what an edit-update(nodeVector) click changes. Meaningless (but still
	// streamed, default 0) on a node whose VectorLen is 0.
	VectorTheta float32 `buf:"f32"` // node vector direction: θ from world +y (radians) = thetaIdx*step
	VectorPhi   float32 `buf:"f32"` // node vector direction: φ azimuth around +y (radians) = phiIdx*step
	Selected    uint8   `buf:"u8"`  // persistent: 1 = this node is the click-selected node
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
	// There is no per-bead Radius column. Under bead CRUD (MODEL.md "Moving a node is
	// CRUD on the edge beads that touch it", nodes/Wiring/bead_crud.go) the single global
	// wire.BeadRadius/wire.BeadStepR lattice constants already make every chain's beads
	// touch their own neighbours on the chain exactly — a per-edge radius (added in
	// commit d50fab83, removed here) is unnecessary and was removed along with the
	// residue it existed to absorb.
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
	bufLayoutCamera{},
	bufLayoutOverlay{},
	bufLayoutScene{},
	bufLayoutEvent{},
}
