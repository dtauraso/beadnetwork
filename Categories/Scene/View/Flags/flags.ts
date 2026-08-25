// OVERLAY_FLAGS_START
const OVERLAY_FLAG_NAMES = [
  "tori",
  "scenePoles",
  "nodePoles",
  "handholds",
  "labelsGlobal",
  "overlays",

  "nodeBody",
  "nodeRing",
  "ringPick",
  "selectionRing",
  "hoverRing",
  "sceneVectors",
  "edgeVectors",
  "topVectors",
  "ruleChannels",
  "nodePoleSphere",
  "allPoleSpheres",
] as const;
// OVERLAY_FLAGS_END

export type OverlayFlag = (typeof OVERLAY_FLAG_NAMES)[number];

export const OVERLAY_FLAG_ORDER = OVERLAY_FLAG_NAMES;
