import { columnBytes, columnF32 } from "../Buffer/column-values";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { readText, decodeAt } from "../PanelOverlay/panel-columns";
import { EDIT_BG, EDIT_EDGE, FONT_PX } from "./rules-values";
import {
  COL_STREAM_RULES_PANEL_ROW_EDITING, COL_STREAM_RULES_PANEL_DRAFT_TEXT,
  COL_STREAM_RULES_PANEL_DRAFT_X, COL_STREAM_RULES_PANEL_DRAFT_Y,
  COL_STREAM_RULES_PANEL_DRAFT_W, COL_STREAM_RULES_PANEL_DRAFT_H,
} from "./columns-gen";

export function rulesDraftOpen(): boolean {
  const editing = columnBytes(COL_STREAM_RULES_PANEL_ROW_EDITING);
  if (!editing) return false;
  for (let i = 0; i < editing.byteLength; i++) {
    if (editing.getUint8(i) !== 0) return true;
  }
  return false;
}

export function isRowEditing(row: number): boolean {
  const editing = columnBytes(COL_STREAM_RULES_PANEL_ROW_EDITING);
  return !!editing && row < editing.byteLength && editing.getUint8(row) !== 0;
}

export function draftText(): string {
  const text = readText(COL_STREAM_RULES_PANEL_DRAFT_TEXT);
  return text ? decodeAt(text, 0, text.length) : "";
}

export function drawThetaDraft(c: CanvasRenderingContext2D): void {
  if (!rulesDraftOpen()) return;
  const text = readText(COL_STREAM_RULES_PANEL_DRAFT_TEXT);
  const w = columnF32(COL_STREAM_RULES_PANEL_DRAFT_W);
  const h = columnF32(COL_STREAM_RULES_PANEL_DRAFT_H);
  if (!text || w <= 0 || h <= 0) return;
  const x = columnF32(COL_STREAM_RULES_PANEL_DRAFT_X);
  const y = columnF32(COL_STREAM_RULES_PANEL_DRAFT_Y);

  roundRect(c, x + 0.5, y + 0.5, Math.min(w, 60) - 1, h - 1, 3);
  c.fillStyle = EDIT_BG;
  c.fill();
  c.strokeStyle = EDIT_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = "#111";
  c.font = panelFont(FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(`${decodeAt(text, 0, text.length)} π`, x + 5, y + h / 2);
}
