import { ownerCounts } from "../../Buffer/column-owners";
import { edgeBytes, edgeF32, edgeI32, edgeU8 } from "./edge-leaves";

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
  if (!edgeBytes(0, "sx")) return null;

  return {
    edgeCount: edges,
    segment(row) {
      return [
        edgeF32(row, "sx"),
        edgeF32(row, "sy"),
        edgeF32(row, "sz"),
        edgeF32(row, "ex"),
        edgeF32(row, "ey"),
        edgeF32(row, "ez"),
      ];
    },
    srcNodeRow(row) {
      return edgeI32(row, "srcNodeRow", -1);
    },
    dstNodeRow(row) {
      return edgeI32(row, "dstNodeRow", -1);
    },
    deltaR(row) {
      return edgeF32(row, "deltaR");
    },
    dragActive(row) {

      return edgeBytes(row, "dragActive") ? edgeU8(row, "dragActive") !== 0 : true;
    },
  };
}
