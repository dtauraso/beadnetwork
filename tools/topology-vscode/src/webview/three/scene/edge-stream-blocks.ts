import { getLatestEdgeStreamFrames } from "../../snapshot-buffer";
import { decodeEdgeStreamFrame, type DecodedEdgeStreamFrame } from "../decode/buffer-decode-edge";
import {
  readEdgeSX, readEdgeSY, readEdgeSZ, readEdgeEX, readEdgeEY, readEdgeEZ,
} from "../../../schema/buffer-layout";

export interface EdgeAccessor {

  edgeCount: number;

  segment(row: number): [number, number, number, number, number, number];

}

function decodedFor(frames: ReadonlyMap<number, ArrayBuffer>, row: number): DecodedEdgeStreamFrame | null {
  const buf = frames.get(row);
  return buf ? decodeEdgeStreamFrame(row, buf) : null;
}

export function getEdgeStreamAccessor(): EdgeAccessor | null {
  const frames = getLatestEdgeStreamFrames();
  if (frames.size === 0) return null;
  let maxRow = -1;
  for (const r of frames.keys()) if (r > maxRow) maxRow = r;
  const edgeCount = maxRow + 1;
  return {
    edgeCount,
    segment(row) {
      const d = decodedFor(frames, row);
      if (!d) return [0, 0, 0, 0, 0, 0];
      return [
        readEdgeSX(d.edgeView, 0), readEdgeSY(d.edgeView, 0), readEdgeSZ(d.edgeView, 0),
        readEdgeEX(d.edgeView, 0), readEdgeEY(d.edgeView, 0), readEdgeEZ(d.edgeView, 0),
      ];
    },
  };
}
