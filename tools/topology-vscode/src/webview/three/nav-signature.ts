// nav-signature.ts — navSignature: coarse fingerprint of the buffer-derived nav nodes
// (rounded positions/radii/sphereR/selection). NavGuides bumps a render tick only when this
// changes, so the tori/frames rebuild on real position/selection changes (a drag) without
// per-frame churn — re-rendering on node-geometry store events. Not used on the flag-off path.

import { type NavNode } from "./buffer-nav";

export function navSignature(nav: NavNode[]): string {
  let s = "";
  for (const n of nav) {
    s += `${n.row}:${Math.round(n.center.x)},${Math.round(n.center.y)},${Math.round(n.center.z)},${Math.round(n.radius)},${Math.round(n.sphereR ?? 0)},${n.selected ? 1 : 0},${n.latchedSel ? 1 : 0};`;
  }
  return s;
}
