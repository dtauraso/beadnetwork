import { getLatestNodeStreamFrames, getLatestInteriorStreamFrames, getNodeStreamVersion, getInteriorStreamVersion, subscribeNodeStreamFrame, subscribeInteriorStreamFrame } from "../../../snapshot-buffer";
import {
  decodeNodeStreamFrame,
  type DecodedNodeFrame,
  type DecodedNodeStreamFrame,
} from "../../decode/buffer-decode-node";
import { decodeInteriorStreamFrame } from "../../decode/buffer-decode-interior";
import {
  NODE_STRIDE, INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE, TILT_ARROW_STRIDE, CHANNEL_VECTOR_STRIDE,
  NODE_COL_LABEL_OFF, NODE_COL_LABEL_LEN,
} from "../../../../../Buffer/buffer-layout";

const STR_ENCODER = new TextEncoder();

let lastNodeVersion = -1;
let lastInteriorVersion = -1;
let lastAggregate: DecodedNodeFrame | null = null;

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

export function getNodeStreamFrameForRow(row: number): DecodedNodeStreamFrame | null {
  const buf = getLatestNodeStreamFrames().get(row);
  return buf ? decodeNodeStreamFrame(row, buf) : null;
}

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

  let maxRow = -1;
  for (const r of nodeFrames.keys()) if (r > maxRow) maxRow = r;
  const nodeCount = maxRow + 1;

  const decodedByRow = new Map<number, ReturnType<typeof decodeNodeStreamFrame>>();
  let totalLabelBytes = 0;
  let totalTiltArrows = 0;
  let totalChannelVectors = 0;
  for (let row = 0; row < nodeCount; row++) {
    const buf = nodeFrames.get(row);
    const decoded = buf ? decodeNodeStreamFrame(row, buf) : null;
    decodedByRow.set(row, decoded);
    if (decoded) {
      totalLabelBytes += STR_ENCODER.encode(decoded.label).length;
      totalTiltArrows += decoded.tiltArrowCount;
      totalChannelVectors += decoded.channelVectorCount;
    }
  }
  const tiltArrowBuf = new ArrayBuffer(totalTiltArrows * TILT_ARROW_STRIDE);
  const tiltArrowOut = new Uint8Array(tiltArrowBuf);
  let tiltArrowCursor = 0;
  const channelVectorBuf = new ArrayBuffer(totalChannelVectors * CHANNEL_VECTOR_STRIDE);
  const channelVectorOut = new Uint8Array(channelVectorBuf);
  let channelVectorCursor = 0;

  const interiorCount = nodeCount * INTERIOR_SLOTS_PER_NODE;
  const nodeBytes = nodeCount * NODE_STRIDE;
  const interiorBytes = interiorCount * INTERIOR_STRIDE;

  const nodeBuf = new ArrayBuffer(nodeBytes);
  const nodeOut = new DataView(nodeBuf);
  const interiorBuf = new ArrayBuffer(interiorBytes);
  const interiorOut = new Uint8Array(interiorBuf);
  const labelBytesOut = new Uint8Array(totalLabelBytes);

  let labelCursor = 0;

  for (let row = 0; row < nodeCount; row++) {
    const decoded = decodedByRow.get(row) ?? null;
    const nodeRowBytes = new Uint8Array(nodeBuf, row * NODE_STRIDE, NODE_STRIDE);
    if (decoded) {

      nodeRowBytes.set(new Uint8Array(decoded.nodeView.buffer, decoded.nodeView.byteOffset, NODE_STRIDE));
      const labelEncoded = STR_ENCODER.encode(decoded.label);
      nodeOut.setUint32(row * NODE_STRIDE + NODE_COL_LABEL_OFF, labelCursor, true);
      nodeOut.setUint32(row * NODE_STRIDE + NODE_COL_LABEL_LEN, labelEncoded.length, true);
      labelBytesOut.set(labelEncoded, labelCursor);
      labelCursor += labelEncoded.length;

      const arrowBytes = decoded.tiltArrowCount * TILT_ARROW_STRIDE;
      if (arrowBytes > 0) {
        tiltArrowOut.set(
          new Uint8Array(decoded.tiltArrowView.buffer, decoded.tiltArrowView.byteOffset, arrowBytes),
          tiltArrowCursor,
        );
        tiltArrowCursor += arrowBytes;
      }

      const channelBytes = decoded.channelVectorCount * CHANNEL_VECTOR_STRIDE;
      if (channelBytes > 0) {
        channelVectorOut.set(
          new Uint8Array(decoded.channelVectorView.buffer, decoded.channelVectorView.byteOffset, channelBytes),
          channelVectorCursor,
        );
        channelVectorCursor += channelBytes;
      }
    }

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

  }

  return {
    tick: 0,
    nodeCount,
    nodeView: nodeOut,
    interiorCount,
    interiorView: new DataView(interiorBuf),
    labelBytesCount: totalLabelBytes,
    labelBytes: labelBytesOut,
    tiltArrowCount: totalTiltArrows,
    tiltArrowView: new DataView(tiltArrowBuf),
    channelVectorCount: totalChannelVectors,
    channelVectorView: new DataView(channelVectorBuf),
  };
}
