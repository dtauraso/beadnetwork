import { getLatestNodeStreamFrames, getNodeStreamVersion } from "../../../snapshot-buffer";
import { decodeNodeStreamFrame } from "../../decode/buffer-decode-node";
import { readNodeKindId } from "../../../../schema/buffer-layout/buffer-layout";
import { NODE_KIND_NAMES } from "../../../../schema/node-defs";

const COMM_KIND = "Input";

let lastVersion = -1;
let lastRows: ReadonlySet<number> = new Set();

export function getCommNodeRows(): ReadonlySet<number> {
  const nv = getNodeStreamVersion();
  if (nv === lastVersion) return lastRows;

  const rows = new Set<number>();
  for (const [row, buf] of getLatestNodeStreamFrames()) {
    const decoded = decodeNodeStreamFrame(row, buf);
    if (!decoded) continue;
    if (NODE_KIND_NAMES[readNodeKindId(decoded.nodeView, 0)] === COMM_KIND) rows.add(row);
  }
  lastVersion = nv;
  lastRows = rows;
  return lastRows;
}
