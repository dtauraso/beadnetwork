// node-stream-blocks.ts — the per-node dedicated-stream aggregator, mirroring
// edge-stream-blocks.ts's role for the per-edge streams (memory/feedback_no_single_writer_bridge.md).
//
// Every render-tree consumer reads getNodeFrame() — ONE function returning a
// DecodedNodeFrame-shaped view aggregated from every node row's own dedicated NODE/INTERIOR
// stream frame (Buffer.BuildNodeStreamFrame/BuildInteriorStreamFrame — one goroutine's fd
// each), rewriting only the two offset columns (LabelOff/PortNameOff) that must point into
// the aggregated label/port-name byte sections instead of each frame's own inline bytes.
// This IS a byte copy (the source bytes live in N separate ArrayBuffers, one per node/
// interior fd) — but it happens once per (nodeStreamVersion, interiorStreamVersion) change,
// not once per render-tree consumer per frame (of which there are ~8), via the module-level
// memo below.
//
// A node row with no NODE-stream frame yet (arrived out of order at startup) is treated as
// an all-zero row (radius 0 falls back to NODE_SPHERE_RADIUS in the renderers via `|| `,
// portCount 0) — the same "unresolved" treatment edge-stream-blocks.ts gives a missing row.
// A node row with no INTERIOR-stream frame yet is treated as all-Present=0 (no interior
// beads drawn for that node until its own Update goroutine's first frame arrives).
//
// WIREFOLD_STREAM_FDS "node"+"interior" are now MANDATORY (the combined fallback frame was
// deleted, memory/feedback_no_single_writer_bridge.md's final step) — null means "no frame has arrived yet",
// not "the dedicated path is off".

import { getLatestNodeStreamFrames, getLatestInteriorStreamFrames, getNodeStreamVersion, getInteriorStreamVersion, subscribeNodeStreamFrame, subscribeInteriorStreamFrame } from "../snapshot-buffer";
import {
  decodeNodeStreamFrame, decodeInteriorStreamFrame,
  readNodeStreamLayoutLinkDstNodeRow,
  type DecodedNodeFrame,
  type DecodedNodeStreamFrame,
} from "./buffer-decode";
import {
  NODE_STRIDE, PORT_STRIDE, INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE,
  NODE_COL_LABEL_OFF, NODE_COL_LABEL_LEN,
  PORT_COL_PORT_NAME_OFF, PORT_COL_PORT_NAME_LEN,
  LAYOUT_LINK_STRIDE, LAYOUT_LINK_COL_SRC_NODE_ROW, LAYOUT_LINK_COL_DST_NODE_ROW,
  readNodeCX, readNodeCY, readNodeCZ,
  readChainBeadOX, readChainBeadOY, readChainBeadOZ, readChainBeadLit,
} from "../../schema/buffer-layout";

const STR_ENCODER = new TextEncoder();

// Memo keyed on the (nodeStreamVersion, interiorStreamVersion) pair — both bump on every
// setLatestNodeStreamFrame/setLatestInteriorStreamFrame call (snapshot-buffer.ts), so an
// unchanged pair means every consumer this render tick reads the SAME aggregate rather
// than each re-copying the same bytes.
let lastNodeVersion = -1;
let lastInteriorVersion = -1;
let lastAggregate: DecodedNodeFrame | null = null;

/**
 * getNodeFrame returns the current Node/Interior/Port(+Label/PortName) view, aggregated
 * from every node row's own dedicated stream frame. null until at least one has arrived.
 * Pure read (aside from its own memo) — no store writes.
 */
export function getNodeFrame(): DecodedNodeFrame | null {
  const nodeFrames = getLatestNodeStreamFrames();
  if (nodeFrames.size === 0) {
    return null;
  }
  const nv = getNodeStreamVersion();
  const iv = getInteriorStreamVersion();
  if (nv === lastNodeVersion && iv === lastInteriorVersion && lastAggregate) {
    return lastAggregate;
  }
  const aggregate = buildAggregate(nodeFrames, getLatestInteriorStreamFrames());
  lastNodeVersion = nv;
  lastInteriorVersion = iv;
  lastAggregate = aggregate;
  return aggregate;
}

