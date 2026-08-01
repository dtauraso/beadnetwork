// Unit tests for material-transparent-edge-trigger.ts — the pure edge-trigger decision logic
// behind the polarVectors-overlay fade materials' `transparent` flip. Tests the DECISION
// (reassign+recompile vs. do nothing) against a plain object standing in for a
// THREE.Material — not three.js's own material/shader behaviour, which needs a WebGL
// context and is out of scope here (docs/testing-shape.md).

import { describe, it, expect } from "vitest";
import {
  createTransparentEdgeTrigger,
  applyTransparentEdgeTriggered,
} from "../src/webview/three/material-transparent-edge-trigger";

function fakeMaterial(initialTransparent: boolean) {
  return { transparent: initialTransparent, needsUpdate: false };
}

describe("applyTransparentEdgeTriggered", () => {
  it("reassigns and recompiles on the first call with a value", () => {
    const trigger = createTransparentEdgeTrigger();
    const mat = fakeMaterial(false);
    const changed = applyTransparentEdgeTriggered(trigger, mat, true);
    expect(changed).toBe(true);
    expect(mat.transparent).toBe(true);
    expect(mat.needsUpdate).toBe(true);
  });

  it("does nothing on repeat calls with the same value", () => {
    const trigger = createTransparentEdgeTrigger();
    const mat = fakeMaterial(false);
    applyTransparentEdgeTriggered(trigger, mat, true);
    mat.needsUpdate = false; // simulate three.js consuming the flag after the recompile
    const changed = applyTransparentEdgeTriggered(trigger, mat, true);
    expect(changed).toBe(false);
    expect(mat.needsUpdate).toBe(false);
    expect(mat.transparent).toBe(true);
  });

  it("reports change again when the value flips back", () => {
    const trigger = createTransparentEdgeTrigger();
    const mat = fakeMaterial(false);
    applyTransparentEdgeTriggered(trigger, mat, true);
    mat.needsUpdate = false;
    const changedBack = applyTransparentEdgeTriggered(trigger, mat, false);
    expect(changedBack).toBe(true);
    expect(mat.transparent).toBe(false);
    expect(mat.needsUpdate).toBe(true);
  });

  it("does not force needsUpdate on a false->false steady state either", () => {
    const trigger = createTransparentEdgeTrigger();
    const mat = fakeMaterial(false);
    const changed = applyTransparentEdgeTriggered(trigger, mat, false);
    // First call still counts as a change from "unknown" (prev=null) to false.
    expect(changed).toBe(true);
    mat.needsUpdate = false;
    const steady = applyTransparentEdgeTriggered(trigger, mat, false);
    expect(steady).toBe(false);
    expect(mat.needsUpdate).toBe(false);
  });
});
