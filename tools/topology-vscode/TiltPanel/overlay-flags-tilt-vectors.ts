import { useSyncExternalStore } from "react";
import { columnF32, columnI32, columnU8 } from "../Buffer/column-values";
import { subscribeFrame } from "../src/webview/frame-tick";
import { nodeColumn, ownerCounts } from "../Buffer/column-owners";
import {
  COL_STREAM_NODE_TOP_TILT_VECTOR_LEN, COL_STREAM_NODE_TOP_TILT_VECTOR_IDX,
  COL_STREAM_NODE_LATTICE_POINTS,
} from "../Buffer/column-streams-gen";
import { nodeLabel } from "../src/webview/three/decode/buffer-decode-node";

export interface TiltVectorRow {
  row: number;
  label: string;
  idx: number;

  points: number;
}

let cachedTiltVectorRows: TiltVectorRow[] | null = null;

function tiltVectorRowsEqual(a: TiltVectorRow[], b: TiltVectorRow[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (
      ai.row !== bi.row ||
      ai.idx !== bi.idx ||
      ai.label !== bi.label ||
      ai.points !== bi.points
    ) {
      return false;
    }
  }
  return true;
}

export function readTiltVectorRows(): TiltVectorRow[] | null {
  const nodeCount = ownerCounts().nodes;
  if (nodeCount <= 0) return cachedTiltVectorRows;
  const next: TiltVectorRow[] = [];
  for (let row = 0; row < nodeCount; row++) {
    if (!(columnF32(nodeColumn(row, COL_STREAM_NODE_TOP_TILT_VECTOR_LEN)) > 0)) continue;
    next.push({
      row,
      label: nodeLabel(row),
      idx: columnI32(nodeColumn(row, COL_STREAM_NODE_TOP_TILT_VECTOR_IDX)),
      points: columnU8(nodeColumn(row, COL_STREAM_NODE_LATTICE_POINTS)),
    });
  }
  if (cachedTiltVectorRows && tiltVectorRowsEqual(cachedTiltVectorRows, next)) return cachedTiltVectorRows;
  cachedTiltVectorRows = next;
  return cachedTiltVectorRows;
}

export function useTiltVectorRows(): TiltVectorRow[] | null {
  return useSyncExternalStore(subscribeFrame, readTiltVectorRows, readTiltVectorRows);
}
