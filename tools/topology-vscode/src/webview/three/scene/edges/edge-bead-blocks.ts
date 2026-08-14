import {
  getLatestEdgeStreamFrames, getEdgeStreamVersion,
  getNodeStreamVersion,
} from "../../../snapshot-buffer";
import { decodeEdgeStreamFrame } from "../../decode/buffer-decode-edge";
import {
  readEdgeBeadX, readEdgeBeadY, readEdgeBeadZ, readEdgeBeadValue,
  readEdgeSrcNodeRow,
  readEdgeBeadRingM0, readEdgeBeadRingM1, readEdgeBeadRingM2, readEdgeBeadRingM3,
  readEdgeBeadRingM4, readEdgeBeadRingM5, readEdgeBeadRingM6, readEdgeBeadRingM7,
  readEdgeBeadRingM8, readEdgeBeadRingM9, readEdgeBeadRingM10, readEdgeBeadRingM11,
  readEdgeBeadRingM12, readEdgeBeadRingM13, readEdgeBeadRingM14, readEdgeBeadRingM15,
} from "../../../../schema/buffer-layout/buffer-layout";

export interface EdgeBeadsAgg {
  positions: Float32Array;
  value: Int32Array;

  srcNodeRow: Int32Array;

  ringMatrix: Float32Array;

  count: number;
}

let lastVersion = -1;
let lastNodeVersion = -1;
let lastAgg: EdgeBeadsAgg | null = null;

export function getEdgeBeads(): EdgeBeadsAgg {
  const ev = getEdgeStreamVersion();
  const nv = getNodeStreamVersion();
  if (lastAgg !== null && ev === lastVersion && nv === lastNodeVersion) return lastAgg;

  const decoded = [];
  let total = 0;
  for (const [row, buf] of getLatestEdgeStreamFrames()) {
    const d = decodeEdgeStreamFrame(row, buf);
    if (!d) continue;
    decoded.push(d);
    total += d.beadCount;
  }

  const positions = new Float32Array(total * 3);
  const ringMatrix = new Float32Array(total * 16);
  const value = new Int32Array(total);
  const srcNodeRow = new Int32Array(total);
  let w = 0;
  let m = 0;
  let b = 0;
  for (const d of decoded) {
    const src = readEdgeSrcNodeRow(d.edgeView, 0);

    for (let i = 0; i < d.beadCount; i++) {
      positions[w++] = readEdgeBeadX(d.beadView, i);
      positions[w++] = readEdgeBeadY(d.beadView, i);
      positions[w++] = readEdgeBeadZ(d.beadView, i);
      value[b] = readEdgeBeadValue(d.beadView, i);
      srcNodeRow[b++] = src;

      ringMatrix[m++] = readEdgeBeadRingM0(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM1(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM2(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM3(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM4(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM5(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM6(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM7(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM8(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM9(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM10(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM11(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM12(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM13(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM14(d.beadView, i);
      ringMatrix[m++] = readEdgeBeadRingM15(d.beadView, i);
    }
  }

  lastVersion = ev;
  lastNodeVersion = nv;
  lastAgg = { positions, value, srcNodeRow, ringMatrix, count: total };
  return lastAgg;
}