/** Shape of the LayoutLink block the cascade-link overlay (EdgeTube.tsx) consumes — the
 *  SAME shape the old combined scene frame's LayoutLink block produced (SrcNodeRow/
 *  DstNodeRow, LAYOUT_LINK_STRIDE-byte rows — no EdgeRow: this is its OWN edge between
 *  the two nodes' CENTERS, never the bead-edge graph), so EdgeTube's read logic doesn't
 *  have to change shape. */
export interface LayoutLinkAgg {
  layoutLinkCount: number;
  layoutLinkView: DataView;
}

let lastLayoutLinkVersion = -1;
let lastLayoutLinkAgg: LayoutLinkAgg | null = null;

/**
 * getLayoutLinks returns the current cascade-link overlay pairs, aggregated from every
 * per-node dedicated NODE stream's own outbound cascade-edges (each node streams the
 * pairs for which it is the lexicographically-smaller endpoint — see node_mover.go's
 * cascadeEdges doc comment). Reconstructs full SrcNodeRow/DstNodeRow rows (SrcNodeRow =
 * the node row whose own frame carried that entry) so the aggregate is
 * BYTE-COMPATIBLE with the pre-migration shared block. Empty (layoutLinkCount 0) until
 * at least one node stream frame has arrived.
 */
export function getLayoutLinks(): LayoutLinkAgg {
  const nodeFrames = getLatestNodeStreamFrames();
  if (nodeFrames.size === 0) {
    return { layoutLinkCount: 0, layoutLinkView: new DataView(new ArrayBuffer(0)) };
  }
  const nv = getNodeStreamVersion();
  if (nv === lastLayoutLinkVersion && lastLayoutLinkAgg) {
    return lastLayoutLinkAgg;
  }

  let maxRow = -1;
  for (const r of nodeFrames.keys()) if (r > maxRow) maxRow = r;
  const nodeCount = maxRow + 1;

  const srcRows: number[] = [];
  const dstRows: number[] = [];
  for (let row = 0; row < nodeCount; row++) {
    const buf = nodeFrames.get(row);
    if (!buf) continue;
    const decoded = decodeNodeStreamFrame(row, buf);
    if (!decoded) continue;
    for (let i = 0; i < decoded.layoutLinkCount; i++) {
      srcRows.push(row);
      dstRows.push(readNodeStreamLayoutLinkDstNodeRow(decoded.layoutLinkView, i));
    }
  }

  const layoutLinkCount = srcRows.length;
  const layoutLinkView = new DataView(new ArrayBuffer(layoutLinkCount * LAYOUT_LINK_STRIDE));
  for (let i = 0; i < layoutLinkCount; i++) {
    const off = i * LAYOUT_LINK_STRIDE;
    layoutLinkView.setInt32(off + LAYOUT_LINK_COL_SRC_NODE_ROW, srcRows[i]!, true);
    layoutLinkView.setInt32(off + LAYOUT_LINK_COL_DST_NODE_ROW, dstRows[i]!, true);
  }

  const agg: LayoutLinkAgg = { layoutLinkCount, layoutLinkView };
  lastLayoutLinkVersion = nv;
  lastLayoutLinkAgg = agg;
  return agg;
}

/** Subscribe to either dedicated stream updating (subscribe-fn shape, e.g. for a React
 *  external-store hook) — the per-node analogue of view-blocks.ts's subscribeViewBlocks. */
export function subscribeNodeStreamBlocks(fn: () => void): () => void {
  const unsubNode = subscribeNodeStreamFrame(fn);
  const unsubInterior = subscribeInteriorStreamFrame(fn);
  return () => {
    unsubNode();
    unsubInterior();
  };
}

