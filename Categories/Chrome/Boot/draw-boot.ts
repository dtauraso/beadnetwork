import { canvasFont } from "../canvas-box";
import * as T from "../chrome-theme";
import { stripF32 } from "../Tabs/strip-leaves";

// Go writes the tab strip on its first pass, before anything else is worth
// looking at, so a zero-width strip means Go has not written this scene yet:
// either it is still being built and spawned (the editor runs `go build` on
// every spawn, and a spawn after a Go edit is a cold one), or it died.
//
// Without this the window is simply BLANK for that whole stretch, which reads
// as a broken editor rather than one that is starting. Nothing here is domain
// state: it is drawn from the absence of the one block file Go always writes.
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
