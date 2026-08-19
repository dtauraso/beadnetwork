import { columnF32, columnI32, columnU8, hasColumn } from "../../../../Buffer/column-values";
import { edgeColumn, ownerCounts } from "../../../../Buffer/column-owners";
import {
  COL_STREAM_EDGE_SX, COL_STREAM_EDGE_SY, COL_STREAM_EDGE_SZ,
  COL_STREAM_EDGE_EX, COL_STREAM_EDGE_EY, COL_STREAM_EDGE_EZ,
  COL_STREAM_EDGE_SRC_NODE_ROW, COL_STREAM_EDGE_DST_NODE_ROW,
  COL_STREAM_EDGE_DELTA_R, COL_STREAM_EDGE_DRAG_ACTIVE,
} from "../../../../Buffer/column-streams-gen";

export interface EdgeAccessor {

  edgeCount: number;

  segment(row: number): [number, number, number, number, number, number];

  srcNodeRow(row: number): number;

  dstNodeRow(row: number): number;

  deltaR(row: number): number;

  dragActive(row: number): boolean;

}

export function getEdgeStreamAccessor(): EdgeAccessor | null {
  const { edges } = ownerCounts();
  if (edges <= 0) return null;
  if (!hasColumn(edgeColumn(0, COL_STREAM_EDGE_SX))) return null;

  return {
    edgeCount: edges,
    segment(row) {
      return [
        columnF32(edgeColumn(row, COL_STREAM_EDGE_SX)),
        columnF32(edgeColumn(row, COL_STREAM_EDGE_SY)),
        columnF32(edgeColumn(row, COL_STREAM_EDGE_SZ)),
        columnF32(edgeColumn(row, COL_STREAM_EDGE_EX)),
        columnF32(edgeColumn(row, COL_STREAM_EDGE_EY)),
        columnF32(edgeColumn(row, COL_STREAM_EDGE_EZ)),
      ];
    },
    srcNodeRow(row) {
      return columnI32(edgeColumn(row, COL_STREAM_EDGE_SRC_NODE_ROW), -1);
    },
    dstNodeRow(row) {
      return columnI32(edgeColumn(row, COL_STREAM_EDGE_DST_NODE_ROW), -1);
    },
    deltaR(row) {
      return columnF32(edgeColumn(row, COL_STREAM_EDGE_DELTA_R));
    },
    dragActive(row) {

      const col = edgeColumn(row, COL_STREAM_EDGE_DRAG_ACTIVE);
      return hasColumn(col) ? columnU8(col) !== 0 : true;
    },
  };
}
