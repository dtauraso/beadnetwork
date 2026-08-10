package Buffer

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
	// BeadSteps is a send event's edge bead-step count (docs/bead-model/bead-lattice.md "The
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
	// event row — tools/buffer-schema/check-event-string-section-singular.sh enforces at most
	// one such Off/Len pair on this struct). Used only for genuinely free-form
	// remainder text a breadcrumb payload doesn't fit into a typed column
	// (Value/X/Y/Z/NodeRow/PortRow/TargetRow/TargetPortRow/EdgeRow/Slot). TextLen=0
	// = no text.
	TextOff uint32 `buf:"u32"` // byte offset into the event-text-bytes section
	TextLen uint32 `buf:"u32"` // event text UTF-8 byte length
}
