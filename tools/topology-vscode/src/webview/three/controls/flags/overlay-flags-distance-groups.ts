import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import {
  readOverlayGroupLenTime,
  readOverlayGroupLenInput,
  readOverlayGroupLenGate,
} from "../../../../schema/buffer-layout/buffer-layout";

export interface DistanceGroupLens {
  time: number;
  input: number;
  gate: number;
}

let cachedGroupLens: DistanceGroupLens | null = null;

function distanceGroupLensEqual(a: DistanceGroupLens, b: DistanceGroupLens): boolean {
  return a.time === b.time && a.input === b.input && a.gate === b.gate;
}

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

export function useDistanceGroupLens(): DistanceGroupLens | null {
  return useSyncExternalStore(subscribeViewBlocks, readDistanceGroupLens, readDistanceGroupLens);
}
