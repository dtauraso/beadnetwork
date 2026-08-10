// overlay-flags-selection.ts — a row-keyed READ resource over the Node block's Selected
// column. Split out of overlay-flags.ts — see that file's header for the full
// sibling-file list.

import { useSyncExternalStore } from "react";
import { getNodeFrame } from "../../scene/node-stream-blocks";
import { subscribeViewBlocks } from "../../scene/view-blocks";
import { readNodeSelected } from "../../../../schema/buffer-layout";

/** Decode the SELECTED node's buffer row, or -1 when nothing is selected. Selection is
 *  Go-owned (the Node block's Selected column); this is a second READER of that truth, never
 *  a cache of it — which is what lets the delete key forward a row without TS ever deciding
 *  what is selected. */
export function readSelectedNodeRow(): number {
  const decoded = getNodeFrame();
  if (!decoded) return -1;
  for (let i = 0; i < decoded.nodeCount; i++) {
    if (readNodeSelected(decoded.nodeView, i)) return i;
  }
  return -1;
}

/** React hook: the selected node's row, or -1. */
export function useSelectedNodeRow(): number {
  return useSyncExternalStore(subscribeViewBlocks, readSelectedNodeRow, readSelectedNodeRow);
}
