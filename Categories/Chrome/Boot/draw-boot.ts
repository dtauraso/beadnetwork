import { canvasFont } from "../canvas-box";
import * as T from "../chrome-theme";
import { stripF32 } from "../Tabs/strip-leaves";

export function booting(): boolean {
  return !(stripF32("stripW") > 0);
}

export function bootKey(): string {
  return booting() ? "booting" : "";
}

const FONT_PX = 13;

export function drawBoot(c: CanvasRenderingContext2D): void {
  if (!booting()) return;

  c.font = canvasFont(FONT_PX);
  c.fillStyle = T.TEXT;
  c.textAlign = "left";
  c.textBaseline = "top";
  c.fillText("starting the network…", 12, 12);
}
