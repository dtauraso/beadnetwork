import { ownerCounts } from "../../read/owner-counts";
import { nodeU8 } from "../../../Node/Block/read/node-leaves";

export function readSelectedNodeRow(): number {
  const { nodes } = ownerCounts();
  for (let i = 0; i < nodes; i++) {
    if (nodeU8(i, "selected")) return i;
  }
  return -1;
}
