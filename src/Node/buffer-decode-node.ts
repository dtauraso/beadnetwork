import { BUF_NODE_STREAM_FRAME_HEADER_SIZE } from "../Buffer/frame-tags";
import { STR_DECODER, decodeTrailingEvents, type DecodedEvents } from "../webview/decode/buffer-decode-shared";
import { nodeBytes } from "./node-leaves";

export interface DecodedNodeStreamFrame {
  tick: number;

  events: DecodedEvents;
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
  void import("../webview/log/post").then(({ postLog }) => {
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
  const tick = hdr.getUint32(0, true);

  const events = decodeTrailingEvents(buf, BUF_NODE_STREAM_FRAME_HEADER_SIZE);

  return { tick, events };
}

export function nodeLabel(row: number): string {
  if (row < 0) return "";
  const bytes = nodeBytes(row, "label");
  if (!bytes || bytes.byteLength === 0) return "";
  return STR_DECODER.decode(new Uint8Array(bytes.buffer, bytes.byteOffset, bytes.byteLength));
}
