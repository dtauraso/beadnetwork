// overlay-toggle.ts — the ToggleCfg shape and the pure rules for reading/firing one: what an
// overlay row's value is (Go-owned buffer flag, default-until-first-snapshot fallback) and
// what firing it sends (the same per-flag toggle record everywhere). No JSX.

import { postGoRecord } from "../../vscode-api";
import { encodeOverlaysToggle } from "../../../schema/input-encode";
import type { OverlayFlag } from "../../../messages";
import { postLog } from "../../log/post";
import { useOverlayFlags } from "./overlay-flags";

export type ToggleCfg = {
  flag: OverlayFlag;
  /** Initial value shown before the first buffer snapshot lands (store polarity).
   *  OMIT this for a flag whose Go-owned default is not something TS should assert —
   *  useToggleVal then falls back to `false` (render null-until-first-snapshot, same as
   *  NavGuides.tsx's `bufFlags?.x ?? false`) instead of TS authoring a stand-in value. */
  default?: boolean;
  /** Compute active (highlight) from the raw value. */
  active: (val: boolean) => boolean;
  /** The row's glyph, kept OUT of the label. It used to be part of the label string, which
   *  made its position a per-row accident (`select ⬡` trailed while the rest led) and left
   *  it in the same text run as the words — so a label that wrapped put its second line
   *  under the icon. As its own field it is always rendered first, in its own column, and
   *  the words wrap under the words. A function of the raw value for a glyph that reports
   *  state (labels' ▴/▾). */
  icon: string | ((val: boolean) => string);
  /** Label WORDS ONLY — no glyph (see `icon`). String or function of raw value. */
  label: string | ((val: boolean) => string);
  /** Title string function of active value. */
  title: (active: boolean) => string;
  /** postLog payload factory. */
  payload: (val: boolean) => Record<string, unknown>;
};

export function fireToggle(cfg: ToggleCfg, val: boolean) {
  postLog("guide-btn-click", cfg.payload(val));
  postGoRecord(encodeOverlaysToggle(cfg.flag));
}

/** The value a toggle displays: the Go-owned Overlay buffer columns (the only live truth).
 *  cfg.flag keys the buffer record in store polarity. Falls back to cfg.default until the
 *  first snapshot lands. */
export function useToggleVal(cfg: ToggleCfg): boolean {
  return toggleVal(useOverlayFlags(), cfg);
}

/** useToggleVal's rule as a pure function, so a caller that already holds the flags (a group
 *  header counting its members) reads them by the SAME rule instead of restating it — two
 *  copies could disagree about the pre-first-snapshot fallback and only one would be right. */
export function toggleVal(bufFlags: ReturnType<typeof useOverlayFlags>, cfg: ToggleCfg): boolean {
  // ?? cfg.default only guards the (impossible) missing-key case under noUncheckedIndexedAccess;
  // every OverlayFlag is always present in the record, so `false` is preserved. When cfg omits
  // `default` (a flag whose Go default TS should not assert), this falls back to `false` until
  // the first snapshot lands — the same "null-until-first-snapshot" shape as NavGuides.tsx.
  if (bufFlags) return bufFlags[cfg.flag] ?? cfg.default ?? false;
  return cfg.default ?? false;
}
