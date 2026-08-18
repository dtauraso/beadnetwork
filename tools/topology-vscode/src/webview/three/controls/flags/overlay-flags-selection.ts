import { useSyncExternalStore } from "react";
import { getNodeFrame } from "../../scene/nodes/node-frame-aggregate";
import { subscribeViewBlocks } from "../../scene/view-blocks";
import { nodeSelected } from "../../../../../Node/node-frame";

export function readSelectedNodeRow(): number {
  const decoded = getNodeFrame();
  if (!decoded) return -1;
  for (let i = 0; i < decoded.nodeCount; i++) {
    if (nodeSelected(decoded.nodeView, i)) return i;
  }
  return -1;
}

export function useSelectedNodeRow(): number {
  return useSyncExternalStore(subscribeViewBlocks, readSelectedNodeRow, readSelectedNodeRow);
}
