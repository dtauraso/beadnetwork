import { canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import { EDIT_BG, EDIT_EDGE, FONT_PX } from "./rules-values";
import { rulesBytes, rulesF32, rulesText } from "./rules-leaves";

export function rulesDraftOpen(): boolean {
  const editing = rulesBytes("rowEditing");
  if (!editing) return false;
  for (let i = 0; i < editing.byteLength; i++) {
    if (editing.getUint8(i) !== 0) return true;
  }
  return false;
}

export function isRowEditing(row: number): boolean {
  const editing = rulesBytes("rowEditing");
  return !!editing && row < editing.byteLength && editing.getUint8(row) !== 0;
}

export function draftText(): string {
  const text = rulesText("draftText");
  return text ? decodeAt(text, 0, text.length) : "";
}

export function drawThetaDraft(c: CanvasRenderingContext2D): void {
  if (!rulesDraftOpen()) return;
  const text = rulesText("draftText");
  const w = rulesF32("draftW");
  const h = rulesF32("draftH");
  if (!text || w <= 0 || h <= 0) return;
  const x = rulesF32("draftX");
  const y = rulesF32("draftY");

  roundRect(c, x + 0.5, y + 0.5, Math.min(w, 60) - 1, h - 1, 3);
  c.fillStyle = EDIT_BG;
  c.fill();
  c.strokeStyle = EDIT_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = "#111";
  c.font = canvasFont(FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(`${decodeAt(text, 0, text.length)} π`, x + 5, y + h / 2);
}
