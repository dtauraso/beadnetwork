import { roundRect } from "../../canvas-box";
import * as T from "../../chrome-theme";
import { chipF32 } from "./chip-leaves";

export function fitChipKey(): string {
  return [chipF32("x"), chipF32("y"), chipF32("w")].join(",");
}

export function drawFitChip(c: CanvasRenderingContext2D): void {
  const w = chipF32("w");
  const h = chipF32("h");
  if (w <= 0 || h <= 0) return;
  const x = chipF32("x");
  const y = chipF32("y");

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  const s = Math.min(w, h) * 0.56;
  const cx = x + w / 2;
  const left = cx - s / 2, right = cx + s / 2;
  const top = y + (h - s) / 2, bottom = top + s;
  const eave = top + s * 0.45;
  const wall = s * 0.14;

  c.strokeStyle = T.TEXT;
  c.lineWidth = 1.5;
  c.lineJoin = "round";
  c.beginPath();
  c.moveTo(left, eave);
  c.lineTo(cx, top);
  c.lineTo(right, eave);
  c.moveTo(left + wall, eave - wall * 0.9);
  c.lineTo(left + wall, bottom);
  c.lineTo(right - wall, bottom);
  c.lineTo(right - wall, eave - wall * 0.9);
  c.stroke();
}
