import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "../../Input/messages";
import { columnU8 } from "../../Buffer/column-values";
import {
  COL_STREAM_OVERLAY_SCENE_TORI, COL_STREAM_OVERLAY_SCENE_POLES,
  COL_STREAM_OVERLAY_NODE_POLES, COL_STREAM_OVERLAY_HANDHOLDS,
  COL_STREAM_OVERLAY_LABELS_GLOBAL, COL_STREAM_OVERLAY_OVERLAYS_VIS,
  COL_STREAM_OVERLAY_NODE_BODY, COL_STREAM_OVERLAY_NODE_RING, COL_STREAM_OVERLAY_RING_PICK,
  COL_STREAM_OVERLAY_SELECTION_RING, COL_STREAM_OVERLAY_HOVER_RING,
  COL_STREAM_OVERLAY_SCENE_VECTORS, COL_STREAM_OVERLAY_RULE_CHANNELS,
  COL_STREAM_OVERLAY_NODE_POLE_SPHERE, COL_STREAM_OVERLAY_ALL_POLE_SPHERES,
} from "../../Overlay/columns-gen";

export type OverlayFlagVals = Record<OverlayFlag, boolean>;

export function readOverlayFlags(): OverlayFlagVals {
  return {
    tori: !!columnU8(COL_STREAM_OVERLAY_SCENE_TORI),
    scenePoles: !!columnU8(COL_STREAM_OVERLAY_SCENE_POLES),
    nodePoles: !!columnU8(COL_STREAM_OVERLAY_NODE_POLES),
    handholds: !!columnU8(COL_STREAM_OVERLAY_HANDHOLDS),

    labelsGlobal: !columnU8(COL_STREAM_OVERLAY_LABELS_GLOBAL),
    overlays: !!columnU8(COL_STREAM_OVERLAY_OVERLAYS_VIS),
    nodeBody: !!columnU8(COL_STREAM_OVERLAY_NODE_BODY),
    nodeRing: !!columnU8(COL_STREAM_OVERLAY_NODE_RING),
    ringPick: !!columnU8(COL_STREAM_OVERLAY_RING_PICK),
    selectionRing: !!columnU8(COL_STREAM_OVERLAY_SELECTION_RING),
    hoverRing: !!columnU8(COL_STREAM_OVERLAY_HOVER_RING),
    sceneVectors: !!columnU8(COL_STREAM_OVERLAY_SCENE_VECTORS),
    ruleChannels: !!columnU8(COL_STREAM_OVERLAY_RULE_CHANNELS),
    nodePoleSphere: !!columnU8(COL_STREAM_OVERLAY_NODE_POLE_SPHERE),
    allPoleSpheres: !!columnU8(COL_STREAM_OVERLAY_ALL_POLE_SPHERES),
  };
}

export function overlayFlag(name: OverlayFlag): boolean {
  return readOverlayFlags()[name];
}

export function overlayFlagSignature(): string {
  const vals = readOverlayFlags();
  return OVERLAY_FLAG_ORDER.map((flag) => (vals[flag] ? "1" : "0")).join("");
}
