import {
  NODE_RING_POINT_STRIDE,
  BEAD_RING_POINT_STRIDE,
  BUF_LAYOUT_FINGERPRINT_HASH,
} from "../../../../Buffer/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import {
  SHADING_PARAM_NODE_RING_SURFACE_NU,
  SHADING_PARAM_NODE_RING_SURFACE_NV,
  SHADING_PARAM_BEAD_RING_SURFACE_NU,
  SHADING_PARAM_BEAD_RING_SURFACE_NV,
} from "../../../../Buffer/shading-params";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

const RING_SURFACE_POINT_COUNT = SHADING_PARAM_NODE_RING_SURFACE_NU * SHADING_PARAM_NODE_RING_SURFACE_NV;
const RING_SURFACE_STRIDE = RING_SURFACE_POINT_COUNT * NODE_RING_POINT_STRIDE;

const BEAD_RING_SURFACE_POINT_COUNT = SHADING_PARAM_BEAD_RING_SURFACE_NU * SHADING_PARAM_BEAD_RING_SURFACE_NV;
const BEAD_RING_SURFACE_STRIDE = BEAD_RING_SURFACE_POINT_COUNT * BEAD_RING_POINT_STRIDE;

export const SCENE_TABS_HEADER_SIZE = 4;

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

  ringSurfacePointsView: DataView;
  beadRingSurfacePointsView: DataView;

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
  const expectedLen = BUF_VIEW_FRAME_HEADER_SIZE + RING_SURFACE_STRIDE + BEAD_RING_SURFACE_STRIDE + SCENE_TABS_HEADER_SIZE;
  if (buf.byteLength < expectedLen) return null;

  const hdr = new DataView(buf, 0, BUF_VIEW_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  const frameLayout = hdr.getUint32(4, true);
  if (frameLayout !== BUF_LAYOUT_FINGERPRINT_HASH) {
    reportLayoutSkew(frameLayout);
    return null;
  }

  let off = BUF_VIEW_FRAME_HEADER_SIZE;





  const ringSurfacePointsView = new DataView(buf, off, RING_SURFACE_STRIDE);
  off += RING_SURFACE_STRIDE;

  const beadRingSurfacePointsView = new DataView(buf, off, BEAD_RING_SURFACE_STRIDE);
  off += BEAD_RING_SURFACE_STRIDE;

  const tabs = decodeSceneTabs(buf, off);
  off = tabs.end;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return {
    tick,
    ringSurfacePointsView, beadRingSurfacePointsView,
    sceneTabs: tabs.names, sceneTabSelected: tabs.selected,
    eventCount, eventView, eventTextView,
  };
}
