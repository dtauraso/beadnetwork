// overlay-flags-scene.ts — a row-keyed READ resource over the buffer's SceneEditable and
// SceneKinds columns (Overlay block). Split out of overlay-flags.ts — see that file's
// header for the full sibling-file list.

import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import { readOverlaySceneEditable, readOverlaySceneKinds } from "../../../../schema/buffer-layout";

/** Decode whether THIS scene can be structurally edited (Overlay block SceneEditable —
 *  SceneTab.Editable, Go's own per-scene property). false before the first snapshot: a
 *  palette that appears for an instant in a scene that cannot take one is worse than a
 *  palette that appears a frame late. */
export function readSceneEditable(): boolean {
  const blocks = getViewBlocks();
  if (!blocks) return false;
  return readOverlaySceneEditable(blocks.overlayView) !== 0;
}

/** React hook: whether this scene can be structurally edited. */
export function useSceneEditable(): boolean {
  return useSyncExternalStore(subscribeViewBlocks, readSceneEditable, readSceneEditable);
}

/** Decode the BITMASK of kind ids this scene accepts (Overlay block SceneKinds — bit N = the
 *  kind whose KindId is N). 0 before the first snapshot, which offers nothing: a palette that
 *  briefly offers kinds this scene has no place for is worse than one that appears a frame
 *  late. */
export function readSceneKinds(): number {
  const blocks = getViewBlocks();
  if (!blocks) return 0;
  return readOverlaySceneKinds(blocks.overlayView);
}

/** React hook: the scene's accepted-kind mask. */
export function useSceneKinds(): number {
  return useSyncExternalStore(subscribeViewBlocks, readSceneKinds, readSceneKinds);
}
