import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import { readOverlaySceneEditable, readOverlaySceneKinds } from "../../../../../Buffer/buffer-layout";

export function readSceneEditable(): boolean {
  const blocks = getViewBlocks();
  if (!blocks) return false;
  return readOverlaySceneEditable(blocks.overlayView) !== 0;
}

export function useSceneEditable(): boolean {
  return useSyncExternalStore(subscribeViewBlocks, readSceneEditable, readSceneEditable);
}

export function readSceneKinds(): number {
  const blocks = getViewBlocks();
  if (!blocks) return 0;
  return readOverlaySceneKinds(blocks.overlayView);
}

export function useSceneKinds(): number {
  return useSyncExternalStore(subscribeViewBlocks, readSceneKinds, readSceneKinds);
}
