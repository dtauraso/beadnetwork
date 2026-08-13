import { useSyncExternalStore } from "react";
import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "../../../../messages";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import {
  readOverlaySceneTori,
  readOverlayScenePoles,
  readOverlayNodePoles,
  readOverlaySelSpherePoles,
  readOverlayHandholds,
  readOverlayLabelsGlobal,
  readOverlayOverlaysVis,
  readOverlayNodeBody,
  readOverlayNodeRing,
  readOverlayRingPick,
  readOverlaySelectionRing,
  readOverlayHoverRing,
  readOverlayReachSphere,
  readOverlaySceneVectors,
} from "../../../../schema/buffer-layout/buffer-layout";

export type OverlayFlagVals = Record<OverlayFlag, boolean>;

let cachedVals: OverlayFlagVals | null = null;

function overlayFlagsEqual(a: OverlayFlagVals, b: OverlayFlagVals): boolean {
  return OVERLAY_FLAG_ORDER.every((flag) => a[flag] === b[flag]);
}

export function readOverlayFlags(): OverlayFlagVals | null {
  const blocks = getViewBlocks();
  if (!blocks) return cachedVals;
  const v = blocks.overlayView;
  const next: OverlayFlagVals = {
    tori: !!readOverlaySceneTori(v),
    scenePoles: !!readOverlayScenePoles(v),
    nodePoles: !!readOverlayNodePoles(v),
    selSpherePoles: !!readOverlaySelSpherePoles(v),
    handholds: !!readOverlayHandholds(v),

    labelsGlobal: !readOverlayLabelsGlobal(v),
    overlays: !!readOverlayOverlaysVis(v),
    nodeBody: !!readOverlayNodeBody(v),
    nodeRing: !!readOverlayNodeRing(v),
    ringPick: !!readOverlayRingPick(v),
    selectionRing: !!readOverlaySelectionRing(v),
    hoverRing: !!readOverlayHoverRing(v),
    reachSphere: !!readOverlayReachSphere(v),
    sceneVectors: !!readOverlaySceneVectors(v),
  };
  if (cachedVals && overlayFlagsEqual(cachedVals, next)) return cachedVals;
  cachedVals = next;
  return cachedVals;
}

export function overlayOn(read: (v: DataView) => number): boolean {
  const blocks = getViewBlocks();
  if (!blocks) return false;
  return read(blocks.overlayView) !== 0;
}

export function useOverlayFlags(): OverlayFlagVals | null {
  return useSyncExternalStore(subscribeViewBlocks, readOverlayFlags, readOverlayFlags);
}

