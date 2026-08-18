import { useSyncExternalStore } from "react";
import { getNodeFrame, subscribeNodeStreamBlocks } from "../../scene/nodes/node-frame-aggregate";
import {
  readNodeTopTiltVectorLen,
  readNodeTopTiltVectorIdx,
  readNodeLatticePoints,
  readNodeRoundsToParallel,
  readNodeMsgsToParallel,
} from "../../../../../Buffer/buffer-layout";
import { nodeLabel } from "../../decode/buffer-decode-node";

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
  const decoded = getNodeFrame();
  if (!decoded) return cachedTiltVectorRows;
  const { nodeCount, nodeView } = decoded;
  const next: TiltVectorRow[] = [];
  for (let row = 0; row < nodeCount; row++) {
    if (!(readNodeTopTiltVectorLen(nodeView, row) > 0)) continue;
    next.push({
      row,
      label: nodeLabel(decoded, row),
      idx: readNodeTopTiltVectorIdx(nodeView, row),
      points: readNodeLatticePoints(nodeView, row),
      roundsToParallel: readNodeRoundsToParallel(nodeView, row),
      msgsToParallel: readNodeMsgsToParallel(nodeView, row),
    });
  }
  if (cachedTiltVectorRows && tiltVectorRowsEqual(cachedTiltVectorRows, next)) return cachedTiltVectorRows;
  cachedTiltVectorRows = next;
  return cachedTiltVectorRows;
}

export function useTiltVectorRows(): TiltVectorRow[] | null {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readTiltVectorRows, readTiltVectorRows);
}
