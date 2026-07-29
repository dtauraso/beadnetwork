// edge-stream-blocks.ts — the per-edge dedicated-stream either/or, mirroring
// view-blocks.ts's role for the VIEW stream (memory/feedback_no_single_writer_bridge.md).
// EdgeTube.tsx reads edge geometry/selection through this ONE
// accessor rather than each re-implementing the "no per-edge stream frame has arrived yet,
// so draw nothing this frame" null-check.

import { getLatestEdgeStreamFrames } from "../snapshot-buffer";
import { decodeEdgeStreamFrame, type DecodedEdgeStreamFrame } from "./buffer-decode";
import { readEdgeSrcPortRow, readEdgeDstPortRow, readEdgeSelected } from "../../schema/buffer-layout";

export interface EdgeAccessor {
  /** One past the highest edge ROW that has posted at least one dedicated-stream frame —
   *  NOT frames.size (a row can arrive out of order at startup; using the size would
   *  misnumber a sparse row set as a dense 0..size-1 range, corrupting the row identity
   *  every downstream pick/selection lookup depends on). A row with no frame yet reads as
   *  "unresolved" (-1 / not selected / no beads), same treatment writeEdgeBlock gives an
   *  unresolved port. */
  edgeCount: number;
  srcPortRow(row: number): number;
  dstPortRow(row: number): number;
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
    srcPortRow(row) {
      const d = decodedFor(frames, row);
      return d ? readEdgeSrcPortRow(d.edgeView, 0) : -1;
    },
    dstPortRow(row) {
      const d = decodedFor(frames, row);
      return d ? readEdgeDstPortRow(d.edgeView, 0) : -1;
    },
    selected(row) {
      const d = decodedFor(frames, row);
      return d ? readEdgeSelected(d.edgeView, 0) > 0 : false;
    },
  };
}
