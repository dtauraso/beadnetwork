import { postGoRecord } from "../../../vscode-api";
import { encodeOverlaysToggle } from "../../../../schema/input-encode";
import type { OverlayFlag } from "../../../../messages";
import { postLog } from "../../../log/post";
import { useOverlayFlags } from "../flags/overlay-flags";

export type ToggleCfg = {
  flag: OverlayFlag;

  default?: boolean;

  active: (val: boolean) => boolean;

  icon: string | ((val: boolean) => string);

  label: string | ((val: boolean) => string);

  title: (active: boolean) => string;

  payload: (val: boolean) => Record<string, unknown>;
};

export function fireToggle(cfg: ToggleCfg, val: boolean) {
  postLog("guide-btn-click", cfg.payload(val));
  postGoRecord(encodeOverlaysToggle(cfg.flag));
}

export function useToggleVal(cfg: ToggleCfg): boolean {
  return toggleVal(useOverlayFlags(), cfg);
}

export function toggleVal(bufFlags: ReturnType<typeof useOverlayFlags>, cfg: ToggleCfg): boolean {

  if (bufFlags) return bufFlags[cfg.flag] ?? cfg.default ?? false;
  return cfg.default ?? false;
}
