import type { BufferLabelPos } from "../../webview/three/scene/buffer-scene-shared";
import { panelFont, roundRect } from "../../PanelOverlay/panel-box";
import * as T from "../../webview/three/controls/chrome-theme";

const PAD_X = 8;
const PAD_Y = 3;
const RADIUS = 4;
const GAP_ABOVE_ANCHOR = 4;

let positions: BufferLabelPos[] = [];
let epoch = 0;

export function setLabelPositions(next: BufferLabelPos[]): void {
  positions = next;
  epoch++;
}

export function labelEpoch(): number {
  return epoch;
}

export function drawLabels(c: CanvasRenderingContext2D): void {
  if (positions.length === 0) return;

  c.font = panelFont(T.FONT_SIZE);
  c.textAlign = "left";
  c.textBaseline = "middle";

  const h = T.FONT_SIZE * 1.25 + 2 * PAD_Y;
  for (const p of positions) {
    const text = p.label || String(p.row);
    const w = c.measureText(text).width + 2 * PAD_X;
    const x = p.px - w / 2;
    const y = p.py - GAP_ABOVE_ANCHOR - h;

    roundRect(c, x, y, w, h, RADIUS);
    c.fillStyle = T.CHIP;
    c.fill();
    c.strokeStyle = T.BORDER;
    c.lineWidth = 1;
    c.stroke();

    c.fillStyle = T.TEXT;
    c.fillText(text, x + PAD_X, y + h / 2);
  }
}
