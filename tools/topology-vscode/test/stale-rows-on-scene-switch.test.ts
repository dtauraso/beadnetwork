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
// So these assert ISOLATION, not removal. The table is chosen by the frame's own SPAWN
// GENERATION, stamped by the host before the process writes a byte — routing is wired, not
// inferred from arrival order. A tab switch restarts Go, so a new generation is exactly what
// separates one tab's rows from another's.

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

// Generation 1 stands for the process showing the ring; generation 2 for the one showing the
// pair after a switch. The host bumps this per spawn; the webview never derives it.
const RING = 1;
const PAIR = 2;

describe("each process generation owns its own row tables", () => {
  beforeEach(() => {
    resetSceneIdentityForTest();
    setLatestViewFrame(viewFrameForScene(0), RING);
    for (let r = 0; r < 9; r++) setLatestNodeStreamFrame(r, rowFrame(r), RING);
    for (let r = 0; r < 10; r++) setLatestEdgeStreamFrame(r, rowFrame(r), RING);
  });

  it("shows none of the previous process's node rows once the new one streams", () => {
    expect(getLatestNodeStreamFrames().size).toBe(9);
    setLatestNodeStreamFrame(0, rowFrame(100), PAIR);
    // The 2-node scene streamed one row; rows 1..8 of the ring must not be inherited.
    expect(getLatestNodeStreamFrames().size).toBe(1);
  });

  it("shows none of the previous process's edge rows either", () => {
    setLatestEdgeStreamFrame(0, rowFrame(100), PAIR);
    expect(getLatestEdgeStreamFrames().size).toBe(1);
  });

  // The reason isolation must be by KEY rather than by clearing: an interior frame is
  // written only when a node's held value changes, so a cleared interior row has nothing to
  // re-send it. A held bead must survive anything short of its own process ending.
  it("never destroys an interior row that its own process still owns", () => {
    setLatestInteriorStreamFrame(0, rowFrame(7), RING);
    expect(getLatestInteriorStreamFrames().size).toBe(1);
    // More ring traffic of every kind — none of it may disturb the held interior row.
    for (let r = 0; r < 9; r++) setLatestNodeStreamFrame(r, rowFrame(r), RING);
    setLatestViewFrame(viewFrameForScene(0), RING);
    expect(getLatestInteriorStreamFrames().size).toBe(1);
    expect(new Uint8Array(getLatestInteriorStreamFrames().get(0)!)[0]).toBe(7);
  });

  // THE RACE THIS DESIGN REMOVES. A frame still in flight from the outgoing process, landing
  // after the new one has started, must not appear in the new process's table — and must not
  // evict what the new process already wrote.
  it("files a late frame from the OLD process away from the new one's table", () => {
    setLatestNodeStreamFrame(0, rowFrame(100), PAIR);
    setLatestNodeStreamFrame(3, rowFrame(3), RING); // arrives late, from the dead process
    const live = getLatestNodeStreamFrames();
    expect(live.size).toBe(1);
    expect(live.has(3)).toBe(false);
    expect(new Uint8Array(live.get(0)!)[0]).toBe(100);
  });

  // Ring -> pair -> ring is THREE processes, and the third gets its own table rather than
  // reusing the first's. A fresh process re-emits what it has, so nothing stale survives.
  it("gives a revisited scene a fresh table, not the previous visit's", () => {
    setLatestNodeStreamFrame(0, rowFrame(100), PAIR);
    setLatestNodeStreamFrame(0, rowFrame(200), 3); // ring again, a third spawn
    const live = getLatestNodeStreamFrames();
    expect(live.size).toBe(1);
    expect(new Uint8Array(live.get(0)!)[0]).toBe(200);
  });
});
