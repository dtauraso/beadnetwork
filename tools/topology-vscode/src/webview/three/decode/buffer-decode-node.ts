import { TILT_ARROW_STRIDE, CHANNEL_VECTOR_STRIDE } from "../../../../Buffer/buffer-layout";
import { BUF_NODE_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";
import { columnBytes } from "../../../../Buffer/column-values";
import { nodeColumn } from "../../../../Buffer/column-owners";
import { COL_STREAM_NODE_LABEL } from "../../../../Buffer/column-streams-gen";

export interface DecodedNodeStreamFrame {
  tick: number;

  tiltArrowCount: number;
  tiltArrowView: DataView;

  channelVectorCount: number;
  channelVectorView: DataView;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

let reportedShortFrame = false;

function reportShortNodeFrame(got: number, expected: number): void {
  if (reportedShortFrame) return;
  reportedShortFrame = true;

  const message =
    `node stream frame is ${got} bytes but this webview's layout needs ${expected}. ` +
    `Go and the webview are built against different buffer layouts, so EVERY node frame is ` +
    `being dropped and nothing on a node will update. Run "Developer: Reload Window" — ` +
    `reopening the file reloads only the webview, not the extension host that spawns Go.`;
  if (typeof window === "undefined") {
    // eslint-disable-next-line no-console
    console.error(`[wirefold] node-frame-layout-skew: ${message}`);
    return;
  }
  void import("../../log/post").then(({ postLog }) => {
    postLog("load-error", { reason: "node-frame-layout-skew", message, gotBytes: got, expectedBytes: expected });
  });
}

const lastNodeStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedNodeStreamByRow = new Map<number, DecodedNodeStreamFrame | null>();

export function decodeNodeStreamFrame(row: number, buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (lastNodeStreamBufByRow.get(row) === buf) {
    return lastDecodedNodeStreamByRow.get(row) ?? null;
  }
  const decoded = decodeNodeStreamFrameUncached(buf);
  lastNodeStreamBufByRow.set(row, buf);
  lastDecodedNodeStreamByRow.set(row, decoded);
  return decoded;
}

function decodeNodeStreamFrameUncached(buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (buf.byteLength < BUF_NODE_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_NODE_STREAM_FRAME_HEADER_SIZE);
  const tick               = hdr.getUint32(0,  true);
  const tiltArrowCount     = hdr.getUint32(8,  true);
  const channelVectorCount = hdr.getUint32(12, true);

  const expectedLen = BUF_NODE_STREAM_FRAME_HEADER_SIZE
    + tiltArrowCount * TILT_ARROW_STRIDE
    + channelVectorCount * CHANNEL_VECTOR_STRIDE;
  if (buf.byteLength < expectedLen) {
    reportShortNodeFrame(buf.byteLength, expectedLen);
    return null;
  }

  let off = BUF_NODE_STREAM_FRAME_HEADER_SIZE;
  const tiltArrowView = new DataView(buf, off, tiltArrowCount * TILT_ARROW_STRIDE);
  off += tiltArrowCount * TILT_ARROW_STRIDE;

  const channelVectorView = new DataView(buf, off, channelVectorCount * CHANNEL_VECTOR_STRIDE);
  off += channelVectorCount * CHANNEL_VECTOR_STRIDE;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, tiltArrowCount, tiltArrowView, channelVectorCount, channelVectorView, eventCount, eventView, eventTextView };
}

export function nodeLabel(row: number): string {
  if (row < 0) return "";
  const bytes = columnBytes(nodeColumn(row, COL_STREAM_NODE_LABEL));
  if (!bytes || bytes.byteLength === 0) return "";
  return STR_DECODER.decode(new Uint8Array(bytes.buffer, bytes.byteOffset, bytes.byteLength));
}
