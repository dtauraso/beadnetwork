// node_frame_input.go — the WIRE-FRAME shape a NodeGeometry hands its injected packer, as
// opposed to node_geometry_parts.go's composer sub-structs (a node's own OWNED state).
// NodeFrameInput is not state this type holds between calls; it is the argument built once
// per emit and handed to NodeFrameBuilder, a genuinely separate concern from "what a node
// composes itself out of" — split out of node_geometry_parts.go
// (docs/planning/movedispatch-decomposition.md §20).
package nodeactor

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// NodeFrameInput is one node's own frame values, handed BY NAME to the injected packer
// (NodeFrameBuilder → Buffer.BuildNodeStreamFrame, mapped field-by-name in main.go so this
// package stays Buffer-independent).
//
// WHY A STRUCT: this was a ~35-parameter positional argument list — 20 consecutive
// float32s, then 5 uint8s, then 2 int32s — where transposing two adjacent same-typed
// arguments COMPILED CLEANLY and streamed a silently wrong scene. Named fields make that
// bug class unrepresentable (memory/feedback_make_bug_class_unrepresentable.md).
//
// Field order MIRRORS the packed column order for readability; the wire format is defined
// by Buffer's packing walk, never by this struct's order.
type NodeFrameInput struct {
	// Tick is this node's own clock tick, stamped on the frame.
	Tick uint32
	// NodeRow is this node's stable buffer row; NodeID is its numeric identity
	// (NodeRow+1 by construction — ROW ID = NODE ID - 1, persistence-ownership.md).
	NodeRow int32
	NodeID  int32
	// CX/CY/CZ: this node's own world-space centre.
	CX, CY, CZ float32
	// Radius: this node's own drawn radius. SphereR: the scene sphere's radius.
	Radius, SphereR float32
	// The vertical- and flat-ring normals.
	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32
	// PoleTheta/PolePhi: this node's own local-frame pole — its scene-polar direction
	// reversed, so the frame points back at the scene centre (world +y before HasPos).
	PoleTheta, PolePhi float32
	// RingAxisTheta/RingAxisPhi: the DRAWN ring's axis, separate from the navigation pole.
	RingAxisTheta, RingAxisPhi float32
	// TopTiltVectorLen encodes WHETHER-AND-HOW-FAR: 0 where the scene draws no vector,
	// otherwise the node's own radius (the vector runs centre → top).
	TopTiltVectorLen float32
	// TopTiltVectorTheta: the vector's OWN direction, separate from the ring axis. Never a
	// free float — index × this node's own lattice step. θ-only, there is no φ.
	TopTiltVectorTheta float32
	// BottomTiltVectorTheta: a half turn in θ from the top index, mirrored straight from
	// this node's own bottomThetaIdx; it shares the top's length column.
	BottomTiltVectorTheta float32
	// CoplanarNormalTheta: a fixed quarter turn in θ from the top index, so turning the
	// tilt visibly turns the normal WITH it.
	CoplanarNormalTheta float32
	// ReceivedVectorLen: same whether-and-how-far convention as TopTiltVectorLen — ZERO
	// when nothing has been received on this node's tilt-vector channel, so a node with
	// nothing received stays distinguishable from one whose received direction is 0.
	ReceivedVectorLen float32
	// ReceivedVectorTheta: that received direction (meaningless at zero length).
	ReceivedVectorTheta float32
	// This node's OWN selection-UI bits and its static NODE_DEFS kind index.
	Selected, KindID, Hovered, LatchedSel uint8
	// LatticePoints: this node's own lattice size — the N the θ columns were converted
	// against.
	LatticePoints uint8
	// RoundsToParallel/MsgsToParallel: this node's own rounds- and messages-to-rest counts.
	RoundsToParallel, MsgsToParallel int32
	// Label: this node's own label bytes, packed inline in its own frame.
	Label string
	// The PARALLEL chain-bead slices (one entry per placeholder bead, node-local offsets).
	ChainBeadOX, ChainBeadOY, ChainBeadOZ []float32
	ChainBeadLit                          []uint8
	ChainBeadLitValue                     []int32
	// Events rides this frame's trailing EVENTS section (nil from a plain tick-driven
	// write). Each must claim this node's OWN row — writeStreamFrame panics otherwise.
	Events []wire.RowEvent
}

// NodeFrameBuilder is the injected node-frame packer (Buffer.BuildNodeStreamFrame behind a
// Buffer-independent seam — see NodeFrameInput). One named-struct argument, not a
// positional run.
type NodeFrameBuilder func(f NodeFrameInput) []byte