function buildAggregate(
  nodeFrames: ReadonlyMap<number, ArrayBuffer>,
  interiorFrames: ReadonlyMap<number, ArrayBuffer>,
): DecodedNodeFrame {
  // edgeCount-style sizing (edge-stream-blocks.ts's getEdgeStreamAccessor): one past the
  // highest row that has posted a frame, NOT frames.size — a sparse row set (arrived out
  // of order at startup) must not be misnumbered as a dense 0..size-1 range.
  let maxRow = -1;
  for (const r of nodeFrames.keys()) if (r > maxRow) maxRow = r;
  const nodeCount = maxRow + 1;

  const decodedByRow = new Map<number, ReturnType<typeof decodeNodeStreamFrame>>();
  let totalPortCount = 0;
  let totalLabelBytes = 0;
  let totalPortNameBytes = 0;
  for (let row = 0; row < nodeCount; row++) {
    const buf = nodeFrames.get(row);
    const decoded = buf ? decodeNodeStreamFrame(row, buf) : null;
    decodedByRow.set(row, decoded);
    if (decoded) {
      totalPortCount += decoded.portCount;
      totalLabelBytes += STR_ENCODER.encode(decoded.label).length;
      totalPortNameBytes += decoded.portNameBytes.byteLength;
    }
  }

  const interiorCount = nodeCount * INTERIOR_SLOTS_PER_NODE;
  const nodeBytes = nodeCount * NODE_STRIDE;
  const interiorBytes = interiorCount * INTERIOR_STRIDE;
  const portBytes = totalPortCount * PORT_STRIDE;

  const nodeBuf = new ArrayBuffer(nodeBytes);
  const nodeOut = new DataView(nodeBuf);
  const interiorBuf = new ArrayBuffer(interiorBytes);
  const interiorOut = new Uint8Array(interiorBuf);
  const portBuf = new ArrayBuffer(portBytes);
  const portOut = new Uint8Array(portBuf);
  const labelBytesOut = new Uint8Array(totalLabelBytes);
  const portNameBytesOut = new Uint8Array(totalPortNameBytes);

  let labelCursor = 0;
  let portNameCursor = 0;
  let portCursor = 0;

  for (let row = 0; row < nodeCount; row++) {
    const decoded = decodedByRow.get(row) ?? null;
    const nodeRowBytes = new Uint8Array(nodeBuf, row * NODE_STRIDE, NODE_STRIDE);
    if (decoded) {
      // Copy this node's own NODE_STRIDE row verbatim, then rewrite LabelOff to point
      // into the aggregated label-bytes section (LabelLen is already correct — it came
      // straight from this row's own bytes, unchanged by the copy).
      nodeRowBytes.set(new Uint8Array(decoded.nodeView.buffer, decoded.nodeView.byteOffset, NODE_STRIDE));
      const labelEncoded = STR_ENCODER.encode(decoded.label);
      nodeOut.setUint32(row * NODE_STRIDE + NODE_COL_LABEL_OFF, labelCursor, true);
      nodeOut.setUint32(row * NODE_STRIDE + NODE_COL_LABEL_LEN, labelEncoded.length, true);
      labelBytesOut.set(labelEncoded, labelCursor);
      labelCursor += labelEncoded.length;

      // Port rows: NodeRow column is already the global node row (BuildNodeStreamFrame
      // stamps it), so the raw bytes carry over verbatim except PortNameOff, which must
      // point into the aggregated port-name-bytes section.
      for (let p = 0; p < decoded.portCount; p++) {
        const srcOff = p * PORT_STRIDE;
        const rowBytes = new Uint8Array(decoded.portView.buffer, decoded.portView.byteOffset + srcOff, PORT_STRIDE);
        portOut.set(rowBytes, portCursor * PORT_STRIDE);
        const nameOff = decoded.portView.getUint32(srcOff + PORT_COL_PORT_NAME_OFF, true);
        const nameLen = decoded.portView.getUint32(srcOff + PORT_COL_PORT_NAME_LEN, true);
        const portOutView = new DataView(portBuf, portCursor * PORT_STRIDE, PORT_STRIDE);
        portOutView.setUint32(PORT_COL_PORT_NAME_OFF, portNameCursor, true);
        portOutView.setUint32(PORT_COL_PORT_NAME_LEN, nameLen, true);
        portNameBytesOut.set(decoded.portNameBytes.subarray(nameOff, nameOff + nameLen), portNameCursor);
        portNameCursor += nameLen;
        portCursor++;
      }
    }
    // A row with no frame yet stays all-zero (nodeRowBytes is already zero-initialized by
    // `new ArrayBuffer`) — the "unresolved" treatment this file's header comment describes.

    const interiorRowBytes = new Uint8Array(interiorBuf, row * INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE);
    const interiorFrameBuf = interiorFrames.get(row);
    const interiorDecoded = interiorFrameBuf ? decodeInteriorStreamFrame(row, interiorFrameBuf) : null;
    if (interiorDecoded) {
      interiorRowBytes.set(new Uint8Array(
        interiorDecoded.interiorView.buffer,
        interiorDecoded.interiorView.byteOffset,
        INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE,
      ));
    }
    // Missing interior frame ⇒ stays all-zero (Present=0 for every slot) — no interior
    // beads drawn for this node until its own Update goroutine's first frame lands.
  }

  return {
    tick: 0,
    nodeCount,
    nodeView: nodeOut,
    interiorCount,
    interiorView: new DataView(interiorBuf),
    portCount: totalPortCount,
    portView: new DataView(portBuf),
    labelBytesCount: totalLabelBytes,
    labelBytes: labelBytesOut,
    portNameBytes: portNameBytesOut,
  };
}

