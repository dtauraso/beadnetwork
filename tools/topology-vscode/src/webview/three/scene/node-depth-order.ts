let drawOrder: Int32Array | null = null;

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

  rows.sort((a, b) => distSq[b]! - distSq[a]!);
  return Int32Array.from(rows);
}

export function setNodeDrawOrder(order: Int32Array): void {
  drawOrder = order;
}

export function resolveNodeDrawSlot(slot: number): number {
  if (!drawOrder || slot < 0 || slot >= drawOrder.length) return slot;
  return drawOrder[slot]!;
}
