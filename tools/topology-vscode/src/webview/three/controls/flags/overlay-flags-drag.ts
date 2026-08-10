// overlay-flags-drag.ts — a row-keyed READ resource over the buffer's DragNodeRow column
// (Overlay block) and the dragged node's resolved human name. Split out of
// overlay-flags.ts — see that file's header for the full sibling-file list.

import { useSyncExternalStore } from "react";
import { getNodeFrame, subscribeNodeStreamBlocks } from "../../scene/node-stream-blocks";
import { getViewBlocks, subscribeViewBlocks } from "../../scene/view-blocks";
import { readOverlayDragNodeRow } from "../../../../schema/buffer-layout";
import { nodeLabel } from "../../decode/buffer-decode-node";

/** Decode the row index of the node currently being dragged (Overlay block
 *  DragNodeRow column, Go's gesture FSM g.dragNode resolved via NodeRowFor), or -1
 *  when idle. Returns -1 if no snapshot / decode failure yet. */
export function readDragNodeRow(): number {
  const blocks = getViewBlocks();
  if (!blocks) return -1;
  return readOverlayDragNodeRow(blocks.overlayView);
}

/** React hook: re-renders the caller when the dragged node's row changes (drag
 *  start/end). Returns -1 when no drag is in progress. */
export function useDragNodeRow(): number {
  return useSyncExternalStore(subscribeViewBlocks, readDragNodeRow, readDragNodeRow);
}

/** Decode the human name of the node currently being dragged, by resolving
 *  DragNodeRow against the Node block's own Label section — identity rides row
 *  index, the name is never sidecar'd (same plumbing readAbcDragRows uses per-row).
 *  Returns "" when idle or the row can't be resolved yet (no node frame decoded). */
export function readDraggedNodeName(): string {
  const row = readDragNodeRow();
  if (row < 0) return "";
  const decoded = getNodeFrame();
  if (!decoded || row >= decoded.nodeCount) return "";
  return nodeLabel(decoded, row);
}

/** subscribeDraggedNodeName subscribes to BOTH the VIEW stream (DragNodeRow lives in
 *  the Overlay block) and the node stream (the Label section lives in the Node block)
 *  — either arrival can change the resolved name (drag start/end flips the row; a
 *  node-frame arrival while dragging can (re)decode its label). */
function subscribeDraggedNodeName(fn: () => void): () => void {
  const unsubView = subscribeViewBlocks(fn);
  const unsubNode = subscribeNodeStreamBlocks(fn);
  return () => {
    unsubView();
    unsubNode();
  };
}

/** React hook: re-renders the caller when the dragged node's name changes (drag
 *  start/end, or a node-frame arrival while dragging). Returns "" when idle. */
export function useDraggedNodeName(): string {
  return useSyncExternalStore(subscribeDraggedNodeName, readDraggedNodeName, readDraggedNodeName);
}
