// nodeIdForRow/rowForNodeId: the executable form of "fd = base + nodeRow, ROW ID = NODE
// ID - 1" (see runCommand.ts's doc comment on these functions and NODE_BASE_FD above it).
// Pure arithmetic — no vscode/child_process mocking needed.

import { describe, it, expect } from "vitest";
import { nodeIdForRow, rowForNodeId } from "../src/runCommand";

describe("nodeIdForRow / rowForNodeId", () => {
  it("row 0 is node 1 (1-based node ids, 0-based rows)", () => {
    expect(nodeIdForRow(0)).toBe(1);
    expect(rowForNodeId(1)).toBe(0);
  });

  it("is the inverse in both directions across a range of rows", () => {
    for (let row = 0; row < 20; row++) {
      expect(rowForNodeId(nodeIdForRow(row))).toBe(row);
    }
    for (let nodeId = 1; nodeId <= 20; nodeId++) {
      expect(nodeIdForRow(rowForNodeId(nodeId))).toBe(nodeId);
    }
  });

  it("a gap (an idle pipe / deleted node) does not shift the mapping for other rows", () => {
    // Node 3 is deleted, leaving row 2 idle — rows on either side keep their own ids
    // unchanged; nothing shifts to fill the gap (persistence-ownership.md: "a gap row
    // still needs its own dedicated node/interior/drive fds allocated").
    const rowsPresent = [0, 1, 3, 4]; // row 2 (node 3) is the gap
    expect(rowsPresent.map(nodeIdForRow)).toEqual([1, 2, 4, 5]);
  });
});
