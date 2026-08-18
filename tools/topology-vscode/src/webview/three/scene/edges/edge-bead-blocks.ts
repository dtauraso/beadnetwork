import { columnBytes, columnVersion } from "../../../../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../../../../Buffer/column-owners";
import {
  COL_STREAM_EDGE_BEAD_X, COL_STREAM_EDGE_BEAD_Y, COL_STREAM_EDGE_BEAD_Z,
  COL_STREAM_EDGE_BEAD_VALUE, COL_STREAM_EDGE_BEAD_EDGE_ROW,
  COL_STREAM_EDGE_BEAD_RING_M0, COL_STREAM_EDGE_BEAD_RING_M1,
  COL_STREAM_EDGE_BEAD_RING_M2, COL_STREAM_EDGE_BEAD_RING_M3,
  COL_STREAM_EDGE_BEAD_RING_M4, COL_STREAM_EDGE_BEAD_RING_M5,
  COL_STREAM_EDGE_BEAD_RING_M6, COL_STREAM_EDGE_BEAD_RING_M7,
  COL_STREAM_EDGE_BEAD_RING_M8, COL_STREAM_EDGE_BEAD_RING_M9,
  COL_STREAM_EDGE_BEAD_RING_M10, COL_STREAM_EDGE_BEAD_RING_M11,
  COL_STREAM_EDGE_BEAD_RING_M12, COL_STREAM_EDGE_BEAD_RING_M13,
  COL_STREAM_EDGE_BEAD_RING_M14, COL_STREAM_EDGE_BEAD_RING_M15,
} from "../../../../../Buffer/column-streams-gen";

export interface EdgeBeadsAgg {
  positions: Float32Array;
  value: Int32Array;

  srcNodeRow: Int32Array;

  edgeRow: Int32Array;

  ringMatrix: Float32Array;

  count: number;
}

const RING_COLS = [
  COL_STREAM_EDGE_BEAD_RING_M0, COL_STREAM_EDGE_BEAD_RING_M1,
  COL_STREAM_EDGE_BEAD_RING_M2, COL_STREAM_EDGE_BEAD_RING_M3,
  COL_STREAM_EDGE_BEAD_RING_M4, COL_STREAM_EDGE_BEAD_RING_M5,
  COL_STREAM_EDGE_BEAD_RING_M6, COL_STREAM_EDGE_BEAD_RING_M7,
  COL_STREAM_EDGE_BEAD_RING_M8, COL_STREAM_EDGE_BEAD_RING_M9,
  COL_STREAM_EDGE_BEAD_RING_M10, COL_STREAM_EDGE_BEAD_RING_M11,
  COL_STREAM_EDGE_BEAD_RING_M12, COL_STREAM_EDGE_BEAD_RING_M13,
  COL_STREAM_EDGE_BEAD_RING_M14, COL_STREAM_EDGE_BEAD_RING_M15,
];

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

    if (!ys || !zs || !vs || !es) continue;
    if (ys.byteLength < count * 4 || zs.byteLength < count * 4) continue;
    if (vs.byteLength < count * 4 || es.byteLength < count * 4) continue;

    const rings = RING_COLS.map((c) => columnBytes(nodeColumn(row, c)));

    for (let i = 0; i < count; i++) {
      const o = i * 4;
      positions[b * 3] = xs.getFloat32(o, true);
      positions[b * 3 + 1] = ys.getFloat32(o, true);
      positions[b * 3 + 2] = zs.getFloat32(o, true);
      value[b] = vs.getInt32(o, true);
      edgeRow[b] = es.getInt32(o, true);
      srcNodeRow[b] = row;

      for (let m = 0; m < 16; m++) {
        const col = rings[m];
        ringMatrix[b * 16 + m] = col && col.byteLength >= o + 4 ? col.getFloat32(o, true) : 0;
      }
      b++;
    }
  }

  const count = b;
  lastVersion = v;
  lastAgg = { positions, value, srcNodeRow, edgeRow, ringMatrix, count };
  return lastAgg;
}
