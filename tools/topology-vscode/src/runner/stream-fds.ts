// The fd-ALLOCATION contract (mirrors Buffer/stream_fds.go's doc comment): the ext host
// knows the topology from the spec it holds, computes a base fd PER STREAM KIND, and
// passes it to Go via WIREFOLD_STREAM_FDS = "kind:baseFd,kind:baseFd,…". VIEW_FD is the
// base (and, since view is a singleton stream — one gesture/MoveDispatch goroutine
// network-wide — also the ONLY) fd for the "view" kind: fd = baseFd["view"] + rowIndex,
// rowIndex always 0 for this singleton. This is the FIRST stream migrated off fd 3 onto
// its own dedicated inherited pipe (memory/feedback_no_single_writer_bridge.md).
export const VIEW_FD = 4;

// EDGE_BASE_FD: the base fd for the "edge" stream kind — one dedicated fd PER EDGE ROW,
// fd = EDGE_BASE_FD + edgeRow (edgeRow = that edge's stable seed-order row, matching
// Buffer's Edge block row order — see nodes/Wiring's MoveDispatch.SetEdgeStreams). Sits
// right after the view fd. Layout today: fd 0-2 stdin/stdout/stderr, fd 3 = fd3 (node/
// interior/port dual-path — see the module doc), fd 4 = view (singleton), fd 5..5+E-1 =
// one per edge (E = readCounts(topologyPath).edges below).
export const EDGE_BASE_FD = 5;

// MAX_EDGE_STREAMS bounds the per-edge fd range: one dedicated pipe PER EDGE (see
// EDGE_BASE_FD's doc comment) — fine for current graph sizes (this is a scaling bound the
// no-single-writer-bridge migration accepts explicitly, not an oversight). A topology with
// more edges than this omits the dedicated per-edge streams entirely (edgeCount is
// clamped, WIREFOLD_STREAM_FDS omits "edge", Go never calls SetEdgeStreamActive).
export const MAX_EDGE_STREAMS = 256;

// NODE_BASE_FD / INTERIOR_BASE_FD: the base fds for the "node" and "interior" stream
// kinds — one dedicated fd PER NODE ROW each, fd = base + nodeRow. ROW ID = NODE ID - 1
// (nodeRow is that node's own id minus one, not a position in a seed list — no ordering
// step decides it; see nodes/Wiring's MoveDispatch.SetNodeStreams / SetInteriorStreams and
// main.go), so this range must be sized by the LARGEST node id in the tree (counts.json's
// "nodes" field — see readCounts' doc comment), not by how many node directories exist: a
// gap left by a deleted node still needs its row's fd allocated. Sit right after the edge
// range, computed PER-SPAWN (not module-level constants) since they depend on edgeCount:
// nodeBase = EDGE_BASE_FD + edgeCount, interiorBase = nodeBase + nodeCount. Go's
// stream_fds.go requires BOTH "node" and "interior" WIREFOLD_STREAM_FDS entries present
// together (main.go only wires either when both resolve) — see run() below.

// nodeIdForRow / rowForNodeId make the ROW ID = NODE ID - 1 rule (stated in prose above,
// and in Buffer/stream_fds.go / nodes/Wiring's SetNodeStreams/SetInteriorStreams) into
// actual arithmetic instead of a comment nobody runs. Every place below that used to pass
// a bare per-node-pipe `row` into a cache key or a log/error message now goes through one
// of these, so a divergence between "which pipe" and "which node" would show up as a wrong
// id somewhere visible (a probe log line, an error message, a replayed frame) instead of
// silently-correct positional indexing forever hiding the untested rule. Gaps are legal:
// a deleted node still owns its row (see persistence-ownership.md), so an idle pipe simply
// means that node id never emits — it does NOT shift any other pipe's identity.
export function nodeIdForRow(row: number): number {
  return row + 1;
}
export function rowForNodeId(nodeId: number): number {
  return nodeId - 1;
}

// MAX_NODE_STREAMS bounds the per-node fd range (mirrors MAX_EDGE_STREAMS) — one
// dedicated pipe PER NODE for EACH of node/interior. A topology with more nodes than this
// omits the dedicated per-node NODE/INTERIOR/Port streams entirely.
export const MAX_NODE_STREAMS = 256;

// DRIVE_SLOTS_PER_NODE mirrors Buffer.DriveSlotsPerNode (Go) — the fixed number of
// dedicated "drive" fds allocated per node row, one per gatecommon.DriveHeld goroutine a
// node kind may spawn (docs/interior-stream-framing.md's fix: each such goroutine gets
// its OWN fd instead of sharing the node's "interior" fd with its Update-loop goroutine).
// Kept as a separate mirrored constant rather than a generated one, matching
// MAX_EDGE_STREAMS/MAX_NODE_STREAMS's existing "small bound, hand-kept in parity" shape —
// raise both sides together if a node kind ever needs a third DriveHeld output.
export const DRIVE_SLOTS_PER_NODE = 2;
