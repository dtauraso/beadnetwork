import {
  CAMERA_STRIDE,
  OVERLAY_STRIDE,
  SCENE_STRIDE,
} from "../../../schema/buffer-layout/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../../../schema/buffer-layout/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

export const SCENE_TABS_HEADER_SIZE = 4;

function decodeSceneTabs(buf: ArrayBuffer, offset: number): { names: string[]; selected: number; end: number } {
  if (buf.byteLength < offset + SCENE_TABS_HEADER_SIZE) {
    return { names: [], selected: 0, end: offset };
  }
  const head = new DataView(buf, offset, SCENE_TABS_HEADER_SIZE);
  const count = head.getUint16(0, true);
  const selected = head.getUint16(2, true);
  let off = offset + SCENE_TABS_HEADER_SIZE;
  const names: string[] = [];
  for (let i = 0; i < count; i++) {
    if (buf.byteLength < off + 2) return { names: [], selected: 0, end: off };
    const len = new DataView(buf, off, 2).getUint16(0, true);
    off += 2;
    if (buf.byteLength < off + len) return { names: [], selected: 0, end: off };
    names.push(STR_DECODER.decode(new Uint8Array(buf, off, len)));
    off += len;
  }
  return { names, selected, end: off };
}

export interface DecodedViewFrame {
  tick: number;
  cameraView: DataView;
  overlayView: DataView;
  sceneView: DataView;

  sceneTabs: string[];
  sceneTabSelected: number;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

let lastViewBuf: ArrayBuffer | null = null;
let lastDecodedView: DecodedViewFrame | null = null;

export function decodeViewFrame(buf: ArrayBuffer): DecodedViewFrame | null {
  if (buf === lastViewBuf) return lastDecodedView;
  const decoded = decodeViewFrameUncached(buf);
  lastViewBuf = buf;
  lastDecodedView = decoded;
  return decoded;
}

function decodeViewFrameUncached(buf: ArrayBuffer): DecodedViewFrame | null {
  const expectedLen = BUF_VIEW_FRAME_HEADER_SIZE + CAMERA_STRIDE + OVERLAY_STRIDE + SCENE_STRIDE + SCENE_TABS_HEADER_SIZE;
  if (buf.byteLength < expectedLen) return null;

  const tick = new DataView(buf, 0, BUF_VIEW_FRAME_HEADER_SIZE).getUint32(0, true);
  let off = BUF_VIEW_FRAME_HEADER_SIZE;

  const cameraView = new DataView(buf, off, CAMERA_STRIDE);
  off += CAMERA_STRIDE;

  const overlayView = new DataView(buf, off, OVERLAY_STRIDE);
  off += OVERLAY_STRIDE;

  const sceneView = new DataView(buf, off, SCENE_STRIDE);
  off += SCENE_STRIDE;

  const tabs = decodeSceneTabs(buf, off);
  off = tabs.end;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return {
    tick, cameraView, overlayView, sceneView,
    sceneTabs: tabs.names, sceneTabSelected: tabs.selected,
    eventCount, eventView, eventTextView,
  };
}
