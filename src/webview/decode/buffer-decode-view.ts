import { BUF_LAYOUT_FINGERPRINT_HASH } from "../../Buffer/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../../Buffer/frame-tags";
import { decodeTrailingEvents, type DecodedEvents } from "./buffer-decode-shared";

const reportedLayoutSkews = new Set<number>();

function reportLayoutSkew(frameLayout: number): void {
  if (reportedLayoutSkews.has(frameLayout)) return;
  reportedLayoutSkews.add(frameLayout);

  const message =
    `buffer layout skew: Go is sending layout ${frameLayout}, this webview bundle was built for ` +
    `${BUF_LAYOUT_FINGERPRINT_HASH}. The bundle and the Go binary come from different commits — ` +
    `rebuild the webview (npm run build) and reload the window. ` +
    `Frames are being refused until then, so the scene will not update.`;

  if (typeof window === "undefined") {
    // eslint-disable-next-line no-console
    console.error(`[wirefold] buffer-layout-skew: ${message}`);
    return;
  }
  void import("../log/post").then(({ postLog }) => {
    postLog("load-error", {
      reason: "buffer-layout-skew",
      message,
      frameLayout,
      bundleLayout: BUF_LAYOUT_FINGERPRINT_HASH,
    });
  });
}

export interface DecodedViewFrame {
  tick: number;

  events: DecodedEvents;
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
  if (buf.byteLength < BUF_VIEW_FRAME_HEADER_SIZE) return null;

  const hdr = new DataView(buf, 0, BUF_VIEW_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  const frameLayout = hdr.getUint32(4, true);
  if (frameLayout !== BUF_LAYOUT_FINGERPRINT_HASH) {
    reportLayoutSkew(frameLayout);
    return null;
  }

  const events = decodeTrailingEvents(buf, BUF_VIEW_FRAME_HEADER_SIZE);

  return { tick, events };
}
