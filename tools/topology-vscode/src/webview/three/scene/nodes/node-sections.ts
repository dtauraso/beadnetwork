import {
  getLatestNodeStreamFrames, getNodeStreamVersion, subscribeNodeStreamFrame,
} from "../../../snapshot-buffer";
import { decodeNodeStreamFrame } from "../../decode/buffer-decode-node";
import { TILT_ARROW_STRIDE, CHANNEL_VECTOR_STRIDE } from "../../../../../Buffer/buffer-layout";

export interface NodeSections {
  tiltArrowCount: number;
  tiltArrowView: DataView;

  channelVectorCount: number;
  channelVectorView: DataView;
}

let lastVersion = -1;
let cached: NodeSections | null = null;

export function subscribeNodeStreamBlocks(fn: () => void): () => void {
  return subscribeNodeStreamFrame(fn);
}

export function getNodeSections(): NodeSections | null {
  const frames = getLatestNodeStreamFrames();
  if (frames.size === 0) return null;
  const v = getNodeStreamVersion();
  if (v === lastVersion && cached) return cached;

  let maxRow = -1;
  for (const r of frames.keys()) if (r > maxRow) maxRow = r;

  let arrows = 0;
  let channels = 0;
  const decoded = new Map<number, ReturnType<typeof decodeNodeStreamFrame>>();
  for (let row = 0; row <= maxRow; row++) {
    const buf = frames.get(row);
    const d = buf ? decodeNodeStreamFrame(row, buf) : null;
    decoded.set(row, d);
    if (d) {
      arrows += d.tiltArrowCount;
      channels += d.channelVectorCount;
    }
  }

  const arrowBuf = new ArrayBuffer(arrows * TILT_ARROW_STRIDE);
  const arrowOut = new Uint8Array(arrowBuf);
  const channelBuf = new ArrayBuffer(channels * CHANNEL_VECTOR_STRIDE);
  const channelOut = new Uint8Array(channelBuf);
  let ac = 0;
  let cc = 0;
  for (let row = 0; row <= maxRow; row++) {
    const d = decoded.get(row) ?? null;
    if (!d) continue;
    const ab = d.tiltArrowCount * TILT_ARROW_STRIDE;
    if (ab > 0) {
      arrowOut.set(new Uint8Array(d.tiltArrowView.buffer, d.tiltArrowView.byteOffset, ab), ac);
      ac += ab;
    }
    const cb = d.channelVectorCount * CHANNEL_VECTOR_STRIDE;
    if (cb > 0) {
      channelOut.set(new Uint8Array(d.channelVectorView.buffer, d.channelVectorView.byteOffset, cb), cc);
      cc += cb;
    }
  }

  cached = {
    tiltArrowCount: arrows,
    tiltArrowView: new DataView(arrowBuf),
    channelVectorCount: channels,
    channelVectorView: new DataView(channelBuf),
  };
  lastVersion = v;
  return cached;
}
