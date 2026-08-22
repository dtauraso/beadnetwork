import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "./flags";
import { makeLeafValues } from "../valuefile/leaf-values";
import { OVERLAY_FLAG_DEFAULTS } from "./flag-defaults-gen";

export type OverlayFlagVals = Record<OverlayFlag, boolean>;

const values = makeLeafValues<OverlayFlag>("Overlay/paths", OVERLAY_FLAG_ORDER);

const vals = Object.fromEntries(
  OVERLAY_FLAG_ORDER.map((f) => [f, OVERLAY_FLAG_DEFAULTS[f] ?? false]),
) as OverlayFlagVals;

export function startOverlayReads(): void {
  values.u8(OVERLAY_FLAG_ORDER[0]);
}

export function overlayFlagVals(): OverlayFlagVals {
  for (const flag of OVERLAY_FLAG_ORDER) {
    const b = values.bytes(flag);
    vals[flag] = b === undefined ? (OVERLAY_FLAG_DEFAULTS[flag] ?? false) : b.getUint8(0) !== 0;
  }
  return vals;
}
