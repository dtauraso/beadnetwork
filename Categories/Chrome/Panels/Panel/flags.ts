// PANEL_FLAGS_START
const PANEL_FLAG_NAMES = [
  "overlays",
  "node",
  "nodeShape",
  "nodeState",
  "nodePoles",
  "nodeRules",
  "nodeVectors",
  "scene",
  "sceneGuides",
  "scenePoles",
  "sceneVectors",
  "sceneLabels",
] as const;
// PANEL_FLAGS_END

export type PanelFlag = (typeof PANEL_FLAG_NAMES)[number];

export const PANEL_FLAG_ORDER = PANEL_FLAG_NAMES;
