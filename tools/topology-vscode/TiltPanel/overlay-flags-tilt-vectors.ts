import { useSyncExternalStore } from "react";
import { getNodeSections, subscribeNodeStreamBlocks } from "../src/webview/three/scene/nodes/node-sections";
import { columnF32, columnI32, columnU8 } from "../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../Buffer/column-owners";
import {
  COL_STREAM_NODE_TOP_TILT_VECTOR_LEN, COL_STREAM_NODE_TOP_TILT_VECTOR_IDX,
  COL_STREAM_NODE_LATTICE_POINTS,
  COL_STREAM_NODE_ROUNDS_TO_PARALLEL, COL_STREAM_NODE_MSGS_TO_PARALLEL,
} from "../Buffer/column-streams-gen";
import { nodeLabel } from "../src/webview/three/decode/buffer-decode-node";

export interface TiltVectorRow {
  row: number;
  label: string;
  idx: number;

  points: number;

  roundsToParallel: number;

  msgsToParallel: number;
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
      ai.points !== bi.points ||
      ai.roundsToParallel !== bi.roundsToParallel ||
      ai.msgsToParallel !== bi.msgsToParallel
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
      roundsToParallel: columnI32(nodeColumn(row, COL_STREAM_NODE_ROUNDS_TO_PARALLEL)),
      msgsToParallel: columnI32(nodeColumn(row, COL_STREAM_NODE_MSGS_TO_PARALLEL)),
    });
  }
  if (cachedTiltVectorRows && tiltVectorRowsEqual(cachedTiltVectorRows, next)) return cachedTiltVectorRows;
  cachedTiltVectorRows = next;
  return cachedTiltVectorRows;
}

export function useTiltVectorRows(): TiltVectorRow[] | null {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readTiltVectorRows, readTiltVectorRows);
}
