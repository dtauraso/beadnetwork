import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "../Input/messages";
import { makeLeafValues } from "../webview/leaf-values";

export type OverlayFlagVals = Record<OverlayFlag, boolean>;

const values = makeLeafValues<OverlayFlag>("Overlay/paths", OVERLAY_FLAG_ORDER);

const vals = Object.fromEntries(OVERLAY_FLAG_ORDER.map((f) => [f, false])) as OverlayFlagVals;

export function startOverlayReads(): void {
  values.u8(OVERLAY_FLAG_ORDER[0]);
}

export function overlayFlagVals(): OverlayFlagVals {
  for (const flag of OVERLAY_FLAG_ORDER) vals[flag] = values.u8(flag) !== 0;
  return vals;
}
