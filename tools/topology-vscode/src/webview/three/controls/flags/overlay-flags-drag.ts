import { useSyncExternalStore } from "react";
import { columnI32, subscribeColumns } from "../../../../../Buffer/column-values";
import { COL_STREAM_OVERLAY_DRAG_NODE_ROW } from "../../../../../Buffer/column-streams-gen";
import { nodeLabel } from "../../decode/buffer-decode-node";

export function readDragNodeRow(): number {
  return columnI32(COL_STREAM_OVERLAY_DRAG_NODE_ROW, -1);
}

export function useDragNodeRow(): number {
  return useSyncExternalStore(subscribeColumns, readDragNodeRow, readDragNodeRow);
}

export function readDraggedNodeName(): string {
  const row = readDragNodeRow();
  if (row < 0) return "";
  return nodeLabel(row);
}

function subscribeDraggedNodeName(fn: () => void): () => void {
  return subscribeColumns(fn);
}

export function useDraggedNodeName(): string {
  return useSyncExternalStore(subscribeDraggedNodeName, readDraggedNodeName, readDraggedNodeName);
}
