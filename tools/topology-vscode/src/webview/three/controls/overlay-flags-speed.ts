// overlay-flags-speed.ts — a row-keyed READ resource over the buffer's Speed column
// (Overlay block). Split out of overlay-flags.ts — see that file's header for the full
// sibling-file list.

import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../scene/view-blocks";
import { readOverlaySpeed } from "../../../schema/buffer-layout";

/** The current playback-speed multiplier (Overlay block's Speed column) — Go-owned
 *  (RunStdinReader's clock/speed edit handler, seeded at load from view/speed.json).
 *  Read-only reflect for the SpeedSlider so it shows the persisted/live value instead of
 *  a local default that snaps back on reload (memory/feedback_reflect_dont_create_store.md).
 *  Returns null if no snapshot has decoded yet. */
export function readPlaybackSpeed(): number | null {
  const blocks = getViewBlocks();
  if (!blocks) return null;
  return readOverlaySpeed(blocks.overlayView);
}

/** React hook: re-renders the caller when the playback speed changes. Returns null until
 *  the first VIEW snapshot lands. */
export function usePlaybackSpeed(): number | null {
  return useSyncExternalStore(subscribeViewBlocks, readPlaybackSpeed, readPlaybackSpeed);
}
