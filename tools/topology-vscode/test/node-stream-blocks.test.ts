// Unit tests for the per-node dedicated-stream decode + aggregation path
// (schema/frame-tags.ts's BUF_BLOCK_TAG_NODE_STREAM/BUF_BLOCK_TAG_INTERIOR_STREAM,
// buffer-decode.ts's decodeNodeStreamFrame/decodeInteriorStreamFrame, and
// three/node-stream-blocks.ts's getNodeFrame aggregator). No Port section any more
// (docs/channels-not-ports.md — a port carries no geometry, so there is nothing to
// decode/aggregate for it); an edge's endpoints ride its own frame's SX..EZ instead
// (edge-tube-port-resolution.test.ts).

import { describe, it, expect, vi } from "vitest";
import {
  decodeNodeStreamFrame, decodeInteriorStreamFrame,
  readNodeStreamLayoutLinkDstNodeRow,
} from "../src/webview/three/buffer-decode";
import {
  BUF_NODE_STREAM_FRAME_HEADER_SIZE, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE,
  NODE_STREAM_LAYOUT_LINK_STRIDE,
} from "../src/schema/frame-tags";
import {
  NODE_STRIDE, INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE,
  NODE_COL_NODE_ID, NODE_COL_CX, NODE_COL_CY, NODE_COL_CZ, NODE_COL_RADIUS,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readInteriorPresent, readInteriorValue,
  readLayoutLinkSrcNodeRow, readLayoutLinkDstNodeRow,
} from "../src/schema/buffer-layout";

// Every test below that touches the STATEFUL per-node cells (snapshot-buffer.ts's plain
// module-level maps) gets a FRESH module instance via vi.resetModules() + dynamic import —
// those maps persist across tests that share one import, so isolation must be explicit
// (mirrors the "fallback" test's isolation reasoning). freshNodeStreamModules() is the
// shared helper for that pattern.
async function freshNodeStreamModules() {
  vi.resetModules();
  const snapshotBuffer = await import("../src/webview/snapshot-buffer");
  const nodeStreamBlocks = await import("../src/webview/three/node-stream-blocks");
  return { snapshotBuffer, nodeStreamBlocks };
}

// ── helpers ───────────────────────────────────────────────────────────────────

function expectF32(got: number, want: number) {
  expect(got).toBeCloseTo(want, 5);
}

/** Build one node's BUF_BLOCK_TAG_NODE_STREAM frame: [tick][labelLen][layoutLinkCount]
 *  [chainBeadCount=0] + 1 Node row (LabelOff always 0 here) + label bytes + this node's
 *  own outbound LayoutLink rows ([DstNodeRow] each — no EdgeRow). No Port section any
 *  more (docs/channels-not-ports.md). */
function makeNodeStreamFrame(opts: {
  nodeRow: number;
  cx: number; cy: number; cz: number; radius: number;
  label: string;
  layoutLinks?: Array<{ dstNodeRow: number }>;
}): ArrayBuffer {
  const enc = new TextEncoder();
  const labelBytes = enc.encode(opts.label);
  const layoutLinks = opts.layoutLinks ?? [];

  const total = BUF_NODE_STREAM_FRAME_HEADER_SIZE + NODE_STRIDE + labelBytes.length
    + layoutLinks.length * NODE_STREAM_LAYOUT_LINK_STRIDE;
  const buf = new ArrayBuffer(total);
  const dv = new DataView(buf);
  dv.setUint32(0, 7, true); // tick
  dv.setUint32(4, labelBytes.length, true);
  dv.setUint32(8, layoutLinks.length, true);
  dv.setUint32(12, 0, true); // chainBeadCount = 0

  let off = BUF_NODE_STREAM_FRAME_HEADER_SIZE;
  // NodeId = nodeRow+1 (ROW ID = NODE ID - 1): keeps every decode below agreeing with the
  // row it's decoded on, so this helper doesn't itself trip the id/row mismatch report
  // decodeNodeStreamFrame now makes (task/row-fd-identity-parity). A dedicated mismatch is
  // exercised separately in stream-fixture.test.ts, not diluted across this file's frames.
  dv.setInt32(off + NODE_COL_NODE_ID, opts.nodeRow + 1, true);
  dv.setFloat32(off + NODE_COL_CX, opts.cx, true);
  dv.setFloat32(off + NODE_COL_CY, opts.cy, true);
  dv.setFloat32(off + NODE_COL_CZ, opts.cz, true);
  dv.setFloat32(off + NODE_COL_RADIUS, opts.radius, true);
  // LabelOff/LabelLen are set by Go to (0, labelLen) — see BuildNodeStreamFrame's doc
  // comment; the aggregator rewrites LabelOff, so its exact value here doesn't matter to
  // the aggregate assertions, but decodeNodeStreamFrame reads label straight from bytes,
  // not from these columns.
  off += NODE_STRIDE;

  new Uint8Array(buf, off, labelBytes.length).set(labelBytes);
  off += labelBytes.length;

  layoutLinks.forEach((ll, i) => {
    const rowOff = off + i * NODE_STREAM_LAYOUT_LINK_STRIDE;
    dv.setInt32(rowOff, ll.dstNodeRow, true);
  });

  return buf;
}

