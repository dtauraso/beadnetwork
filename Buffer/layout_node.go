package Buffer

// bufLayoutNode defines one row of the nodes column block.
// One row per node. Persistent geometry. (Recv/Fire/Send/Arrive/Done events are
// carried by the EVENT block only — see bufLayoutEvent in layout_event.go — not
// by per-node columns.)
type bufLayoutNode struct {
	// NodeId is this node's stable, load-time numeric identity (1-based; ROW ID = NODE ID -
	// 1, so the loader/mover enforce Id == row+1 by construction — see
	// persistence-ownership.md and nodes/Wiring/node_mover.go). Before this column, a node's
	// identity existed ONLY as the row it arrived on: both Go and TS reconstructed it from
	// position by the SAME offline rule, so a frame could never contradict where it landed —
	// a permutation would render silently in the wrong place. This column makes that
	// contradiction detectable: the decoder compares Id against (arrival row + 1) and reports
	// a mismatch loudly (buffer-decode-node.ts's decodeNodeStreamFrame) instead of trusting the
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
	// one that wants an edge to lie in the ring plane (nodes/Wiring/nodegeom's PoleContainingEdge) —
	// and a scene that wants neither streams the torus's own +Z, which draws exactly as an
	// unrotated ring did. Reusing the pole for both would have made a rendering choice
	// change what navigation reads.
	RingAxisTheta float32 `buf:"f32"` // drawn ring's plane normal: θ from world +y (radians)
	RingAxisPhi   float32 `buf:"f32"` // drawn ring's plane normal: φ azimuth around +y (radians)
	// TopTiltVectorLen is how long the node's own drawn TOP TILT VECTOR is, along the same
	// axis as its ring (RingAxisTheta/Phi above), starting at the node's centre. ZERO means
	// this node draws no tilt vectors at all — so one column says both whether and how far,
	// for the top vector and for the bottom one below (which has no length column of its
	// own, the same way CoplanarNormal does not). A scene that wants no tilt vectors needs
	// no second flag anywhere. Go decides per scene; the renderer draws what it is given.
	TopTiltVectorLen float32 `buf:"f32"` // node top-tilt-vector length along the ring axis; 0 = no vectors
	// TopTiltVectorTheta is the top tilt vector's OWN direction, same angle convention as
	// PoleTheta/RingAxisTheta above (θ from world +y) — SEPARATE from RingAxisTheta so a
	// scene/user can point a node's tilt vector somewhere other than its ring axis. There is
	// no φ column: the whole tilt-vector model is θ-only (task/drop-tilt-vector-phi — every
	// φ in this exchange was always 0). Each node's mover holds this as an INTEGER index,
	// not a free float (memory/feedback_abc_times_constant_not_rederive.md): the streamed
	// value is index * nodes/Wiring.CurveParamTiltVectorAngleStep, and the index is what an
	// edit-update(tiltVector) click changes. Meaningless (but still streamed, default 0) on
	// a node whose TopTiltVectorLen is 0.
	TopTiltVectorTheta float32 `buf:"f32"` // top tilt vector direction: θ from world +y (radians) = thetaIdx*step
	// BottomTiltVectorTheta is the BOTTOM TILT VECTOR: the top tilt
	// vector turned a half turn (180°) in θ, so it points out of the node's other side. Same
	// angle convention and the same length as the top, drawn whenever TopTiltVectorLen is
	// non-zero, so there is no second length column.
	//
	// The half turn is ADDED, unmodified, by both nodes of a pair (nodes/PairNode). Both land
	// in the same drawn direction, since 180° in θ is the same place either way; the
	// addition is index bookkeeping only.
	BottomTiltVectorTheta float32 `buf:"f32"` // bottom tilt vector direction: θ from world +y (radians)
	CoplanarNormalTheta   float32 `buf:"f32"` // second vector direction: θ from world +y (radians)
	// ReceivedVectorLen/Theta are a THIRD drawn vector: the direction that LAST
	// ARRIVED on this node's own tilt-vector channel (nodes/Wiring/tilt_vector_channel.go),
	// kept by the RECEIVING node's own goroutine and replaced (never accumulated) by the
	// next arrival — see nodes/PairNode/node.go's handleVectorCycle.
	// Same "one column says both whether and how far" convention as TopTiltVectorLen above:
	// ZERO means this node has received nothing yet (or was reset), so a node with
	// nothing received is distinguishable from one whose received direction happens to be
	// 0 — the latter still streams a non-zero length. A RESET (this node's own
	// TiltEditIn Reset, or a Reset marker arriving on the channel) clears it back to zero:
	// a stale received arrow left hanging would contradict the reset's stop-and-return
	// meaning. Meaningless (but still streamed, default 0) on a node whose kind never
	// claims a vector channel — every kind but PairNode today.
	ReceivedVectorLen   float32 `buf:"f32"` // received-vector length; 0 = nothing received yet (or reset)
	ReceivedVectorTheta float32 `buf:"f32"` // received vector direction: θ from world +y (radians)
	Selected            uint8   `buf:"u8"`  // persistent: 1 = this node is the click-selected node
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
	// node's own nodeMover (nodes/Wiring/move_dispatch_construct.go). Replaces the old TS-owned
	// `latchedSel` React state in NavGuides.tsx (that was a second, TS-invented selection
	// concept unreachable from Go); the render path now just reads this column.
	LatchedSel uint8 `buf:"u8"` // 1 = this is the last-selected node (persists through deselect)
	// LatticePoints is THIS node's own pair-lattice point count — the N its own
	// TopTiltVectorTheta/BottomTiltVectorTheta/CoplanarNormalTheta/ReceivedVectorTheta
	// columns above were converted against (2π/LatticePoints per index step, this node's
	// own nodeGeometry.latticePoints — see writeStreamFrame's latticeThetaStep). Streamed
	// so a renderer-side reader (TiltVectorAnglePanel) can invert an angle back to an index
	// at the CURRENT count instead of assuming the compile-time default
	// (CurveParamTiltVectorAngleStep, fixed at 24 points) — the panel used to divide by
	// that fixed step regardless of what the scene's lattice actually holds, so it read
	// wrong the moment the scene setting changed the point count. Defaults to
	// FullTurnThetaIdx (24) on a node that never adopts a different count, matching
	// nodeGeometry.latticePoints' own default.
	LatticePoints uint8 `buf:"u8"` // this node's own lattice point count (4..64, multiple of 4)
	// RoundsToParallel is how many vector-exchange ROUNDS this node spent between the
	// exchange opening (START) and its own rule coming to rest — one round is one arrival
	// this node answered, so it counts this node's own arrivals, not the pair's combined
	// channel traffic (which is 4 per round for a pair: each end one send and one receive).
	//
	// It FREEZES at rest and does not keep climbing. The exchange keeps circulating after
	// both ends settle — stepFromVector replies to every arrival whether or not it moved —
	// so a live counter would measure how long the scene has been open rather than how far
	// the tilt had to travel. Zero means not yet at rest, or opened already at rest (a pair
	// preset to exactly a quarter turn adopts the perpendicular machine and never steps).
	RoundsToParallel int32 `buf:"i32"` // rounds from START until this node's rule came to rest; 0 = not yet / none needed
	// MsgsToParallel is the same span counted in MESSAGES rather than rounds — every
	// vector-channel receive AND send this node performed, including the START opener on
	// the node that sends it. A node that answered two arrivals did four messages, so this
	// is twice RoundsToParallel on the answering side and one more than that on the opener.
	//
	// Both are streamed rather than one being divided out on the render side, because the
	// two differ by the opener and a reader dividing by two would be silently wrong for the
	// node that opened the exchange.
	MsgsToParallel int32 `buf:"i32"` // vector-channel messages (receives + sends) over the same span
}
