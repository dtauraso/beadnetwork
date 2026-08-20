import { decodeViewFrame } from "../webview/decode/buffer-decode-view";
import { logfmt } from "./probe/logfmt";
import { decodeEventLines } from "../webview/decode/decode-event-line";
import type { DecodedEvents } from "../webview/decode/buffer-decode-shared";

export function decodeBufferLog(viewFrameBuf: ArrayBuffer, breadcrumbsOnly = false): string {
  const dv = decodeViewFrame(viewFrameBuf);
  if (!dv) return "";
  return decodeStreamFrameEvents(dv.events, breadcrumbsOnly);
}

export function decodeStreamFrameEvents(events: DecodedEvents, breadcrumbsOnly = false): string {
  const now = Date.now();
  let out = "";
  for (const line of decodeEventLines(events, breadcrumbsOnly)) {
    out += logfmt({ ts_ms: now, src: "go", ...line }) + "\n";
  }
  return out;
}
