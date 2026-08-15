import { ByteWriter, enumIndex } from "./byte-writer";
import {
  IN_KIND_EDIT_UPDATE,
  IN_UPDATE_KINDS,
} from "./input-layout-gen";
import {
  IN_OVERLAY_ATTR_TOGGLE,
  IN_CLOCK_ATTR_SPEED,
  IN_DISTANCE_GROUP_ATTR_LENGTH,
  IN_SCENE_ATTR_SELECTED,
  IN_PANEL_ATTR_TOGGLE,
  IN_NODE_ATTR_DRAG_PHI,
  IN_NODE_ATTR_DRAG_MAX_THETA,
  IN_NODE_ATTR_DRAG_ACTIVE,
} from "./input-attrs";
import type { OverlayFlag, PanelFlag } from "../../messages";
import { OVERLAY_FLAG_ORDER, PANEL_FLAG_ORDER } from "../../messages";

export function encodeOverlaysToggle(flag: OverlayFlag): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "overlays"));
  w.u8(IN_OVERLAY_ATTR_TOGGLE);
  w.u8(enumIndex(OVERLAY_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}

export function encodePanelsToggle(flag: PanelFlag): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "panels"));
  w.u8(IN_PANEL_ATTR_TOGGLE);
  w.u8(enumIndex(PANEL_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}

export function encodeClockSpeed(speed: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "clock"));
  w.u8(IN_CLOCK_ATTR_SPEED);
  w.u8(Math.round(speed * 4));
  return w.toArrayBuffer();
}

export function encodeDistanceGroupAdjust(groupIndex: number, dir: "up" | "down"): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "distanceGroup"));
  w.u8(IN_DISTANCE_GROUP_ATTR_LENGTH);
  w.u8(groupIndex);
  w.u8(dir === "up" ? 1 : 0);
  return w.toArrayBuffer();
}

export function encodeSceneSelected(tabIndex: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_SELECTED);
  w.u8(tabIndex);
  return w.toArrayBuffer();
}

export function encodeNodeDragPhiToggle(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "node"));
  w.u8(IN_NODE_ATTR_DRAG_PHI);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeNodeDragMaxTheta(nodeRow: number, degrees: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "node"));
  w.u8(IN_NODE_ATTR_DRAG_MAX_THETA);
  w.u8(nodeRow);
  w.f32(degrees);
  return w.toArrayBuffer();
}

export function encodeEdgeDragActiveToggle(edgeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "edge"));
  w.u8(IN_NODE_ATTR_DRAG_ACTIVE);
  w.u8(edgeRow);
  return w.toArrayBuffer();
}

export function encodeNodeDragActiveToggle(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "node"));
  w.u8(IN_NODE_ATTR_DRAG_ACTIVE);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

