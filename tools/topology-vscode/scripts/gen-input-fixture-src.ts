


















import { encodeRawInput } from "../src/schema/input-encode";
import { INPUT_LAYOUT_FINGERPRINT } from "../src/schema/input-layout-gen";
import type { RawInputEvent } from "../src/messages";
import * as fs from "fs";

const events: RawInputEvent[] = [

  {
    kind: "pointerdown",
    x: 101.5, y: 102.25, rectLeft: 103.125, rectTop: 104.0625, rectWidth: 105.5, rectHeight: 106.25,
    button: 0, ctrl: true, shift: false, alt: true, meta: false,
    deltaX: 107.5, deltaY: 108.25, fov: 45.5,
    hit: { kind: "port", isInput: true, nodeRow: 11, portRow: 12, edgeRow: 13 },
  },


  {
    kind: "pointermove",
    x: 201.75, y: 202.125, rectLeft: 203.0625, rectTop: 204.5, rectWidth: 205.25, rectHeight: 206.375,
    button: -1, ctrl: false, shift: true, alt: false, meta: true,
    deltaX: 207.125, deltaY: 208.0625, fov: 60.25,
    hit: { kind: "node", isInput: false, nodeRow: -1, portRow: 21, edgeRow: 22 },
  },

  {
    kind: "pointerup",
    x: 301.0625, y: 302.75, rectLeft: 303.125, rectTop: 304.25, rectWidth: 305.375, rectHeight: 306.5,
    button: 2, ctrl: true, shift: true, alt: false, meta: false,
    deltaX: 307.25, deltaY: 308.125, fov: 75.75,
    hit: { kind: "edge", isInput: true, nodeRow: 31, portRow: -1, edgeRow: 32 },
  },

  {
    kind: "wheel",
    x: 401.5, y: 402.25, rectLeft: 403.0625, rectTop: 404.125, rectWidth: 405.75, rectHeight: 406.375,
    button: -1, ctrl: false, shift: false, alt: true, meta: true,
    deltaX: -407.5, deltaY: 408.625, fov: 30.125,
    hit: { kind: "torus", isInput: false, nodeRow: 41, portRow: 42, edgeRow: -1 },
  },

  {
    kind: "pointermove",
    x: 501.125, y: 502.875, rectLeft: 503.25, rectTop: 504.0625, rectWidth: 505.5, rectHeight: 506.75,
    button: 1, ctrl: true, shift: false, alt: false, meta: true,
    deltaX: 507.0625, deltaY: -508.5, fov: 90.5,
    hit: { kind: "handhold", isInput: true, nodeRow: 51, portRow: 52, edgeRow: 53 },
  },

  {
    kind: "pointerup",
    x: 601.375, y: 602.5, rectLeft: 603.625, rectTop: 604.75, rectWidth: 605.0625, rectHeight: 606.125,
    button: 0, ctrl: false, shift: true, alt: true, meta: false,
    deltaX: 607.5, deltaY: 608.25, fov: 15.5,
    hit: { kind: "empty", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 },
  },

  {
    kind: "home",
    x: 701.25, y: 702.5, rectLeft: 703.75, rectTop: 704.125, rectWidth: 705.375, rectHeight: 706.0625,
    button: 3, ctrl: true, shift: true, alt: true, meta: true,
    deltaX: 707.875, deltaY: 708.625, fov: 100.75,
    hit: { kind: "empty", isInput: true, nodeRow: -1, portRow: -1, edgeRow: -1 },
  },
];

function toHex(buf: ArrayBuffer): string {
  return Array.from(new Uint8Array(buf))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

export function generate(): { fingerprint: string; records: { event: RawInputEvent; hex: string }[] } {
  return {
    fingerprint: INPUT_LAYOUT_FINGERPRINT,
    records: events.map((event) => ({ event, hex: toHex(encodeRawInput(event)) })),
  };
}



const outPath = process.argv[2];
if (outPath) {
  fs.writeFileSync(outPath, JSON.stringify(generate(), null, 2) + "\n");
}
