// stdin_msg_types.go — the SHAPES on the editor→Go seam.
//
// One job: declare what a decoded message IS. Nothing here reads a byte, frames a record,
// or dispatches anything — input_codec.go fills these structs, Wiring's stdin_reader.go
// deframes, Wiring's stdin_dispatch.go routes. They live together because they are the
// vocabulary this package shares with itself, and because the field-by-field commentary is
// the contract: which numeric row means what, and what deliberately does NOT cross the
// bridge.
//
// None of these structs carry json tags: the seam is framed binary end to end, and the
// field ORDER is pinned by INPUT_LAYOUT_FINGERPRINT (input_fingerprint.go), not by a tag.

package inputcodec

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// EdgeEndpoints identifies the source and target node IDs (and the port handles)
// for one edge. Handles are needed to recompute the port-to-port arc length.
type EdgeEndpoints struct {
	Source       string
	Target       string
	SourceHandle string
	TargetHandle string
}

// StdinMsg is the single editor→Go bridge shape. For type=="edit", op is the sole
// remaining value "update", which sets an attribute on a typed entity — the sole live
// entity is overlays: Attr=="toggle" (Flag names one overlay). The other top-level types
// are raw-input (Event) and the bare command (save).
//
// These structs carry NO json tags: this seam is framed binary end to end and nothing
// unmarshals them (input_codec.go decodes the record). The wire field order is the
// INPUT_LAYOUT_FINGERPRINT, not a struct tag.
type StdinMsg struct {
	Type string
	Op   string
	Kind string
	Attr string
	Flag string
	// Num is the numeric payload for an op=="update" that carries a value rather than a
	// flag name — currently only clock/speed (the playback multiplier, sent in QUARTER-UNITS:
	// an integer 0..8 that clockAttrHandlers divides by 4 to get the real multiplier — see
	// its comment). Zero otherwise.
	Num int
	// X/Y is the drop's NDC for scene/create — where on SCREEN the node was dropped. Zero
	// for every other message. Screen rather than world because turning a drop into a place
	// needs the camera, which is Go's; and a point rather than a target because which node
	// the new one connects to is Go's decision from its own geometry.
	X, Y float64
	// Event is the payload for the top-level type=="raw-input" message; nil otherwise.
	Event *RawInputMsg
}

// RawInputMsg carries the payload for a top-level type=="raw-input" message (Phase 6):
// a single RAW pointer/wheel event plus the stateless three.js raycast hit. Go's gesture
// state machine (gesture.go) decides what it means — TS does not interpret it. Mirrors the
// TS RawInputEvent (messages.ts); the field ORDER is pinned by INPUT_LAYOUT_FINGERPRINT
// (input_codec.go), not by struct tags — this seam is framed binary, never JSON.
type RawInputMsg struct {
	Kind       string // pointerdown | pointermove | pointerup | wheel | home
	X          float64
	Y          float64 // client pixel X/Y
	RectLeft   float64
	RectTop    float64
	RectWidth  float64
	RectHeight float64
	Button     int // 0 primary, 2 secondary; -1 for move/wheel
	Ctrl       bool
	Shift      bool
	Alt        bool
	Meta       bool
	DeltaX     float64
	DeltaY     float64
	Fov        float64
	Hit        RawHit
}

// RawHit is the classified raycast hit: which rendered entity is under the pointer. Kind ∈
// port|handhold|node|edge|torus|empty. Topology facts (e.g. connected?) are NOT carried —
// Go's FSM decides those from its own held state. There is no world point on this record:
// any ray/plane unprojection Go needs is computed Go-side from the raw pointer NDC + Go's
// own camera/surface state (pointerOnRingPlane / rayDirThroughNDC in gesture.go).
type RawHit struct {
	Kind string
	// PortRow is the numeric buffer PORT-ROW index for a port hit (the port InstancedMesh
	// instanceId == its buffer port row). -1 (or absent) when not a port hit. Go resolves
	// this row → (node, port) via its own port-row table (portFromHit); no port name crosses
	// the bridge.
	PortRow int
	// EdgeRow is the numeric buffer EDGE-ROW index for an edge hit (the edge's pick-halo
	// carries its buffer edge row). -1 (or absent) when not an edge hit. Go resolves this
	// row → edge label via its own edge-row table (edgeFromHit); no edge label crosses the
	// bridge.
	EdgeRow int
	// NodeRow is the numeric buffer NODE-ROW index for a node hit (the node InstancedMesh
	// instanceId == its buffer node row). -1 (or absent) when not a node hit. Go resolves
	// this row → node id via its own node-row table (nodeFromHit); no node id crosses the
	// bridge.
	NodeRow int
	IsInput bool
}

// SlotRegistry maps "targetNodeId.targetHandle" → *PacedWire.
// It is the stable, slot-keyed identity for the wire owned by each destination port,
// consumed by md.Bind to seed edgeMovers (the create/delete edit ops that once indexed
// it were removed end-to-end).
type SlotRegistry map[string]*wire.PacedWire
