import type { BufferLabelPos } from "../buffer-scene-shared";
import { postLog } from "../../../log/post";

const REPORT_ABOVE_PX = 1.5;

let moving = false;
let peak = 0;
let frames = 0;
let sequence = 0;

export function probeLabelLag(
  wanted: BufferLabelPos[], shown: Map<number, { px: number; py: number }>,
): void {
  if (shown.size === 0) return;

  let worst = 0;
  let worstRow = -1;
  for (const b of wanted) {
    const a = shown.get(b.row);
    if (!a) continue;
    const d = Math.hypot(a.px - b.px, a.py - b.py);
    if (d > worst) {
      worst = d;
      worstRow = b.row;
    }
  }

  if (worst > REPORT_ABOVE_PX) {
    if (!moving) {
      moving = true;
      peak = 0;
      frames = 0;
      sequence++;
    }
    frames++;
    if (worst > peak) peak = worst;
    postLog("label-lag", {
      sequence,
      frame: frames,
      row: worstRow,
      behindPx: worst.toFixed(2),
      peakPx: peak.toFixed(2),
    });
    return;
  }

  if (moving) {
    moving = false;
    postLog("label-lag-settled", {
      sequence,
      motionFrames: frames,
      peakPx: peak.toFixed(2),
      restingPx: worst.toFixed(2),
    });
  }
}
