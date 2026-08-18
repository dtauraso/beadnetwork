import { EDGE_BEAD_STRIDE } from "../../../../../../Buffer/buffer-layout";
import { BUF_BEAD_STREAM_FRAME_HEADER_SIZE } from "../../../../../../Buffer/frame-tags";
import { decodeTrailingEvents } from "./buffer-decode-shared";

export interface DecodedBeadStreamFrame {
  tick: number;

  nodeRow: number;

  beadCount: number;
  beadView: DataView;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastBeadStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedBeadStreamByRow = new Map<number, DecodedBeadStreamFrame | null>();

export function decodeBeadStreamFrame(row: number, buf: ArrayBuffer): DecodedBeadStreamFrame | null {
  if (lastBeadStreamBufByRow.get(row) === buf) {
    return lastDecodedBeadStreamByRow.get(row) ?? null;
  }
  const decoded = decodeBeadStreamFrameUncached(buf);
  lastBeadStreamBufByRow.set(row, buf);
  lastDecodedBeadStreamByRow.set(row, decoded);
  return decoded;
}

function decodeBeadStreamFrameUncached(buf: ArrayBuffer): DecodedBeadStreamFrame | null {
  if (buf.byteLength < BUF_BEAD_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_BEAD_STREAM_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);
  const nodeRow = hdr.getInt32(4, true);
  const beadCount = hdr.getUint32(8, true);

  let off = BUF_BEAD_STREAM_FRAME_HEADER_SIZE;
  const beadBytes = beadCount * EDGE_BEAD_STRIDE;
  if (buf.byteLength < off + beadBytes) return null;
  const beadView = new DataView(buf, off, beadBytes);
  off += beadBytes;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, nodeRow, beadCount, beadView, eventCount, eventView, eventTextView };
}
