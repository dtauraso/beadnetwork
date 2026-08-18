import { useSyncExternalStore } from "react";
import { PANEL_FLAG_ORDER, type PanelFlag } from "../../../../messages";
import { columnU8, subscribeColumns } from "../../../../../Buffer/column-values";
import {
  COL_STREAM_PANEL_OVERLAYS,
  COL_STREAM_PANEL_NODE,
  COL_STREAM_PANEL_NODE_SHAPE,
  COL_STREAM_PANEL_NODE_STATE,
  COL_STREAM_PANEL_NODE_POLES,
  COL_STREAM_PANEL_NODE_RULES,
  COL_STREAM_PANEL_SCENE,
  COL_STREAM_PANEL_SCENE_GUIDES,
  COL_STREAM_PANEL_SCENE_POLES,
  COL_STREAM_PANEL_SCENE_VECTORS,
  COL_STREAM_PANEL_SCENE_LABELS,
} from "../../../../../Buffer/column-streams-gen";

export type PanelFlagVals = Record<PanelFlag, boolean>;

let cachedVals: PanelFlagVals | null = null;

function panelFlagsEqual(a: PanelFlagVals, b: PanelFlagVals): boolean {
  return PANEL_FLAG_ORDER.every((flag) => a[flag] === b[flag]);
}

export function readPanelFlags(): PanelFlagVals | null {
  const next: PanelFlagVals = {
    overlays: !!columnU8(COL_STREAM_PANEL_OVERLAYS),
    node: !!columnU8(COL_STREAM_PANEL_NODE),
    nodeShape: !!columnU8(COL_STREAM_PANEL_NODE_SHAPE),
    nodeState: !!columnU8(COL_STREAM_PANEL_NODE_STATE),
    nodePoles: !!columnU8(COL_STREAM_PANEL_NODE_POLES),
    nodeRules: !!columnU8(COL_STREAM_PANEL_NODE_RULES),
    scene: !!columnU8(COL_STREAM_PANEL_SCENE),
    sceneGuides: !!columnU8(COL_STREAM_PANEL_SCENE_GUIDES),
    scenePoles: !!columnU8(COL_STREAM_PANEL_SCENE_POLES),
    sceneVectors: !!columnU8(COL_STREAM_PANEL_SCENE_VECTORS),
    sceneLabels: !!columnU8(COL_STREAM_PANEL_SCENE_LABELS),
  };
  if (cachedVals && panelFlagsEqual(cachedVals, next)) return cachedVals;
  cachedVals = next;
  return cachedVals;
}

export function usePanelFlags(): PanelFlagVals | null {
  return useSyncExternalStore(subscribeColumns, readPanelFlags, readPanelFlags);
}
