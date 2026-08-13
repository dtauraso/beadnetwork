import {
  readOverlaySceneTori, readOverlayScenePoles, readOverlayNodePoles,
  readOverlaySelSpherePoles, readOverlayHandholds, readOverlayLabelsGlobal,
  readOverlayOverlaysVis,
  readOverlayNodeBody,
  readOverlayNodeRing,
  readOverlayRingPick,
  readOverlaySelectionRing,
  readOverlayHoverRing,
  readOverlayReachSphere,
  readOverlaySceneVectors,
} from "../../../schema/buffer-layout/buffer-layout";
import type { ViewBlocksOrNull } from "./decode-event-line";

export const OVERLAY_KINDS = new Set([
  "scene-tori", "scene-poles", "node-poles", "sel-sphere-poles",
  "handholds", "labels-global", "overlays-vis",
  "node-body", "node-ring", "ring-pick", "selection-ring", "hover-ring", "reach-sphere",
  "scene-vectors",
]);

export function overlayFlag(vb: ViewBlocksOrNull, kind: string): number {
  const v = vb.overlayView;
  if (!v) return 0;
  switch (kind) {
    case "scene-tori": return readOverlaySceneTori(v);
    case "scene-poles": return readOverlayScenePoles(v);
    case "node-poles": return readOverlayNodePoles(v);
    case "sel-sphere-poles": return readOverlaySelSpherePoles(v);
    case "handholds": return readOverlayHandholds(v);
    case "labels-global": return readOverlayLabelsGlobal(v);
    case "overlays-vis": return readOverlayOverlaysVis(v);
    case "node-body": return readOverlayNodeBody(v);
    case "node-ring": return readOverlayNodeRing(v);
    case "ring-pick": return readOverlayRingPick(v);
    case "selection-ring": return readOverlaySelectionRing(v);
    case "hover-ring": return readOverlayHoverRing(v);
    case "reach-sphere": return readOverlayReachSphere(v);
    case "scene-vectors": return readOverlaySceneVectors(v);
    default: return 0;
  }
}
