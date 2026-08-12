import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import { readOverlaySpeed } from "../../../../schema/buffer-layout";

export function readPlaybackSpeed(): number | null {
  const blocks = getViewBlocks();
  if (!blocks) return null;
  return readOverlaySpeed(blocks.overlayView);
}

export function usePlaybackSpeed(): number | null {
  return useSyncExternalStore(subscribeViewBlocks, readPlaybackSpeed, readPlaybackSpeed);
}
