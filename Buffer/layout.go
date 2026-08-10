// Buffer/layout.go — single source of truth for the agnostic content-buffer
// column layout schema.
//
// tools/gen-node-defs reads every *.go file directly under Buffer/ (scanning the
// directory, not one named file — parseBufferLayoutDir in
// tools/gen-node-defs/buf_layout_parse.go, same shape as
// parseInputLayoutFingerprintDir) and emits:
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
// → block "Bead"). Each block struct lives in its own Buffer/layout_*.go file
// (layout_node.go, layout_chain_bead.go, layout_interior.go, layout_edge.go,
// layout_camera.go, layout_overlay.go, layout_scene.go, layout_event.go); this
// file carries only the version consts, the historical-note comment, and the
// staticcheck anchor (schemaTypes below) that references every block type so
// none of them is flagged unused. BLOCK ORDER in the generated fingerprint is
// NOT file-scan order — it is the fixed bufBlockOrder list in
// tools/gen-node-defs/buf_layout_parse.go, so splitting this schema across
// files changed nothing the fingerprint depends on.
//
// BUF_LAYOUT_VERSION is bumped whenever any column definition changes; the
// generated files carry the same version so a stale regeneration is immediately
// visible.

package Buffer

// BufLayoutVersion is the schema version. Bump when any column changes.
const BufLayoutVersion = 41

// BufInteriorSlotsPerNode is the fixed number of interior grid slots reserved per
// node in the Interior block (a 2x2 held/interior-bead grid: slot = row*2 + col).
// The Interior block carries exactly nodeCount*BufInteriorSlotsPerNode rows in
// stable node order, so it needs no separate count in the header — the decoder
// derives its length from nodeCount. Not a per-column generated field (there is no
// bufLayoutInterior column for it), but gen-node-defs DOES read this const directly
// (parseBufferLayoutDir) and emits it as generated TS (INTERIOR_SLOTS_PER_NODE in
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

// The Bead block is GONE. It was one row per live in-flight bead (world X/Y/Z + Value),
// repacked every tick as the bead moved. Nothing draws a moving bead any more: a traversal
// is rendered as the LIT bead of a node-owned fixed chain (see bufLayoutChainBead's Lit
// column in layout_chain_bead.go and docs/bead-model/beads-are-the-edge.md), so there is no
// per-tick position to carry.

// The Port block is GONE (docs/bead-model/channels-not-ports.md): a port is a load-time
// channel-binding ROLE (PortSpec, a.In()/a.Out()), never a place, so it has no ring
// anchor, no world position, and no buffer row of its own any more. An edge's
// endpoints now ride the Edge block's own SX..EZ (bufLayoutEdge, layout_edge.go);
// hover/select address the NODE, not a per-port row.

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
