import { getEdgeStreamAccessor } from "../../../Node/Edge/edge-stream-blocks";
import { nodeF32, nodeI32, nodeU8 as nodeU8Value } from "../../../Node/node-leaves";
import type { NodeValueName } from "../../../Node/node-values-gen";
import { formatPi } from "./pi-fraction";
import { canvasFont, roundRect } from "../../canvas-box";

export const ROW_NODE_HEAD = 1;
export const ROW_HOLDER_HEAD = 2;
export const ROW_LINE = 3;

export const CHECK_NONE = 0;
export const CHECK_NODE_DRAG = 1;
export const CHECK_SELF_DRAG = 2;
export const CHECK_EDGE_DRAG = 3;
export const CHECK_KIND_RULE = 4;

export const VAL_NONE = 0;
export const VAL_SELF_R = 2;
export const VAL_SELF_PHI = 3;
export const VAL_SELF_THETA = 4;
export const VAL_DRAG_R = 5;
export const VAL_DRAG_PHI = 6;
export const VAL_DRAG_THETA = 7;

const RAD_TO_PI = 1 / Math.PI;

const nodeOn = (row: number, name: NodeValueName) => nodeU8Value(row, name) !== 0;

export function nodeDragActive(row: number): boolean {
  return nodeOn(row, "dragActive");
}

export function nodeRuleGroup(row: number): { id: number; size: number } {
  return {
    id: nodeI32(row, "ruleGroupId"),
    size: nodeI32(row, "ruleGroupSize"),
  };
}

export function checkValue(check: number, nodeRow: number, edgeRow: number): boolean {
  switch (check) {
    case CHECK_NODE_DRAG: return nodeDragActive(nodeRow);
    case CHECK_SELF_DRAG: return nodeOn(nodeRow, "selfActive");
    case CHECK_KIND_RULE: return nodeOn(nodeRow, "kindRuleActive");
    case CHECK_EDGE_DRAG: {
      const edges = getEdgeStreamAccessor();
      return edgeRow >= 0 && !!edges && edges.dragActive(edgeRow);
    }
    default: return false;
  }
}

function thetaText(maxTheta: number): string {
  if (maxTheta < 0) return "free";
  const span = formatPi(maxTheta * RAD_TO_PI);
  return `∈ [−${span}, +${span}]`;
}

export function valueText(
  value: number,
  nodeRow: number,
  edgeRow: number,
): { text: string; free: boolean } {
  switch (value) {
    case VAL_SELF_R: {
      const on = nodeOn(nodeRow, "selfRLocked");
      return { text: on ? "held" : "free", free: !on };
    }
    case VAL_SELF_PHI: {
      const on = nodeOn(nodeRow, "selfPhiLocked");
      return { text: on ? "locked" : "free", free: !on };
    }
    case VAL_SELF_THETA: {
      const t = nodeF32(nodeRow, "selfThetaMax");
      return { text: thetaText(t), free: t < 0 };
    }
    case VAL_DRAG_R: {
      const on = nodeOn(nodeRow, "dragRLocked");
      if (!on) return { text: "free", free: true };
      const edges = getEdgeStreamAccessor();
      const r = edgeRow >= 0 && edges ? Math.abs(edges.deltaR(edgeRow)) : 0;
      return { text: `held ${Math.round(r * 10) / 10}`, free: false };
    }
    case VAL_DRAG_PHI: {
      const on = nodeOn(nodeRow, "dragPhiLocked");
      return { text: on ? "locked" : "free", free: !on };
    }
    case VAL_DRAG_THETA: {
      const t = nodeF32(nodeRow, "dragThetaMax");
      return { text: thetaText(t), free: t < 0 };
    }
    default: return { text: "", free: false };
  }
}

export function drawCheckbox(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  on: boolean,
): void {
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, 3);
  c.fillStyle = on ? EDIT_EDGE : "#fff";
  c.fill();
  c.strokeStyle = on ? EDIT_EDGE : MUTED;
  c.lineWidth = 1;
  c.stroke();
  if (!on) return;
  c.fillStyle = "#fff";
  c.font = canvasFont(9, 900);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText("✓", x + w / 2, y + h / 2);
}

export const PANEL_BG = "#fff";
export const PANEL_EDGE = "#ddd";
export const INK = "#333";
export const NODE_INK = "#222";
export const MUTED = "#666";
export const FREE_INK = "#999";
export const GLYPH_INK = "#1976d2";
export const FREE_GLYPH = "#aaa";
export const EDIT_BG = "#e3f2fd";
export const EDIT_EDGE = "#1976d2";
export const PILL_BG = "#ede7f6";
export const PILL_EDGE = "#b39ddb";
export const PILL_INK = "#5e35b1";
export const INACTIVE_ALPHA = 0.45;
export const FONT_PX = 12;
export const HEAD_FONT_PX = 11;
