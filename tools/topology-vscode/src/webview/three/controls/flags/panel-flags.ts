import { useSyncExternalStore } from "react";
import { PANEL_FLAG_ORDER, type PanelFlag } from "../../../../messages";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import {
  readPanelOverlays,
  readPanelNode,
  readPanelNodeShape,
  readPanelNodeState,
  readPanelNodeReach,
  readPanelNodePoles,
  readPanelScene,
  readPanelSceneGuides,
  readPanelScenePoles,
  readPanelSceneVectors,
  readPanelSceneLabels,
} from "../../../../schema/buffer-layout/buffer-layout";

export type PanelFlagVals = Record<PanelFlag, boolean>;

let cachedVals: PanelFlagVals | null = null;

function panelFlagsEqual(a: PanelFlagVals, b: PanelFlagVals): boolean {
  return PANEL_FLAG_ORDER.every((flag) => a[flag] === b[flag]);
}

export function readPanelFlags(): PanelFlagVals | null {
  const blocks = getViewBlocks();
  if (!blocks) return cachedVals;
  const v = blocks.panelView;
  const next: PanelFlagVals = {
    overlays: !!readPanelOverlays(v),
    node: !!readPanelNode(v),
    nodeShape: !!readPanelNodeShape(v),
    nodeState: !!readPanelNodeState(v),
    nodeReach: !!readPanelNodeReach(v),
    nodePoles: !!readPanelNodePoles(v),
    scene: !!readPanelScene(v),
    sceneGuides: !!readPanelSceneGuides(v),
    scenePoles: !!readPanelScenePoles(v),
    sceneVectors: !!readPanelSceneVectors(v),
    sceneLabels: !!readPanelSceneLabels(v),
  };
  if (cachedVals && panelFlagsEqual(cachedVals, next)) return cachedVals;
  cachedVals = next;
  return cachedVals;
}

export function panelOn(read: (v: DataView) => number): boolean {
  const blocks = getViewBlocks();
  if (!blocks) return false;
  return read(blocks.panelView) !== 0;
}

export function usePanelFlags(): PanelFlagVals | null {
  return useSyncExternalStore(subscribeViewBlocks, readPanelFlags, readPanelFlags);
}
