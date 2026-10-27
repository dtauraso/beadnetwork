import type { LabelPos } from "../../../Start/extension/webview/scene/scene-tags";
import { canvasFont, roundRect } from "../../Chrome/canvas-box";
import * as T from "../../Chrome/chrome-theme";

const PAD_X = 8;
const PAD_Y = 3;
const RADIUS = 4;
const GAP_ABOVE_ANCHOR = 4;

let positions: LabelPos[] = [];
let epoch = 0;

export interface ArcLabel { parts: { text: string; color: string }[]; px: number; py: number; }

const PART_GAP = 6;

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
    const widths = p.parts.map((part) => c.measureText(part.text).width);
    const textW = widths.reduce((s, v) => s + v, 0) + PART_GAP * Math.max(0, p.parts.length - 1);
    const w = textW + 2 * PAD_X;
    const x = p.px - w / 2;
    const y = p.py - h / 2;
    roundRect(c, x, y, w, h, RADIUS);
    c.fillStyle = T.CHIP;
    c.fill();
    c.strokeStyle = T.BORDER;
    c.lineWidth = 1;
    c.stroke();
    let tx = x + PAD_X;
    p.parts.forEach((part, i) => {
      c.fillStyle = part.color;
      c.fillText(part.text, tx, y + h / 2);
      tx += (widths[i] ?? 0) + PART_GAP;
    });
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
