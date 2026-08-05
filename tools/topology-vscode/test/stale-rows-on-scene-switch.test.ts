// stale-rows-on-scene-switch.test.ts — one tab's streamed rows must be unreachable from
// the other.
//
// THE BUG THIS PINS. The per-row maps were keyed by ROW ALONE, so the ring's row 3 and the
// pair's row 3 were the same cell. Switching from a 9-node scene to a 2-node one left rows
// 2..8 holding the previous scene's frames, and both diagrams drew at once.
//
// The first fix cleared the maps on a scene change. That repaired the visible symptom and
// introduced a quieter one: INTERIOR bead frames are written only when a node's held value
// CHANGES, so a cleared interior row has nothing to re-send it — a Pulse holding a steady
// value simply loses its bead. Node geometry, re-sent every tick, hid the problem by
// repopulating instantly.
//
// So these assert ISOLATION, not removal: each scene has its own table, one tab cannot see
// or overwrite the other's rows, and switching away and back finds them intact.

import { describe, it, expect, beforeEach } from "vitest";
import {
  setLatestViewFrame,
  setLatestNodeStreamFrame,
  setLatestEdgeStreamFrame,
  setLatestInteriorStreamFrame,
  getLatestNodeStreamFrames,
  getLatestEdgeStreamFrames,
  getLatestInteriorStreamFrames,
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

describe("each scene owns its own row tables", () => {
  beforeEach(() => {
    resetSceneIdentityForTest();
    // Scene 0 (ring): a 9-node / 10-edge scene, filled out the way the live streams would.
    setLatestViewFrame(viewFrameForScene(0));
    for (let r = 0; r < 9; r++) setLatestNodeStreamFrame(r, rowFrame(r));
    for (let r = 0; r < 10; r++) setLatestEdgeStreamFrame(r, rowFrame(r));
  });

  it("shows none of the previous scene's node rows after a switch", () => {
    expect(getLatestNodeStreamFrames().size).toBe(9);
    setLatestViewFrame(viewFrameForScene(1));
    // The 2-node scene has streamed nothing yet, and must not inherit rows 0..8.
    expect(getLatestNodeStreamFrames().size).toBe(0);
  });

  it("shows none of the previous scene's edge rows either", () => {
    setLatestViewFrame(viewFrameForScene(1));
    expect(getLatestEdgeStreamFrames().size).toBe(0);
  });

  it("keeps each scene's rows when switching AWAY and BACK", () => {
    setLatestViewFrame(viewFrameForScene(1));
    setLatestNodeStreamFrame(0, rowFrame(100));
    setLatestNodeStreamFrame(1, rowFrame(101));
    expect(getLatestNodeStreamFrames().size).toBe(2);

    setLatestViewFrame(viewFrameForScene(0));
    // The ring's own nine rows are still there — NOT cleared by the excursion.
    expect(getLatestNodeStreamFrames().size).toBe(9);

    setLatestViewFrame(viewFrameForScene(1));
    expect(getLatestNodeStreamFrames().size).toBe(2);
  });

  // The reason isolation must be by KEY rather than by clearing: an interior frame is
  // written only when a node's held value changes, so a cleared interior row has nothing to
  // re-send it. Switching away and back must return the held bead, not an empty node.
  it("preserves INTERIOR rows across a switch, which a clear would silently destroy", () => {
    setLatestInteriorStreamFrame(0, rowFrame(7));
    expect(getLatestInteriorStreamFrames().size).toBe(1);

    setLatestViewFrame(viewFrameForScene(1));
    expect(getLatestInteriorStreamFrames().size).toBe(0); // the pair's own table, still empty

    setLatestViewFrame(viewFrameForScene(0));
    expect(getLatestInteriorStreamFrames().size).toBe(1); // the ring's held bead survived
  });

  it("writes arriving under one scene never land in the other's table", () => {
    setLatestViewFrame(viewFrameForScene(1));
    setLatestNodeStreamFrame(3, rowFrame(200));
    setLatestViewFrame(viewFrameForScene(0));
    const ringRow3 = getLatestNodeStreamFrames().get(3);
    expect(ringRow3).toBeDefined();
    // Row 3 of the ring is the ring's own frame, not the one written while the pair showed.
    expect(new Uint8Array(ringRow3!)[0]).toBe(3);
  });
});
