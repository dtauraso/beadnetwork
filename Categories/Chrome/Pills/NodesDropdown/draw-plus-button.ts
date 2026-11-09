import { roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import * as T from "../../chrome-theme";
import { pillF32, pillText } from "./pill-leaves";

function textOf(name: "plusFill" | "plusStroke"): string {
  const t = pillText(name);
  return t ? decodeAt(t, 0, t.length) : "";
}

export function drawPlusButton(c: CanvasRenderingContext2D): void {
  const w = pillF32("plusW");
  const h = pillF32("plusH");
  if (w <= 0 || h <= 0) return;
  const x = pillF32("plusX");
  const y = pillF32("plusY");

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  const fill = textOf("plusFill") || "#e3f2fd";
  const stroke = textOf("plusStroke") || "#1565c0";
  const s = Math.min(w, h) * 0.46;
  const nx = x + (w - s) / 2 - s * 0.08;
  const ny = y + (h - s) / 2 + s * 0.08;

  c.save();
  c.shadowColor = stroke;
  c.shadowBlur = s * 0.5;
  roundRect(c, nx, ny, s, s, s * 0.28);
  c.fillStyle = fill;
  c.fill();
  c.restore();

  const r = s * 0.3;
  const bx = nx + s;
  const by = ny;
  c.beginPath();
  c.arc(bx, by, r, 0, Math.PI * 2);
  c.fillStyle = stroke;
  c.fill();
  c.strokeStyle = fill;
  c.lineWidth = Math.max(1, r * 0.35);
  c.lineCap = "round";
  c.beginPath();
  c.moveTo(bx - r * 0.55, by);
  c.lineTo(bx + r * 0.55, by);
  c.moveTo(bx, by - r * 0.55);
  c.lineTo(bx, by + r * 0.55);
  c.stroke();
}
