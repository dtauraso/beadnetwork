import { columnBytes, columnVersion } from "../../../../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../../../../Buffer/column-owners";
import {
  COL_STREAM_EDGE_BEAD_X, COL_STREAM_EDGE_BEAD_Y, COL_STREAM_EDGE_BEAD_Z,
  COL_STREAM_EDGE_BEAD_VALUE, COL_STREAM_EDGE_BEAD_EDGE_ROW,
  COL_STREAM_EDGE_BEAD_RING_MATRIX,
} from "../../../../../Buffer/column-streams-gen";

export interface EdgeBeadsAgg {
  positions: Float32Array;
  value: Int32Array;

  srcNodeRow: Int32Array;

  edgeRow: Int32Array;

  ringMatrix: Float32Array;

  count: number;
}

let lastVersion = -1;
let lastAgg: EdgeBeadsAgg | null = null;

export function getEdgeBeads(): EdgeBeadsAgg {
  const v = columnVersion();
  if (lastAgg !== null && v === lastVersion) return lastAgg;

  const { nodes } = ownerCounts();
  const perNode: Array<{ row: number; count: number }> = [];
  let total = 0;
  for (let row = 0; row < nodes; row++) {
    const xs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_X));
    if (!xs) continue;
    const count = xs.byteLength >> 2;
    if (count === 0) continue;
    perNode.push({ row, count });
    total += count;
  }

  const positions = new Float32Array(total * 3);
  const ringMatrix = new Float32Array(total * 16);
  const value = new Int32Array(total);
  const srcNodeRow = new Int32Array(total);
  const edgeRow = new Int32Array(total);

  let b = 0;
  for (const { row, count } of perNode) {
    const xs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_X))!;
    const ys = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_Y));
    const zs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_Z));
    const vs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_VALUE));
    const es = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_EDGE_ROW));
    const rings = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_RING_MATRIX));

    if (!ys || !zs || !vs || !es || !rings) continue;
    if (ys.byteLength < count * 4 || zs.byteLength < count * 4) continue;
    if (vs.byteLength < count * 4 || es.byteLength < count * 4) continue;
    if (rings.byteLength < count * 64) continue;

    for (let i = 0; i < count; i++) {
      const o = i * 4;
      positions[b * 3] = xs.getFloat32(o, true);
      positions[b * 3 + 1] = ys.getFloat32(o, true);
      positions[b * 3 + 2] = zs.getFloat32(o, true);
      value[b] = vs.getInt32(o, true);
      edgeRow[b] = es.getInt32(o, true);
      srcNodeRow[b] = row;

      for (let m = 0; m < 16; m++) {
        ringMatrix[b * 16 + m] = rings.getFloat32(i * 64 + m * 4, true);
      }
      b++;
    }
  }

  const count = b;
  lastVersion = v;
  lastAgg = { positions, value, srcNodeRow, edgeRow, ringMatrix, count };
  return lastAgg;
}
