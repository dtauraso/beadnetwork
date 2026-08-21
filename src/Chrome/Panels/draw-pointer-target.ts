import { canvasFont, roundRect } from "../../webview/canvas-box";
import { decodeAt } from "../../webview/leaf-text";
import * as T from "../../webview/canvas-theme";
import { pointerF32, pointerU8, pointerText } from "./pointer-target-leaves";

const KIND_NOTHING = 0;
const KIND_REFUSING = 2;

const TIP_PAD_X = 6;
const TIP_HEIGHT = 18;
const TIP_RADIUS = 4;
const TIP_BG = "rgba(20,20,24,0.92)";

export function pointerTargetKey(): string {
  return [
    pointerF32("x"), pointerF32("y"),
    pointerF32("w"), pointerF32("h"),
    pointerU8("kind"),
    pointerF32("tipX"), pointerF32("tipY"),
  ].join(",");
}

export function pointerTargetCursor(): string {
  switch (pointerU8("kind")) {
    case KIND_NOTHING: return "default";
    case KIND_REFUSING: return "not-allowed";
    default: return "pointer";
  }
}

export function drawPointerHighlight(c: CanvasRenderingContext2D): void {
  if (pointerU8("kind") === KIND_NOTHING) return;
  const w = pointerF32("w");
  const h = pointerF32("h");
  if (w <= 0 || h <= 0) return;
  roundRect(
    c, pointerF32("x"), pointerF32("y"),
    w, h, T.RADIUS_ITEM,
  );
  c.fillStyle = T.HOVER_ROW;
  c.fill();
}

export function drawPointerTip(c: CanvasRenderingContext2D): void {
  const text = pointerText("tipText");
  if (!text || text.length === 0) return;
  const label = decodeAt(text, 0, text.length);
  if (!label) return;

  const x = pointerF32("tipX");
  const y = pointerF32("tipY");

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
