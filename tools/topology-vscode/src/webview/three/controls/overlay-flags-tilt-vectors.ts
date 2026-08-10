// overlay-flags-tilt-vectors.ts — a row-keyed READ resource over each node's own
// TopTiltVectorLen/TopTiltVectorTheta/LatticePoints/RoundsToParallel/MsgsToParallel
// columns (Node block). Split out of overlay-flags.ts — see that file's header for the
// full sibling-file list.

import { useSyncExternalStore } from "react";
import { getNodeFrame, subscribeNodeStreamBlocks } from "../scene/node-stream-blocks";
import {
  readNodeTopTiltVectorLen,
  readNodeTopTiltVectorTheta,
  readNodeLatticePoints,
  readNodeRoundsToParallel,
  readNodeMsgsToParallel,
} from "../../../schema/buffer-layout";
import { nodeLabel } from "../decode/buffer-decode-node";

/** One row of the per-node tilt-vector-angle panel: read-only reflect of a single node's
 *  own TopTiltVectorTheta (Buffer/layout.go), as the ALREADY-MULTIPLIED
 *  radians the buffer carries — TS holds no step constant of its own
 *  (nodes/Wiring.CurveParamTiltVectorAngleStep is Go's). row is the node's buffer ROW
 *  (never an id/name — no sidecar), label its human label for display only. There is no
 *  φ field: every tilt vector in this model is θ-only (TiltVectorAnglePanel.tsx's own
 *  doc comment). */
export interface TiltVectorRow {
  row: number;
  label: string;
  theta: number;
  /** This node's own streamed lattice point count (Buffer/layout.go's LatticePoints) — the
   *  N `theta` was converted against, so a reader can invert it back to an index at the
   *  CURRENT count instead of assuming a fixed compile-time step. */
  points: number;
  /** This node's own streamed rounds-to-rest count (Buffer/layout.go's RoundsToParallel) —
   *  vector-exchange rounds between the exchange opening and this node's rule settling.
   *  Frozen at rest by Go, so it does not climb while the settled exchange keeps
   *  circulating; 0 means not yet at rest, or opened already at rest. */
  roundsToParallel: number;
  /** The same span counted in vector-channel messages (Buffer/layout.go's MsgsToParallel) —
   *  every receive and every reply this node performed. Streamed separately rather than
   *  derived as 2×rounds, because Go owns the relationship between the two. */
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
      ai.theta !== bi.theta ||
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

/** Decode every node whose TopTiltVectorLen is > 0 (Go's "this node draws a tilt vector"
 *  answer — same column TiltVectors.tsx gates its own draw on) into a TiltVectorRow list,
 *  or null if no node frame has decoded yet. An EMPTY (non-null) list is the "no
 *  groups"-shaped signal for a scene that streams no tilt vectors at all — the panel that
 *  reads this renders nothing for that case, with no scene branch on either side. */
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
      theta: readNodeTopTiltVectorTheta(nodeView, row),
      points: readNodeLatticePoints(nodeView, row),
      roundsToParallel: readNodeRoundsToParallel(nodeView, row),
      msgsToParallel: readNodeMsgsToParallel(nodeView, row),
    });
  }
  if (cachedTiltVectorRows && tiltVectorRowsEqual(cachedTiltVectorRows, next)) return cachedTiltVectorRows;
  cachedTiltVectorRows = next;
  return cachedTiltVectorRows;
}

/** React hook: re-renders the caller when the set of tilt-vector-drawing nodes or any of
 *  their angles changes. Returns null until the first node frame decodes. */
export function useTiltVectorRows(): TiltVectorRow[] | null {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readTiltVectorRows, readTiltVectorRows);
}
