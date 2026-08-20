import { canvasFont, roundRect } from "../../webview/canvas-box";
import * as T from "../../webview/canvas-theme";

export const CARET_W = 20;
export const POPOVER_PAD = 6;
export const ROW_PAD_X = 6;
export const GLYPH_W = 8;
export const PILL_PAD_X = 9;

export function drawPill(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  label: string, open: boolean, active = false,
): void {
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = active ? T.ACCENT : T.CHIP;
  c.fill();
  c.strokeStyle = active ? T.ACCENT : T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = active ? T.ACCENT_INK : T.TEXT;
  c.font = canvasFont(T.FONT_SIZE, T.FONT_WEIGHT_LABEL);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(label, x + PILL_PAD_X, y + h / 2);

  c.font = canvasFont(T.FONT_SIZE_GLYPH);
  c.textAlign = "center";
  c.fillText(open ? "▲" : "▼", x + w - CARET_W / 2 - 3, y + h / 2);
}

export function drawHeadingText(
  c: CanvasRenderingContext2D, text: string, x: number, y: number,
): void {
  c.font = canvasFont(T.FONT_SIZE_HEADING);
  c.letterSpacing = T.HEADING_TRACKING;
  c.fillText(text.toUpperCase(), x, y);
  c.letterSpacing = "0px";
}

const SHADOW_COLOR = "rgba(0,0,0,0.5)";
const SHADOW_BLUR = 24;
const SHADOW_DY = 8;

export function drawPopoverBox(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_PANEL);

  c.save();
  c.shadowColor = SHADOW_COLOR;
  c.shadowBlur = SHADOW_BLUR;
  c.shadowOffsetY = SHADOW_DY;
  c.fillStyle = T.SURFACE;
  c.fill();
  c.restore();

  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();
}
