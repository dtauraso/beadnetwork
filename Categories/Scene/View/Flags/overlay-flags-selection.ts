import { ownerCounts } from "../../owner-counts";
import { nodeU8 } from "../../../Node/node-leaves";

export function readSelectedNodeRow(): number {
  const { nodes } = ownerCounts();
  for (let i = 0; i < nodes; i++) {
    if (nodeU8(i, "selected")) return i;
  }
  return -1;
}
