// stale-rows-on-scene-switch.test.ts — switching scenes must not leave the previous
// diagram on screen.
//
// THE BUG THIS PINS. A tab switch restarts Go against another topology tree. The extension
// host clears its own per-row caches at every spawn, but the webview is NOT reloaded by a
// respawn, so snapshot-buffer.ts's per-row maps survived it. Switching from a 9-node scene
// to a 2-node scene left rows 2..8 holding the old scene's frames — and since Go never
// emits for a row its current scene does not have, nothing ever overwrote them. Both
// diagrams rendered at once.
//
// The assertion is therefore about ROW REMOVAL, not row content: rows the new scene does
// not have must be GONE, not merely stale.

import { describe, it, expect, beforeEach } from "vitest";
import {
  setLatestViewFrame,
  setLatestNodeStreamFrame,
  setLatestEdgeStreamFrame,
  getLatestNodeStreamFrames,
  getLatestEdgeStreamFrames,
  resetSceneIdentityForTest,
} from "../src/webview/snapshot-buffer";
import { SCENE_TABS_HEADER_SIZE } from "../src/webview/three/buffer-decode";
import { BUF_VIEW_FRAME_HEADER_SIZE } from "../src/schema/frame-tags";
import { CAMERA_STRIDE, OVERLAY_STRIDE, SCENE_STRIDE } from "../src/schema/buffer-layout";

const ENC = new TextEncoder();

/** A VIEW frame reporting the Go-owned tab strip with `selected` showing. */
function viewFrameForScene(selected: number): ArrayBuffer {
  const names = ["ring", "pair"].map((n) => ENC.encode(n));
  const tabsLen = SCENE_TABS_HEADER_SIZE + names.reduce((a, b) => a + 2 + b.byteLength, 0);
  const blocksLen = BUF_VIEW_FRAME_HEADER_SIZE + CAMERA_STRIDE + OVERLAY_STRIDE + SCENE_STRIDE;
  const buf = new ArrayBuffer(blocksLen + tabsLen + 4);
  const dv = new DataView(buf);
  let off = blocksLen;
  dv.setUint16(off, names.length, true);
  dv.setUint16(off + 2, selected, true);
  off += SCENE_TABS_HEADER_SIZE;
  for (const b of names) {
    dv.setUint16(off, b.byteLength, true);
    off += 2;
    new Uint8Array(buf, off, b.byteLength).set(b);
    off += b.byteLength;
  }
  dv.setUint32(off, 0, true); // no events
  return buf;
}

const rowFrame = (row: number) => new Uint8Array([row]).buffer;

describe("switching scenes drops the previous scene's rows", () => {
  beforeEach(() => {
    resetSceneIdentityForTest();
    // Establish scene 0 as the identity the maps are held FOR, then fill it out as a
    // 9-node / 10-edge scene the way the live streams would.
    setLatestViewFrame(viewFrameForScene(0));
    for (let r = 0; r < 9; r++) setLatestNodeStreamFrame(r, rowFrame(r));
    for (let r = 0; r < 10; r++) setLatestEdgeStreamFrame(r, rowFrame(r));
  });

  it("removes every node row when the VIEW frame reports a different scene", () => {
    expect(getLatestNodeStreamFrames().size).toBe(9);
    setLatestViewFrame(viewFrameForScene(1));
    // The 2-node scene has not streamed anything yet: the map must be EMPTY, not still
    // holding rows 2..8 from the 9-node scene.
    expect(getLatestNodeStreamFrames().size).toBe(0);
  });

  it("removes every edge row too (the wires of the old diagram, not just its nodes)", () => {
    expect(getLatestEdgeStreamFrames().size).toBe(10);
    setLatestViewFrame(viewFrameForScene(1));
    expect(getLatestEdgeStreamFrames().size).toBe(0);
  });

  it("keeps rows when the scene is unchanged — a plain restart or a per-tick frame must not blank the diagram", () => {
    setLatestViewFrame(viewFrameForScene(0));
    setLatestViewFrame(viewFrameForScene(0));
    expect(getLatestNodeStreamFrames().size).toBe(9);
    expect(getLatestEdgeStreamFrames().size).toBe(10);
  });

  it("does not clear on the FIRST frame, which would discard the replay to a remounted webview", () => {
    // A webview that remounts is sent the host's cached frames, and the first VIEW frame
    // may arrive AFTER them. Establishing the identity must not throw those away.
    //
    // beforeEach left 9 node rows in place; forgetting the identity makes the next VIEW
    // frame the "first" one again, even though it names a different scene than the frames
    // on hand. Those 9 rows must survive — the count is unchanged, which is exactly the
    // claim: no clear ran.
    resetSceneIdentityForTest();
    setLatestViewFrame(viewFrameForScene(1));
    expect(getLatestNodeStreamFrames().size).toBe(9);
  });
});
