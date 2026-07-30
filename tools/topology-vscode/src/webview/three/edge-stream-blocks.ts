// edge-stream-blocks.ts — the per-edge dedicated-stream either/or, mirroring
// view-blocks.ts's role for the VIEW stream (memory/feedback_no_single_writer_bridge.md).
// EdgeTube.tsx reads edge geometry/selection through this ONE
// accessor rather than each re-implementing the "no per-edge stream frame has arrived yet,
// so draw nothing this frame" null-check.

import { getLatestEdgeStreamFrames } from "../snapshot-buffer";
import { decodeEdgeStreamFrame, type DecodedEdgeStreamFrame } from "./buffer-decode";
import {
  readEdgeSX, readEdgeSY, readEdgeSZ, readEdgeEX, readEdgeEY, readEdgeEZ,
  readEdgeSelected,
} from "../../schema/buffer-layout";

export interface EdgeAccessor {
  /** One past the highest edge ROW that has posted at least one dedicated-stream frame —
   *  NOT frames.size (a row can arrive out of order at startup; using the size would
   *  misnumber a sparse row set as a dense 0..size-1 range, corrupting the row identity
   *  every downstream pick/selection lookup depends on). A row with no frame yet reads as
   *  "unresolved" (segment 0,0,0->0,0,0 / not selected). */
  edgeCount: number;
  /** This edge's own SEGMENT — node surface to node surface (docs/channels-not-ports.md).
   *  No port row to resolve any more: the edge carries its own endpoints directly. */
  segment(row: number): [number, number, number, number, number, number];
  selected(row: number): boolean;
  // No beads() accessor: the Bead block is gone. A traversal renders as the LIT bead of
  // the source node's own fixed chain (ChainBeadInstances), not a moving position on the
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
    selected(row) {
      const d = decodedFor(frames, row);
      return d ? readEdgeSelected(d.edgeView, 0) > 0 : false;
    },
  };
}
