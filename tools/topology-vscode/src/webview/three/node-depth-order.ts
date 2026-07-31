// node-depth-order.ts — per-frame back-to-front draw order for NodeInstances.
//
// WHY this exists: the node body/ring/ring-pick InstancedMeshes are transparent with
// depthWrite=false (glass look, buffer-scene-shared.ts / NodeInstances.tsx) — three.js can
// sort separate OBJECTS against each other by depth, but it draws the instances WITHIN one
// InstancedMesh in index order, which here is buffer row order (node id). With nothing
// writing depth, whichever instance paints LAST wins the pixel, so a higher row always
// covered a lower one regardless of camera distance (a node with a small id drew behind one
// with a larger id even when it was nearer the camera). The fix is to choose the per-frame
// INDEX ORDER ourselves: sort node rows far-to-near from the live camera and draw in that
// order, so the nearest node paints last and wins.
//
// This is pure rendering derived state, not domain state (MODEL.md drift rule /
// check-no-webview-state.sh): it is recomputed every frame from the streamed node positions
// and the live camera, never persisted, and never fed back into Go. It is stored at module
// scope (mirroring node-stream-blocks.ts's getNodeFrame()/view-blocks.ts's
// getViewBlocks() pattern) purely so the pick path (scene-content.tsx) can resolve a hit's
// instanceId — which after sorting is a DRAW SLOT, not a node row — back to the node row it
// actually landed on, without routing the order through React state or props drilling.

let drawOrder: Int32Array | null = null;

/**
 * Sort `n` node rows back-to-front against the camera and return the permutation:
 * order[drawSlot] = nodeRow. NodeInstances writes instance `drawSlot` (not `row`) using each
 * row's own transform/color, so instance draw order becomes far-to-near and the nearest node
 * paints last (wins the pixel under depthWrite=false). Pure function of the inputs — no
 * three.js Camera object required, just its world position, so this is testable without a
 * renderer.
 */
export function computeNodeDepthOrder(
  n: number,
  nodeX: (row: number) => number,
  nodeY: (row: number) => number,
  nodeZ: (row: number) => number,
  camX: number,
  camY: number,
  camZ: number,
): Int32Array {
  const rows = new Array<number>(n);
  const distSq = new Float64Array(n);
  for (let row = 0; row < n; row++) {
    rows[row] = row;
    const dx = nodeX(row) - camX;
    const dy = nodeY(row) - camY;
    const dz = nodeZ(row) - camZ;
    distSq[row] = dx * dx + dy * dy + dz * dz;
  }
  // Farthest first (largest distSq) so the sort is back-to-front and the nearest node's
  // instance is written/drawn LAST.
  rows.sort((a, b) => distSq[b]! - distSq[a]!);
  return Int32Array.from(rows);
}

/** Published each frame by NodeInstances' useFrame, alongside the instance matrices it
 *  computes the order from — read by the pick path (scene-content.tsx) to turn a hit's
 *  instanceId (a draw slot once instances are reordered) back into a node row. */
export function setNodeDrawOrder(order: Int32Array): void {
  drawOrder = order;
}

/** Resolve a draw-slot instanceId back to its node row via the last-published order. Returns
 *  the slot unchanged if no order has been published yet (e.g. a pick before the first
 *  useFrame ran) — identity is the correct fallback since row and slot coincide until the
 *  first sort. */
export function resolveNodeDrawSlot(slot: number): number {
  if (!drawOrder || slot < 0 || slot >= drawOrder.length) return slot;
  return drawOrder[slot]!;
}
