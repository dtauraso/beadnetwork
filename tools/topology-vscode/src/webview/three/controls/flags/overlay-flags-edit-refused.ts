



import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import { readOverlayEditRefused } from "../../../../schema/buffer-layout";


export function readEditRefused(): number {
  const blocks = getViewBlocks();
  if (!blocks) return 0;
  return readOverlayEditRefused(blocks.overlayView);
}


export function useEditRefused(): number {
  return useSyncExternalStore(subscribeViewBlocks, readEditRefused, readEditRefused);
}
