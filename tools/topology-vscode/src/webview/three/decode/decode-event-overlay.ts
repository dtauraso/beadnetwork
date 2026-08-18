import { columnU8 } from "../../../../Buffer/column-values";
import {
  COL_STREAM_OVERLAY_SCENE_TORI, COL_STREAM_OVERLAY_SCENE_POLES, COL_STREAM_OVERLAY_NODE_POLES,
  COL_STREAM_OVERLAY_NODE_POLE_SPHERE, COL_STREAM_OVERLAY_ALL_POLE_SPHERES,
  COL_STREAM_OVERLAY_HANDHOLDS, COL_STREAM_OVERLAY_LABELS_GLOBAL,
  COL_STREAM_OVERLAY_OVERLAYS_VIS,
  COL_STREAM_OVERLAY_NODE_BODY,
  COL_STREAM_OVERLAY_NODE_RING,
  COL_STREAM_OVERLAY_RING_PICK,
  COL_STREAM_OVERLAY_SELECTION_RING,
  COL_STREAM_OVERLAY_HOVER_RING,
  COL_STREAM_OVERLAY_SCENE_VECTORS,
} from "../../../../Buffer/column-streams-gen";

export const OVERLAY_KINDS = new Set([
  "scene-tori", "scene-poles", "node-poles",
  "handholds", "labels-global", "overlays-vis",
  "node-body", "node-ring", "ring-pick", "selection-ring", "hover-ring",
  "scene-vectors",
]);

export function overlayFlag(kind: string): number {
  switch (kind) {
    case "scene-tori": return columnU8(COL_STREAM_OVERLAY_SCENE_TORI);
    case "scene-poles": return columnU8(COL_STREAM_OVERLAY_SCENE_POLES);
    case "node-poles": return columnU8(COL_STREAM_OVERLAY_NODE_POLES);
    case "handholds": return columnU8(COL_STREAM_OVERLAY_HANDHOLDS);
    case "labels-global": return columnU8(COL_STREAM_OVERLAY_LABELS_GLOBAL);
    case "overlays-vis": return columnU8(COL_STREAM_OVERLAY_OVERLAYS_VIS);
    case "node-body": return columnU8(COL_STREAM_OVERLAY_NODE_BODY);
    case "node-ring": return columnU8(COL_STREAM_OVERLAY_NODE_RING);
    case "ring-pick": return columnU8(COL_STREAM_OVERLAY_RING_PICK);
    case "selection-ring": return columnU8(COL_STREAM_OVERLAY_SELECTION_RING);
    case "hover-ring": return columnU8(COL_STREAM_OVERLAY_HOVER_RING);
    case "scene-vectors": return columnU8(COL_STREAM_OVERLAY_SCENE_VECTORS);
    case "node-pole-sphere": return columnU8(COL_STREAM_OVERLAY_NODE_POLE_SPHERE);
    case "all-pole-spheres": return columnU8(COL_STREAM_OVERLAY_ALL_POLE_SPHERES);
    default: return 0;
  }
}
