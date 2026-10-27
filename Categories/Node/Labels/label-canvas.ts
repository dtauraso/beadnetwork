import type { LabelPos } from "../../../Start/extension/webview/scene/scene-tags";
import { canvasFont, roundRect } from "../../Chrome/canvas-box";
import * as T from "../../Chrome/chrome-theme";

const PAD_X = 8;
const PAD_Y = 3;
const RADIUS = 4;
const GAP_ABOVE_ANCHOR = 4;

let positions: LabelPos[] = [];
let epoch = 0;

export interface ArcLabel { text: string; color: string; px: number; py: number; }

let arcLabels: ArcLabel[] = [];

export function setLabelPositions(next: LabelPos[]): void {
  positions = next;
  epoch++;
}

export function setArcLabels(next: ArcLabel[]): void {
  if (next.length === 0 && arcLabels.length === 0) return;
  arcLabels = next;
  epoch++;
}

function drawArcLabels(c: CanvasRenderingContext2D): void {
  if (arcLabels.length === 0) return;
  c.font = canvasFont(T.FONT_SIZE);
  c.textAlign = "left";
  c.textBaseline = "middle";
  const h = T.FONT_SIZE * 1.25 + 2 * PAD_Y;
  for (const p of arcLabels) {
    const w = c.measureText(p.text).width + 2 * PAD_X;
    const x = p.px - w / 2;
    const y = p.py - h / 2;
    roundRect(c, x, y, w, h, RADIUS);
    c.fillStyle = T.CHIP;
    c.fill();
    c.strokeStyle = p.color;
    c.lineWidth = 1.5;
    c.stroke();
    c.fillStyle = p.color;
    c.fillText(p.text, x + PAD_X, y + h / 2);
  }
}

export function labelEpoch(): number {
  return epoch;
}

export function drawLabels(c: CanvasRenderingContext2D): void {
  drawArcLabels(c);
  if (positions.length === 0) return;

  c.font = canvasFont(T.FONT_SIZE);
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
