import { canvasFont, roundRect } from "../../../webview/canvas-box";
import { decodeAt } from "../../../Buffer/column-reads";
import * as T from "../../../webview/canvas-theme";
import { chipF32, chipText } from "./chip-leaves";

export function fitChipKey(): string {
  return [chipF32("x"), chipF32("y"), chipF32("w")].join(",");
}

export function drawFitChip(c: CanvasRenderingContext2D): void {
  const label = chipText("labelText");
  const w = chipF32("w");
  const h = chipF32("h");
  if (!label || w <= 0 || h <= 0) return;
  const x = chipF32("x");
  const y = chipF32("y");

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
