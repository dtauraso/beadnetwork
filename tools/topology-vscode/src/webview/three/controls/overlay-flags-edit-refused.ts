// overlay-flags-edit-refused.ts — a row-keyed READ resource over the buffer's
// EditRefused column (Overlay block). Split out of overlay-flags.ts — see that file's
// header for the full sibling-file list.

import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../scene/view-blocks";
import { readOverlayEditRefused } from "../../../schema/buffer-layout";

/** Decode how many structural edits Go has REFUSED this run (Overlay block EditRefused).
 *  A count, not a flag: a second refusal has to be distinguishable from the first, or making
 *  the same mistake twice looks like the editor ignoring you. 0 before the first snapshot. */
export function readEditRefused(): number {
  const blocks = getViewBlocks();
  if (!blocks) return 0;
  return readOverlayEditRefused(blocks.overlayView);
}

/** React hook: re-renders the caller each time Go refuses a structural edit. */
export function useEditRefused(): number {
  return useSyncExternalStore(subscribeViewBlocks, readEditRefused, readEditRefused);
}