/** Build one node's BUF_BLOCK_TAG_INTERIOR_STREAM frame: [tick] + fixed
 *  INTERIOR_SLOTS_PER_NODE × INTERIOR_STRIDE rows. */
function makeInteriorStreamFrame(fill: (dv: DataView, rowOff: (slot: number) => number) => void): ArrayBuffer {
  const bytes = INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE;
  const buf = new ArrayBuffer(BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE + bytes);
  const dv = new DataView(buf);
  dv.setUint32(0, 3, true); // tick
  fill(dv, (slot) => BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE + slot * INTERIOR_STRIDE);
  return buf;
}

// ── decodeNodeStreamFrame / decodeInteriorStreamFrame ──────────────────────────

describe("decodeNodeStreamFrame", () => {
  it("decodes one node's geometry and label", () => {
    const buf = makeNodeStreamFrame({
      nodeRow: 5,
      cx: 1, cy: 2, cz: 3, radius: 9,
      label: "widget",
    });
    const d = decodeNodeStreamFrame(5, buf)!;
    expect(d).not.toBeNull();
    expect(d.label).toBe("widget");
    expectF32(readNodeCX(d.nodeView, 0), 1);
    expectF32(readNodeRadius(d.nodeView, 0), 9);
  });

  it("returns null for a truncated buffer", () => {
    expect(decodeNodeStreamFrame(0, new ArrayBuffer(2))).toBeNull();
  });

  it("decodes this node's own outbound layout-links (DstNodeRow only, no SrcNodeRow/EdgeRow — implicit/unused)", () => {
    const buf = makeNodeStreamFrame({
      nodeRow: 3, cx: 0, cy: 0, cz: 0, radius: 1, label: "n3",
      layoutLinks: [{ dstNodeRow: 7 }, { dstNodeRow: 9 }],
    });
    const d = decodeNodeStreamFrame(3, buf)!;
    expect(d).not.toBeNull();
    expect(d.layoutLinkCount).toBe(2);
    expect(readNodeStreamLayoutLinkDstNodeRow(d.layoutLinkView, 0)).toBe(7);
    expect(readNodeStreamLayoutLinkDstNodeRow(d.layoutLinkView, 1)).toBe(9);
  });
});

describe("decodeInteriorStreamFrame", () => {
  it("decodes the fixed 4-slot interior grid", () => {
    const buf = makeInteriorStreamFrame((dv, rowOff) => {
      dv.setUint8(rowOff(0), 1);
      dv.setInt32(rowOff(0) + 1, 1, true);
    });
    const d = decodeInteriorStreamFrame(0, buf)!;
    expect(d).not.toBeNull();
    expect(readInteriorPresent(d.interiorView, 0)).toBe(1);
    expect(readInteriorValue(d.interiorView, 0)).toBe(1);
    expect(readInteriorPresent(d.interiorView, 1)).toBe(0);
  });

  it("returns null for a truncated buffer", () => {
    expect(decodeInteriorStreamFrame(0, new ArrayBuffer(2))).toBeNull();
  });
});

// ── getNodeFrame: aggregation across two nodes ─────────────────────────────────

