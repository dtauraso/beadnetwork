import { columnF32, columnU8 } from "../../schema/buffer-layout/column-values";
import { canvasFont, roundRect } from "../../webview/canvas-box";
import { readText, decodeAt } from "../../schema/buffer-layout/column-reads";
import * as T from "../../webview/canvas-theme";
import {
  COL_STREAM_POINTER_TARGET_X, COL_STREAM_POINTER_TARGET_Y, COL_STREAM_POINTER_TARGET_W,
  COL_STREAM_POINTER_TARGET_H, COL_STREAM_POINTER_TARGET_KIND,
  COL_STREAM_POINTER_TARGET_TIP_X, COL_STREAM_POINTER_TARGET_TIP_Y,
  COL_STREAM_POINTER_TARGET_TIP_TEXT,
} from "./columns-gen";

const KIND_NOTHING = 0;
const KIND_REFUSING = 2;

const TIP_PAD_X = 6;
const TIP_HEIGHT = 18;
const TIP_RADIUS = 4;
const TIP_BG = "rgba(20,20,24,0.92)";

export function pointerTargetKey(): string {
  return [
    columnF32(COL_STREAM_POINTER_TARGET_X), columnF32(COL_STREAM_POINTER_TARGET_Y),
    columnF32(COL_STREAM_POINTER_TARGET_W), columnF32(COL_STREAM_POINTER_TARGET_H),
    columnU8(COL_STREAM_POINTER_TARGET_KIND),
    columnF32(COL_STREAM_POINTER_TARGET_TIP_X), columnF32(COL_STREAM_POINTER_TARGET_TIP_Y),
  ].join(",");
}

export function pointerTargetCursor(): string {
  switch (columnU8(COL_STREAM_POINTER_TARGET_KIND)) {
    case KIND_NOTHING: return "default";
    case KIND_REFUSING: return "not-allowed";
    default: return "pointer";
  }
}

export function drawPointerHighlight(c: CanvasRenderingContext2D): void {
  if (columnU8(COL_STREAM_POINTER_TARGET_KIND) === KIND_NOTHING) return;
  const w = columnF32(COL_STREAM_POINTER_TARGET_W);
  const h = columnF32(COL_STREAM_POINTER_TARGET_H);
  if (w <= 0 || h <= 0) return;
  roundRect(
    c, columnF32(COL_STREAM_POINTER_TARGET_X), columnF32(COL_STREAM_POINTER_TARGET_Y),
    w, h, T.RADIUS_ITEM,
  );
  c.fillStyle = T.HOVER_ROW;
  c.fill();
}

export function drawPointerTip(c: CanvasRenderingContext2D): void {
  const text = readText(COL_STREAM_POINTER_TARGET_TIP_TEXT);
  if (!text || text.length === 0) return;
  const label = decodeAt(text, 0, text.length);
  if (!label) return;

  const x = columnF32(COL_STREAM_POINTER_TARGET_TIP_X);
  const y = columnF32(COL_STREAM_POINTER_TARGET_TIP_Y);

  c.font = canvasFont(T.FONT_SIZE);
  const w = c.measureText(label).width + 2 * TIP_PAD_X;

  roundRect(c, x, y, w, TIP_HEIGHT, TIP_RADIUS);
  c.fillStyle = TIP_BG;
  c.fill();

  c.fillStyle = T.TEXT;
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(label, x + TIP_PAD_X, y + TIP_HEIGHT / 2);
}
