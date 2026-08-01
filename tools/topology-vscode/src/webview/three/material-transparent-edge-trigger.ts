// material-transparent-edge-trigger.ts — edge-triggered flip of THREE.Material.transparent.
//
// WHY this exists: in three.js `transparent` is part of a material's PROGRAM KEY — flipping
// it at runtime requires `material.needsUpdate = true` so the shader/render-state gets
// rebuilt, or the flip applies inconsistently (NodeInstances.tsx/ChainBeadInstances.tsx's
// polarVectors-overlay fade bug: the diagram looked different on first load than after any
// toggle). But the fade materials are flipped from a per-frame useFrame, and setting
// needsUpdate unconditionally there would force a shader recompile 60x/second. The fix is to
// track the PREVIOUS value per material and only reassign + set needsUpdate when the desired
// value actually differs from it — a plain edge trigger, decoupled from three.js so it is
// unit-testable without a WebGL context (docs/testing-shape.md: three.js's own material
// behaviour is not testable here, this pure decision logic is).

/** Per-material previous-value tracker. One instance per material ref. */
export interface TransparentEdgeTrigger {
  /** Last value assigned to `transparent`, or null if never assigned yet. */
  prev: boolean | null;
}

export function createTransparentEdgeTrigger(): TransparentEdgeTrigger {
  return { prev: null };
}

/**
 * Apply `desired` to a material's `transparent` flag only on change (including the first
 * call), setting `needsUpdate = true` alongside it so three.js rebuilds the program for the
 * new value. Returns true if the material was reassigned/recompiled this call, false if
 * nothing changed (the common case, most frames). `material` is typed structurally so this
 * stays independent of any specific THREE.Material subclass and is callable with a plain
 * object from a unit test.
 */
export function applyTransparentEdgeTriggered(
  trigger: TransparentEdgeTrigger,
  material: { transparent: boolean; needsUpdate: boolean },
  desired: boolean,
): boolean {
  if (trigger.prev === desired) return false;
  material.transparent = desired;
  material.needsUpdate = true;
  trigger.prev = desired;
  return true;
}
