import { columnU8 } from "../../schema/buffer-layout/column-values";
import { nodeColumn, ownerCounts } from "../../schema/buffer-layout/column-owners";
import {
  COL_STREAM_NODE_SELECTED,
} from "../../Node/columns-gen";

export function readSelectedNodeRow(): number {
  const { nodes } = ownerCounts();
  for (let i = 0; i < nodes; i++) {
    if (columnU8(nodeColumn(i, COL_STREAM_NODE_SELECTED))) return i;
  }
  return -1;
}
