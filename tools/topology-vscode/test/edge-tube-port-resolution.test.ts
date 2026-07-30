// edge-tube-port-resolution.test.ts — proves EdgeTube.tsx reads an edge's SEGMENT
// straight off that edge's OWN dedicated stream frame (SX..EZ), with no port-row
// indirection through a separate node frame's own Port block (docs/channels-not-ports.md
// — there is no Port block any more). This is the tear-free property in its current form:
// the edge's endpoints are never a second, separately-timed copy of anything a node frame
// carries — they are the edge's own bytes.

import { describe, it, expect, vi } from "vitest";
import {
  BUF_EDGE_STREAM_FRAME_HEADER_SIZE,
} from "../src/schema/frame-tags";
import {
  EDGE_STRIDE, EDGE_COL_SX, EDGE_COL_SY, EDGE_COL_SZ, EDGE_COL_EX, EDGE_COL_EY, EDGE_COL_EZ,
} from "../src/schema/buffer-layout";

async function freshModules() {
  vi.resetModules();
  const snapshotBuffer = await import("../src/webview/snapshot-buffer");
  const edgeStreamBlocks = await import("../src/webview/three/edge-stream-blocks");
  return { snapshotBuffer, edgeStreamBlocks };
}

/** Build one edge's BUF_BLOCK_TAG_EDGE_STREAM frame: [tick] + 1 EDGE_STRIDE row (SX..EZ,
 *  label len 0). */
function makeEdgeStreamFrame(sx: number, sy: number, sz: number, ex: number, ey: number, ez: number): ArrayBuffer {
  const total = BUF_EDGE_STREAM_FRAME_HEADER_SIZE + EDGE_STRIDE;
  const buf = new ArrayBuffer(total);
  const dv = new DataView(buf);
  dv.setUint32(0, 1, true); // tick
  const off = BUF_EDGE_STREAM_FRAME_HEADER_SIZE;
  dv.setFloat32(off + EDGE_COL_SX, sx, true);
  dv.setFloat32(off + EDGE_COL_SY, sy, true);
  dv.setFloat32(off + EDGE_COL_SZ, sz, true);
  dv.setFloat32(off + EDGE_COL_EX, ex, true);
  dv.setFloat32(off + EDGE_COL_EY, ey, true);
  dv.setFloat32(off + EDGE_COL_EZ, ez, true);
  // EdgeLabelLen stays 0 (default) — no label bytes.
  return buf;
}

describe("EdgeTube segment resolution", () => {
  it("resolves an edge's SEGMENT straight off its own dedicated stream frame", async () => {
    const { snapshotBuffer, edgeStreamBlocks } = await freshModules();

    snapshotBuffer.setLatestEdgeStreamFrame(0, makeEdgeStreamFrame(10, 0, 0, 50, 0, 0));

    const edgeAccessor = edgeStreamBlocks.getEdgeStreamAccessor()!;
    expect(edgeAccessor).not.toBeNull();

    const [sx, sy, sz, ex, ey, ez] = edgeAccessor.segment(0);
    expect([sx, sy, sz]).toEqual([10, 0, 0]);
    expect([ex, ey, ez]).toEqual([50, 0, 0]);
  });

  it("a re-emitted frame for the SAME row replaces the segment (no stale copy)", async () => {
    const { snapshotBuffer, edgeStreamBlocks } = await freshModules();

    snapshotBuffer.setLatestEdgeStreamFrame(0, makeEdgeStreamFrame(10, 0, 0, 50, 0, 0));
    const before = edgeStreamBlocks.getEdgeStreamAccessor()!;
    expect(before.segment(0)).toEqual([10, 0, 0, 50, 0, 0]);

    // This edge's own edgeMover re-emits a fresh frame after a node move — the segment
    // moved to (10,0,0)->(99,0,0). No separate node-frame lookup is involved at all.
    snapshotBuffer.setLatestEdgeStreamFrame(0, makeEdgeStreamFrame(10, 0, 0, 99, 0, 0));
    const after = edgeStreamBlocks.getEdgeStreamAccessor()!;
    expect(after.segment(0)).toEqual([10, 0, 0, 99, 0, 0]);
  });
});
