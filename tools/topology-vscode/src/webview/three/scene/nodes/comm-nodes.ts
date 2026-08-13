import { getLatestNodeStreamFrames, getNodeStreamVersion } from "../../../snapshot-buffer";
import { decodeNodeStreamFrame } from "../../decode/buffer-decode-node";
import { readNodeKindId } from "../../../../schema/buffer-layout/buffer-layout";
import { NODE_KIND_NAMES } from "../../../../schema/node-defs";

// COMM_KIND is the node kind whose outgoing edges carry a position rather
// than a value — it tells its neighbours where the constraint it holds puts
// them, and those edges are what the comm-edges overlay draws.
const COMM_KIND = "Input";

let lastVersion = -1;
let lastRows: ReadonlySet<number> = new Set();

// getCommNodeRows is the rows whose OUTGOING edges are communication edges.
// Two renderers ask the same question — one to colour those edges, one to
// leave their beads undrawn — so the answer is worked out once per stream
// version rather than by each of them walking every frame.
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
