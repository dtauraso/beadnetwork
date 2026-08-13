import {
  CAMERA_STRIDE,
  OVERLAY_STRIDE,
  SCENE_STRIDE,
  BUF_LAYOUT_FINGERPRINT_HASH,
} from "../../../schema/buffer-layout/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../../../schema/buffer-layout/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

export const SCENE_TABS_HEADER_SIZE = 4;

// Reported ONCE per distinct layout: the condition holds for every frame until
// somebody rebuilds, and an unthrottled report per frame is what wedged the
// extension host the last time a skew happened.
const reportedLayoutSkews = new Set<number>();

function reportLayoutSkew(frameLayout: number): void {
  if (reportedLayoutSkews.has(frameLayout)) return;
  reportedLayoutSkews.add(frameLayout);

  const message =
    `buffer layout skew: Go is sending layout ${frameLayout}, this webview bundle was built for ` +
    `${BUF_LAYOUT_FINGERPRINT_HASH}. The bundle and the Go binary come from different commits — ` +
    `rebuild the webview (npm run build in tools/topology-vscode) and reload the window. ` +
    `Frames are being refused until then, so the scene will not update.`;

  if (typeof window === "undefined") {
    // eslint-disable-next-line no-console
    console.error(`[wirefold] buffer-layout-skew: ${message}`);
    return;
  }
  void import("../../log/post").then(({ postLog }) => {
    postLog("load-error", {
      reason: "buffer-layout-skew",
      message,
      frameLayout,
      bundleLayout: BUF_LAYOUT_FINGERPRINT_HASH,
    });
  });
}

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

  const hdr = new DataView(buf, 0, BUF_VIEW_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  // Refuse a frame from a buffer layout this bundle was not built for. Decoding
  // it would read its bytes into the wrong columns and draw the result, which is
  // the failure this check exists to end: a 4-byte header change once made every
  // node and bead vanish with nothing anywhere saying why.
  const frameLayout = hdr.getUint32(4, true);
  if (frameLayout !== BUF_LAYOUT_FINGERPRINT_HASH) {
    reportLayoutSkew(frameLayout);
    return null;
  }

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
