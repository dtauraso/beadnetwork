import { columnF32 } from "../../../Buffer/column-values";
import { canvasFont, roundRect } from "../../../webview/canvas-box";
import { readText, decodeAt } from "../../../Buffer/column-reads";
import * as T from "../../../webview/canvas-theme";
import {
  COL_STREAM_FIT_CHIP_X, COL_STREAM_FIT_CHIP_Y, COL_STREAM_FIT_CHIP_W, COL_STREAM_FIT_CHIP_H,
  COL_STREAM_FIT_CHIP_LABEL_TEXT,
} from "./columns-gen";

export function fitChipKey(): string {
  return [
    columnF32(COL_STREAM_FIT_CHIP_X), columnF32(COL_STREAM_FIT_CHIP_Y),
    columnF32(COL_STREAM_FIT_CHIP_W),
  ].join(",");
}

export function drawFitChip(c: CanvasRenderingContext2D): void {
  const label = readText(COL_STREAM_FIT_CHIP_LABEL_TEXT);
  const w = columnF32(COL_STREAM_FIT_CHIP_W);
  const h = columnF32(COL_STREAM_FIT_CHIP_H);
  if (!label || w <= 0 || h <= 0) return;
  const x = columnF32(COL_STREAM_FIT_CHIP_X);
  const y = columnF32(COL_STREAM_FIT_CHIP_Y);

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = T.TEXT;
  c.font = canvasFont(T.FONT_SIZE);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(decodeAt(label, 0, label.length), x + w / 2, y + h / 2);
}