describe("getNodeFrame — aggregated dedicated streams", () => {
  it("aggregates node/label across rows, rewriting the LabelOff column to point into the\n" +
     "     aggregated section (not each node's own inline bytes)", async () => {
    const { snapshotBuffer, nodeStreamBlocks } = await freshNodeStreamModules();
    const frame0 = makeNodeStreamFrame({
      nodeRow: 0, cx: 100, cy: 0, cz: 0, radius: 5, label: "alpha",
    });
    const frame1 = makeNodeStreamFrame({
      nodeRow: 1, cx: 200, cy: 0, cz: 0, radius: 6, label: "beta",
    });
    snapshotBuffer.setLatestNodeStreamFrame(0, frame0);
    snapshotBuffer.setLatestNodeStreamFrame(1, frame1);
    snapshotBuffer.setLatestInteriorStreamFrame(0, makeInteriorStreamFrame(() => { /* all absent */ }));
    snapshotBuffer.setLatestInteriorStreamFrame(1, makeInteriorStreamFrame(() => { /* all absent */ }));

    const agg = nodeStreamBlocks.getNodeFrame()!;
    expect(agg).not.toBeNull();
    expect(agg.nodeCount).toBe(2);
    expectF32(readNodeCX(agg.nodeView, 0), 100);
    expectF32(readNodeCX(agg.nodeView, 1), 200);
  });

  it("treats a node row with no arrived NODE-stream frame as an unresolved zero row", async () => {
    const { snapshotBuffer, nodeStreamBlocks } = await freshNodeStreamModules();
    const frame0 = makeNodeStreamFrame({
      nodeRow: 0, cx: 1, cy: 1, cz: 1, radius: 1, label: "only-zero",
    });
    snapshotBuffer.setLatestNodeStreamFrame(0, frame0);
    // Simulate row 2 having arrived (sparse, out of order) but row 1 not yet.
    const frame2 = makeNodeStreamFrame({
      nodeRow: 2, cx: 9, cy: 9, cz: 9, radius: 9, label: "later",
    });
    snapshotBuffer.setLatestNodeStreamFrame(2, frame2);
    snapshotBuffer.setLatestInteriorStreamFrame(0, makeInteriorStreamFrame(() => {}));
    snapshotBuffer.setLatestInteriorStreamFrame(2, makeInteriorStreamFrame(() => {}));

    const agg = nodeStreamBlocks.getNodeFrame()!;
    expect(agg.nodeCount).toBe(3); // one past the highest arrived row
    expectF32(readNodeCX(agg.nodeView, 0), 1);
    // Row 1 never arrived — zeroed, not garbage.
    expectF32(readNodeCX(agg.nodeView, 1), 0);
    expectF32(readNodeRadius(agg.nodeView, 1), 0);
    expectF32(readNodeCX(agg.nodeView, 2), 9);
  });
});

describe("getNodeFrame — no per-node stream frame has arrived yet", () => {
  it("returns null (WIREFOLD_STREAM_FDS is mandatory — no fallback path)", async () => {
    const { nodeStreamBlocks } = await freshNodeStreamModules();
    expect(nodeStreamBlocks.getNodeFrame()).toBeNull();
  });
});

// ── getLayoutLinks: aggregation ─────────────────────────────────────────────────

describe("getLayoutLinks", () => {
  it("aggregates each per-node stream's own outbound layout-links into full Src/Dst rows", async () => {
    const { snapshotBuffer, nodeStreamBlocks } = await freshNodeStreamModules();
    const frame0 = makeNodeStreamFrame({
      nodeRow: 0, cx: 0, cy: 0, cz: 0, radius: 1, label: "a",
      layoutLinks: [{ dstNodeRow: 1 }],
    });
    const frame1 = makeNodeStreamFrame({
      nodeRow: 1, cx: 0, cy: 0, cz: 0, radius: 1, label: "b",
      layoutLinks: [{ dstNodeRow: 2 }],
    });
    const frame2 = makeNodeStreamFrame({
      nodeRow: 2, cx: 0, cy: 0, cz: 0, radius: 1, label: "c",
      // no outbound layout-links (b<c and c<? — c is never source in this fixture)
    });
    snapshotBuffer.setLatestNodeStreamFrame(0, frame0);
    snapshotBuffer.setLatestNodeStreamFrame(1, frame1);
    snapshotBuffer.setLatestNodeStreamFrame(2, frame2);

    const agg = nodeStreamBlocks.getLayoutLinks();
    expect(agg.layoutLinkCount).toBe(2);
    // Row order is source-node-row order (0 then 1) — SrcNodeRow is the reconstructed
    // implicit source, DstNodeRow carried straight from that node's own frame.
    expect(readLayoutLinkSrcNodeRow(agg.layoutLinkView, 0)).toBe(0);
    expect(readLayoutLinkDstNodeRow(agg.layoutLinkView, 0)).toBe(1);
    expect(readLayoutLinkSrcNodeRow(agg.layoutLinkView, 1)).toBe(1);
    expect(readLayoutLinkDstNodeRow(agg.layoutLinkView, 1)).toBe(2);
  });

  it("returns an empty aggregate when no per-node stream has arrived (WIREFOLD_STREAM_FDS is mandatory — no fallback path)", async () => {
    const { nodeStreamBlocks } = await freshNodeStreamModules();
    const agg = nodeStreamBlocks.getLayoutLinks();
    expect(agg.layoutLinkCount).toBe(0);
  });
});
