// edge-stream-blocks.ts — the per-edge dedicated-stream either/or, mirroring
// view-blocks.ts's role for the VIEW stream (memory/feedback_no_single_writer_bridge.md).
// EdgeTube.tsx reads edge geometry/selection through this ONE
// accessor rather than each re-implementing the "no per-edge stream frame has arrived yet,
// so draw nothing this frame" null-check.

import { getLatestEdgeStreamFrames } from "../snapshot-buffer";
import { decodeEdgeStreamFrame, type DecodedEdgeStreamFrame } from "./buffer-decode";
import {
  readEdgeSX, readEdgeSY, readEdgeSZ, readEdgeEX, readEdgeEY, readEdgeEZ,
} from "../../schema/buffer-layout";

export interface EdgeAccessor {
  /** One past the highest edge ROW that has posted at least one dedicated-stream frame —
   *  NOT frames.size (a row can arrive out of order at startup; using the size would
   *  misnumber a sparse row set as a dense 0..size-1 range, corrupting the row identity
   *  every downstream lookup depends on). A row with no frame yet reads as "unresolved"
   *  (segment 0,0,0->0,0,0). */
  edgeCount: number;
  /** This edge's own SEGMENT — node surface to node surface (docs/channels-not-ports.md).
   *  No port row to resolve any more: the edge carries its own endpoints directly. Still
   *  read for stream-capacity growth (buffer-scene.tsx) and the .probe debug decoder
   *  (buffer-log.ts) — NOT for rendering: the edge's own drawn line/pick halo is REMOVED,
   *  the source node's bead chain is the edge's visual now (docs/beads-are-the-edge.md). */
  segment(row: number): [number, number, number, number, number, number];
  // No selected() accessor: the Edge block's Selected column had exactly one caller (the
  // now-deleted pick halo, EdgeTube.tsx) and is now genuinely unreachable from the UI — the
  // pick target that ever tagged a hit with a buffer edge row is gone, so Go can never
  // receive an edge-select raw-input hit either (raw-input.ts's `edgeOnly` pick,
  // scene-content.tsx's pickBufferEdge). Left AS DATA (not deleted from Buffer/layout.go —
  // that is a bridge/schema change, a separate decision) — see
  // tools/check-no-dead-buffer-column.sh's ALLOWED_DEAD entry for readEdgeSelected.
  // No beads() accessor either: the Bead block is gone. A traversal renders as the LIT bead
  // of the source node's own fixed chain (ChainBeadInstances), not a moving position on the
  // edge stream — docs/beads-are-the-edge.md.
}

function decodedFor(frames: ReadonlyMap<number, ArrayBuffer>, row: number): DecodedEdgeStreamFrame | null {
  const buf = frames.get(row);
  return buf ? decodeEdgeStreamFrame(row, buf) : null;
}

/**
 * getEdgeStreamAccessor returns the per-edge dedicated-stream accessor when at least one
 * edge-stream frame has arrived (the dedicated path is active — WIREFOLD_STREAM_FDS
 * carried an "edge" entry), else null — callers nil-check this and draw nothing that
 * frame (no per-edge stream frame has arrived yet).
 */
export function getEdgeStreamAccessor(): EdgeAccessor | null {
  const frames = getLatestEdgeStreamFrames();
  if (frames.size === 0) return null;
  let maxRow = -1;
  for (const r of frames.keys()) if (r > maxRow) maxRow = r;
  const edgeCount = maxRow + 1;
  return {
    edgeCount,
    segment(row) {
      const d = decodedFor(frames, row);
      if (!d) return [0, 0, 0, 0, 0, 0];
      return [
        readEdgeSX(d.edgeView, 0), readEdgeSY(d.edgeView, 0), readEdgeSZ(d.edgeView, 0),
        readEdgeEX(d.edgeView, 0), readEdgeEY(d.edgeView, 0), readEdgeEZ(d.edgeView, 0),
      ];
    },
  };
}
