import type { BufferLabelPos } from "../buffer-scene-shared";
import { postLog } from "../../../log/post";

const REPORT_ABOVE_PX = 1.5;

let moving = false;
let peak = 0;
let frames = 0;
let sequence = 0;

export function probeLabelLag(shown: BufferLabelPos[], wanted: BufferLabelPos[]): void {
  if (shown.length === 0 || shown.length !== wanted.length) return;

  let worst = 0;
  let worstRow = -1;
  for (let i = 0; i < wanted.length; i++) {
    const a = shown[i];
    const b = wanted[i];
    if (!a || !b || a.row !== b.row) continue;
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
      framesBehind: frames,
      peakPx: peak.toFixed(2),
      restingPx: worst.toFixed(2),
    });
  }
}
