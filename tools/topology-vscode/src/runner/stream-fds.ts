






export const VIEW_FD = 4;







export const EDGE_BASE_FD = 5;






export const MAX_EDGE_STREAMS = 256;






















export function nodeIdForRow(row: number): number {
  return row + 1;
}
export function rowForNodeId(nodeId: number): number {
  return nodeId - 1;
}




export const MAX_NODE_STREAMS = 256;








export const DRIVE_SLOTS_PER_NODE = 2;
