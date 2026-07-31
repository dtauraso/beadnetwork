// Unit tests for node-depth-order.ts — the per-frame back-to-front draw-order permutation
// that fixes node 2 drawing behind node 3 regardless of camera distance (transparent
// depthWrite=false InstancedMeshes only sort BETWEEN objects, not instances within one mesh —
// see node-depth-order.ts's header comment). Tests the ORDERING function itself (positions +
// camera -> sorted indices) and the instanceId -> nodeRow resolution used by the pick path —
// not three.js rendering (docs/testing-shape.md: no canvas, no renderer).

import { describe, it, expect } from "vitest";
import { computeNodeDepthOrder, setNodeDrawOrder, resolveNodeDrawSlot } from "../src/webview/three/node-depth-order";

describe("computeNodeDepthOrder", () => {
  it("sorts three colinear nodes far-to-near from the camera", () => {
    // Nodes at x=0,10,20; camera at x=30 looking back along -x, so node 0 (x=0) is
    // farthest and node 2 (x=20) is nearest. Expect order = [0, 1, 2] (farthest drawn
    // first, nearest drawn last so it wins the pixel under depthWrite=false).
    const xs = [0, 10, 20];
    const order = computeNodeDepthOrder(
      3,
      (row) => xs[row]!,
      () => 0,
      () => 0,
      30, 0, 0,
    );
    expect(Array.from(order)).toEqual([0, 1, 2]);
  });

  it("reproduces the reported bug scenario: node 2 nearer the camera than node 3", () => {
    // node 2 at x=5 (near camera at x=0), node 3 at x=50 (far). Row order (2 then 3, as
    // buffer rows) would draw node 3 last and cover node 2 even though node 2 is nearer —
    // the depth-sorted order must put node 2 (nearer) LAST so it wins.
    const rows = [2, 3]; // buffer row identity of each array slot, for readability only
    const xs = [5, 50]; // parallel to `rows`: xs[0] is row 2's x, xs[1] is row 3's x
    void rows;
    const order = computeNodeDepthOrder(
      2,
      (row) => xs[row]!,
      () => 0,
      () => 0,
      0, 0, 0,
    );
    // slot 0 -> row 1 (x=50, far), slot 1 -> row 0 (x=5, near) drawn last.
    expect(Array.from(order)).toEqual([1, 0]);
  });

  it("is stable under an all-equal-distance tie (no crash, every row present exactly once)", () => {
    const order = computeNodeDepthOrder(
      4,
      () => 1, () => 1, () => 1,
      0, 0, 0,
    );
    expect(Array.from(order).sort((a, b) => a - b)).toEqual([0, 1, 2, 3]);
  });

  it("handles n=0", () => {
    const order = computeNodeDepthOrder(0, () => 0, () => 0, () => 0, 0, 0, 0);
    expect(order.length).toBe(0);
  });
});

describe("resolveNodeDrawSlot (pick-path permutation round-trip)", () => {
  it("round-trips drawSlot -> nodeRow -> drawSlot through a published order", () => {
    // order[drawSlot] = nodeRow. A pick reports instanceId=drawSlot; resolveNodeDrawSlot
    // must return the nodeRow that instance was actually drawn from this frame.
    const order = new Int32Array([2, 0, 1]); // slot 0 drew row 2, slot 1 drew row 0, slot 2 drew row 1
    setNodeDrawOrder(order);
    expect(resolveNodeDrawSlot(0)).toBe(2);
    expect(resolveNodeDrawSlot(1)).toBe(0);
    expect(resolveNodeDrawSlot(2)).toBe(1);
    // Round-trip: for every drawSlot, find it again via the row it maps to.
    for (let slot = 0; slot < order.length; slot++) {
      const row = resolveNodeDrawSlot(slot);
      expect(Array.from(order).indexOf(row)).toBe(slot);
    }
  });

  it("falls back to identity for an out-of-range slot", () => {
    setNodeDrawOrder(new Int32Array([1, 0]));
    expect(resolveNodeDrawSlot(5)).toBe(5);
    expect(resolveNodeDrawSlot(-1)).toBe(-1);
  });
});
