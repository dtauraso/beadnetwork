import { panelFont, roundRect } from "./panel-box";
import * as T from "../src/webview/three/controls/chrome-theme";

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
  c.font = panelFont(T.FONT_SIZE, T.FONT_WEIGHT_LABEL);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(label, x + PILL_PAD_X, y + h / 2);

  c.font = panelFont(T.FONT_SIZE_GLYPH);
  c.textAlign = "center";
  c.fillText(open ? "▲" : "▼", x + w - CARET_W / 2 - 3, y + h / 2);
}

export function drawPopoverBox(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_PANEL);
  c.fillStyle = T.SURFACE;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();
}
