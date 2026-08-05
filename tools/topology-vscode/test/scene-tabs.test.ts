// scene-tabs.test.ts — decoding the VIEW frame's Go-owned scene tab strip.
//
// The bytes under test are written by Buffer.BuildSceneTabsSection (Go); this asserts the
// TS half reads the same layout — [count:u16][selected:u16] then count × ([len:u16][utf8]).
// A mismatch here is the silent-misalignment class: the strip would render plausible
// garbage (or nothing) rather than erroring.

import { describe, it, expect } from "vitest";
import { decodeViewFrame, SCENE_TABS_HEADER_SIZE } from "../src/webview/three/buffer-decode";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../src/schema/frame-tags";
import { CAMERA_STRIDE, OVERLAY_STRIDE, SCENE_STRIDE } from "../src/schema/buffer-layout";

const ENC = new TextEncoder();

/** Build a VIEW frame carrying `names` with `selected`, plus an empty events trailer. */
function makeViewFrameWithTabs(names: string[], selected: number): ArrayBuffer {
  const nameBytes = names.map((n) => ENC.encode(n));
  const tabsLen = SCENE_TABS_HEADER_SIZE + nameBytes.reduce((a, b) => a + 2 + b.byteLength, 0);
  const blocksLen = BUF_VIEW_FRAME_HEADER_SIZE + CAMERA_STRIDE + OVERLAY_STRIDE + SCENE_STRIDE;
  // + 4 for the events section's [count:u32] (zero events).
  const buf = new ArrayBuffer(blocksLen + tabsLen + 4);
  const dv = new DataView(buf);

  let off = blocksLen;
  dv.setUint16(off, names.length, true);
  dv.setUint16(off + 2, selected, true);
  off += SCENE_TABS_HEADER_SIZE;
  for (const b of nameBytes) {
    dv.setUint16(off, b.byteLength, true);
    off += 2;
    new Uint8Array(buf, off, b.byteLength).set(b);
    off += b.byteLength;
  }
  dv.setUint32(off, 0, true); // events: none
  return buf;
}

describe("scene tab strip decode", () => {
  it("reads the Go-owned labels and which one is selected", () => {
    const decoded = decodeViewFrame(makeViewFrameWithTabs(["ring", "pair"], 1));
    expect(decoded).not.toBeNull();
    expect(decoded!.sceneTabs).toEqual(["ring", "pair"]);
    expect(decoded!.sceneTabSelected).toBe(1);
  });

  it("reads an empty strip for an untabbed anchor (the section is still present)", () => {
    const decoded = decodeViewFrame(makeViewFrameWithTabs([], 0));
    expect(decoded).not.toBeNull();
    expect(decoded!.sceneTabs).toEqual([]);
  });

  it("still finds the events trailer after a variable-length strip", () => {
    // The tabs section sits BEFORE the events section, so a wrong tabs width would leave
    // the events read misaligned — and this is the assertion that would catch it.
    const decoded = decodeViewFrame(makeViewFrameWithTabs(["ring", "pair"], 0));
    expect(decoded!.eventCount).toBe(0);
  });
});
