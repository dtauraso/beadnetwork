// The fd-allocation arithmetic for ONE spawn: how many dedicated pipes to open, at which
// stdio indices, and what to spell into WIREFOLD_STREAM_FDS — computed from the stored
// topology counts alone, with no side effects (no channel, no goErrorsFile write). Split out
// of runCommand.ts's run() so the arithmetic reads as one pure function instead of being
// interleaved with `this.channel.appendLine`/`appendGoError` calls; the caller still does the
// reporting (see run()'s `for (const w of layout.warnings)` loop), because logging is the
// runner's job, not this file's.
import {
  EDGE_BASE_FD,
  MAX_EDGE_STREAMS,
  MAX_NODE_STREAMS,
  DRIVE_SLOTS_PER_NODE,
  VIEW_FD,
} from "./stream-fds";

export interface SpawnLayout {
  edgeCount: number;
  nodeCount: number;
  nodeBaseFd: number;
  interiorBaseFd: number;
  driveBaseFd: number;
  /** The `stdio` array passed to `cp.spawn` (index 0..2 are stdin/stdout/stderr, 3 is the
   *  reserved unused slot, VIEW_FD is the view stream, and the rest are the edge/node/
   *  interior/drive ranges below it — see the doc comment on the stdio build in run()). */
  stdio: Array<"pipe">;
  /** The value to assign to the child's `WIREFOLD_STREAM_FDS` env var. */
  streamFDsEnv: string;
  /** Operational warnings (a stored count over its MAX_*_STREAMS cap) to report through the
   *  runner's own channel + goErrorsFile — never thrown, since an oversized topology is
   *  legitimate input, not a code bug (see MAX_EDGE_STREAMS's doc comment). */
  warnings: string[];
}

/** Size the dedicated per-edge/per-node/per-interior/per-drive fd ranges from the stored
 *  topology counts. Mirrors Go's own three-way "node"+"interior"+"drive" env check
 *  (main.go) — all three are pushed together or none are. */
export function computeSpawnLayout(counts: { nodes: number; edges: number }): SpawnLayout {
  const warnings: string[] = [];

  // Clamped to MAX_EDGE_STREAMS; a count above the bound omits the dedicated per-edge
  // streams entirely — see MAX_EDGE_STREAMS's doc comment.
  const edgeCountRaw = counts.edges;
  const edgeCount = edgeCountRaw > MAX_EDGE_STREAMS ? 0 : edgeCountRaw;
  if (edgeCountRaw > MAX_EDGE_STREAMS) {
    // Capacity limit reached by legitimate input (a large topology), not a code bug — a
    // panic/throw here would be wrong. But silently zeroing edgeCount disables ALL
    // dedicated per-edge streams with no signal, which is the quietest possible failure
    // for the loudest consequence (this is the same path that used to strand the
    // `pending` leak, fixed in 93d2e9b6). Report it loudly through the same error
    // channel as every other operational problem in this file.
    warnings.push(
      `edge count ${edgeCountRaw} exceeds MAX_EDGE_STREAMS (${MAX_EDGE_STREAMS}); disabling ALL dedicated per-edge streams for this run`,
    );
  }
  // Size the dedicated per-node NODE + INTERIOR fd ranges the same way, right after the
  // edge range (nodeBase = EDGE_BASE_FD + edgeCount, interiorBase = nodeBase + nodeCount —
  // see NODE_BASE_FD's doc comment). Clamped to MAX_NODE_STREAMS; 0 omits the dedicated
  // per-node NODE/INTERIOR/Port streams entirely.
  const nodeCountRaw = counts.nodes;
  const nodeCount = nodeCountRaw > MAX_NODE_STREAMS ? 0 : nodeCountRaw;
  if (nodeCountRaw > MAX_NODE_STREAMS) {
    // Same reasoning as the edge-count case above.
    warnings.push(
      `node count ${nodeCountRaw} exceeds MAX_NODE_STREAMS (${MAX_NODE_STREAMS}); disabling ALL dedicated per-node NODE/INTERIOR streams for this run`,
    );
  }
  const nodeBaseFd = EDGE_BASE_FD + edgeCount;
  const interiorBaseFd = nodeBaseFd + nodeCount;
  // driveBaseFd sits right after the interior range: nodeCount * DRIVE_SLOTS_PER_NODE
  // dedicated fds, one PER (node row, drive slot) — see DRIVE_SLOTS_PER_NODE's doc
  // comment and Buffer/stream_fds.go's StreamKindDrive. Required in lockstep with
  // "node"/"interior" (see the streamFDsEnvParts push below and main.go's matching
  // three-way check) — Go falls back to a loud stderr message and unwired streams
  // rather than a startup panic if this ever drifts from what Go expects (never a
  // crash-loop; see the panic-avoidance note on that fallback in main.go).
  const driveBaseFd = interiorBaseFd + nodeCount;

  // "pipe" opens a readable pipe at each index; the existing stdin(0)/stdout(1)/stderr(2)
  // are unchanged. stdio index 3 is a RESERVED, UNUSED pipe slot: Go no longer writes
  // anything to fd 3 (see run()'s stdio comment for the full index layout).
  const stdio: Array<"pipe"> = ["pipe", "pipe", "pipe", "pipe", "pipe"];
  for (let i = 0; i < edgeCount; i++) stdio.push("pipe");
  for (let i = 0; i < nodeCount; i++) stdio.push("pipe");
  for (let i = 0; i < nodeCount; i++) stdio.push("pipe");
  for (let i = 0; i < nodeCount * DRIVE_SLOTS_PER_NODE; i++) stdio.push("pipe");

  const streamFDsEnvParts = [`view:${VIEW_FD}`];
  if (edgeCount > 0) streamFDsEnvParts.push(`edge:${EDGE_BASE_FD}`);
  // Go's stream_fds.go / main.go only wires the per-node node+interior+drive streams
  // when "node", "interior", AND "drive" env entries ALL resolve — always emit all
  // three together (main.go's three-way check treats a partial set the same as none).
  if (nodeCount > 0) {
    streamFDsEnvParts.push(`node:${nodeBaseFd}`, `interior:${interiorBaseFd}`, `drive:${driveBaseFd}`);
  }
  const streamFDsEnv = streamFDsEnvParts.join(",");

  return { edgeCount, nodeCount, nodeBaseFd, interiorBaseFd, driveBaseFd, stdio, streamFDsEnv, warnings };
}
