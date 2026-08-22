import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "./flags";
import { overlayFlagVals, startOverlayReads, type OverlayFlagVals } from "./overlay-leaves";

export type { OverlayFlagVals };

export function readOverlayFlags(): OverlayFlagVals {
  startOverlayReads();
  return overlayFlagVals();
}

export function overlayFlag(name: OverlayFlag): boolean {
  return readOverlayFlags()[name];
}

export function overlayFlagSignature(): string {
  const vals = readOverlayFlags();
  return OVERLAY_FLAG_ORDER.map((flag) => (vals[flag] ? "1" : "0")).join("");
}
