import { useSyncExternalStore } from "react";
import { columnU8 } from "../../../../../Buffer/column-values";
import { subscribeFrame } from "../../../frame-tick";
import { nodeColumn, ownerCounts } from "../../../../../Buffer/column-owners";
import { COL_STREAM_NODE_SELECTED } from "../../../../../Buffer/column-streams-gen";

export function readSelectedNodeRow(): number {
  const { nodes } = ownerCounts();
  for (let i = 0; i < nodes; i++) {
    if (columnU8(nodeColumn(i, COL_STREAM_NODE_SELECTED))) return i;
  }
  return -1;
}

export function useSelectedNodeRow(): number {
  return useSyncExternalStore(subscribeFrame, readSelectedNodeRow, readSelectedNodeRow);
}
