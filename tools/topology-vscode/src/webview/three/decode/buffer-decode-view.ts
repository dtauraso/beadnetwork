// buffer-decode-view.ts — decodeViewFrame: the dedicated VIEW-stream frame, returning the
// Camera/Overlay/Scene blocks plus the scene-tab strip and this stream's own trailing
// EVENTS section.

import {
  CAMERA_STRIDE,
  OVERLAY_STRIDE,
  SCENE_STRIDE,
} from "../../../schema/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../../../schema/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

/** Byte width of the scene-tabs section header: [count:u16][selected:u16]. Always present
 *  on a VIEW frame, even for zero tabs (see Buffer.BuildSceneTabsSection). */
export const SCENE_TABS_HEADER_SIZE = 4;

// decodeSceneTabs decodes the Go-owned tab strip written by Buffer.BuildSceneTabsSection:
// [count:u16][selected:u16] then count × ([nameLen:u16][utf8 name]). Returns the end offset
// so the caller can continue to the events trailer. A truncated section decodes as no tabs
// rather than throwing — a short frame must not take the whole render path down with it.
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

/** Decoded view over a BUF_BLOCK_TAG_VIEW frame (see frame-tags.ts for its byte layout):
 *  [tick:u32] followed by the Camera, Overlay, and Scene blocks. */
export interface DecodedViewFrame {
  tick: number;
  cameraView: DataView;
  overlayView: DataView;
  sceneView: DataView;
  /** The Go-owned scene tab strip (nodes/Wiring/scene/scene_tabs.go): the labels to draw and
   *  which one is showing. Empty for an untabbed anchor, which is what makes the strip
   *  absent rather than a single dead tab. TS renders these; it never invents one. */
  sceneTabs: string[];
  sceneTabSelected: number;
  /** This VIEW stream's own trailing EVENTS section (camera/overlay/scene events —
   *  every other kind is decentralized to its own owner fd). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

// Single-entry memo — the view frame arrives on its own dedicated fd, decoded
// independently of every other stream.
let lastViewBuf: ArrayBuffer | null = null;
let lastDecodedView: DecodedViewFrame | null = null;

/**
 * Decode a BUF_BLOCK_TAG_VIEW frame ArrayBuffer (the dedicated view-fd stream) into
 * typed camera/overlay/scene views. Returns null if the buffer is too small to be a
 * valid view frame. Pure — no side effects, no store reads/writes. Views alias the
 * original buffer (zero-copy). Memoized on `buf`'s identity.
 */
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
