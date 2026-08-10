// overlay-flags-distance-groups.ts — a row-keyed READ resource over the buffer's
// GroupLenTime/GroupLenInput/GroupLenGate columns (Overlay block). Split out of
// overlay-flags.ts — see that file's header for the full sibling-file list.

import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../scene/view-blocks";
import {
  readOverlayGroupLenTime,
  readOverlayGroupLenInput,
  readOverlayGroupLenGate,
} from "../../../schema/buffer-layout";

/** The "distance home button" toolbar panel's 3 group max-pair-lengths, in Go's
 *  distanceGroupOrder (nodes/Wiring/distance_groups.go): time, input, gate. Read-only
 *  reflect of the Overlay block's GroupLenTime/GroupLenInput/GroupLenGate columns — Go
 *  computes these fresh every VIEW-frame emit; TS holds no group definitions. */
export interface DistanceGroupLens {
  time: number;
  input: number;
  gate: number;
}

let cachedGroupLens: DistanceGroupLens | null = null;

function distanceGroupLensEqual(a: DistanceGroupLens, b: DistanceGroupLens): boolean {
  return a.time === b.time && a.input === b.input && a.gate === b.gate;
}

/** Decode the current 3 group max-pair-lengths, or null if no snapshot yet. Stable
 *  identity while unchanged (useSyncExternalStore compares by identity). */
export function readDistanceGroupLens(): DistanceGroupLens | null {
  const blocks = getViewBlocks();
  if (!blocks) return cachedGroupLens;
  const overlayView = blocks.overlayView;
  const next: DistanceGroupLens = {
    time: readOverlayGroupLenTime(overlayView),
    input: readOverlayGroupLenInput(overlayView),
    gate: readOverlayGroupLenGate(overlayView),
  };
  if (cachedGroupLens && distanceGroupLensEqual(cachedGroupLens, next)) return cachedGroupLens;
  cachedGroupLens = next;
  return cachedGroupLens;
}

/** React hook: re-renders the caller when any of the 3 group max-pair-lengths change. */
export function useDistanceGroupLens(): DistanceGroupLens | null {
  return useSyncExternalStore(subscribeViewBlocks, readDistanceGroupLens, readDistanceGroupLens);
}