/** One node's placeholder chain beads, resolved to WORLD positions. The buffer carries
 *  NODE-LOCAL offsets (Go owns them, docs/beads-are-the-edge.md); this adds that node's own
 *  streamed center, exactly as the Interior block's slots are resolved — one add, no
 *  interpolation, no layout decision on this side.
 *
 *  A chain is the VISUAL of a traversal along this node's outgoing edges. It is NOT a
 *  picture of the node-to-node channels, and its length is not a count of messages: a chain
 *  sits fully populated with nothing traversing it. */
export interface ChainBeadsAgg {
  /** World-space bead centers, flat [x,y,z, x,y,z, …], every node's chains concatenated. */
  positions: Float32Array;
  /** Number of beads (positions.length / 3). */
  count: number;
  /** Per-bead Lit flag, parallel to positions: 1 where a traversal has currently reached.
   *  Go decides it (the source node reads its own wires' in-flight fraction); this layer
   *  only colours what the column says. All-zero is the normal resting state of a chain
   *  with nothing traversing it. */
  lit: Uint8Array;
}

let lastChainVersion = -1;
let lastChainAgg: ChainBeadsAgg | null = null;

/**
 * getChainBeads aggregates every per-node NODE stream's own chain beads into one
 * world-space position array. Each node contributes its own chains only — the source node
 * owns the whole chain for each of its outgoing edges (edges are stored under their source,
 * .claude/rules/persistence-ownership.md), so no bead is contributed twice and no bead's
 * position depends on any other bead's.
 *
 * Empty until at least one node stream frame has arrived. Cached on the node-frame version,
 * mirroring getLayoutLinks, so an unchanged scene re-uses the same array identity.
 */
export function getChainBeads(): ChainBeadsAgg {
  const nodeFrames = getLatestNodeStreamFrames();
  const nv = getNodeStreamVersion();
  if (lastChainAgg !== null && nv === lastChainVersion) {
    return lastChainAgg;
  }
  // Decode per row, same walk getLayoutLinks does — the frame map holds raw buffers.
  const decodedByRow: DecodedNodeStreamFrame[] = [];
  let total = 0;
  for (const [row, buf] of nodeFrames) {
    const decoded = decodeNodeStreamFrame(row, buf);
    if (!decoded) continue;
    decodedByRow.push(decoded);
    total += decoded.chainBeadCount;
  }
  const positions = new Float32Array(total * 3);
  const lit = new Uint8Array(total);
  let w = 0;
  let b = 0;
  for (const decoded of decodedByRow) {
    // This node's own streamed center — the offsets are relative to it (Interior-block
    // convention). One add per bead; no interpolation and no layout decision here.
    const cx = readNodeCX(decoded.nodeView, 0);
    const cy = readNodeCY(decoded.nodeView, 0);
    const cz = readNodeCZ(decoded.nodeView, 0);
    for (let i = 0; i < decoded.chainBeadCount; i++) {
      positions[w++] = cx + readChainBeadOX(decoded.chainBeadView, i);
      positions[w++] = cy + readChainBeadOY(decoded.chainBeadView, i);
      positions[w++] = cz + readChainBeadOZ(decoded.chainBeadView, i);
      lit[b++] = readChainBeadLit(decoded.chainBeadView, i);
    }
  }
  lastChainVersion = nv;
  lastChainAgg = { positions, count: total, lit };
  return lastChainAgg;
}
