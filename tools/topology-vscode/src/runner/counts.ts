import * as fs from "fs";
import * as path from "path";

// readCounts replaces the old countNodes/countEdges tree-walks (see
// .claude/rules/persistence-ownership.md "Counts are stored, never re-derived"). The ext host must know the fd RANGE
// before spawning Go, so it cannot ask Go for this — but it also must not WALK the tree to
// derive it, because that re-implements the on-disk layout in a second language (step 2's
// near-miss, when countEdges had to be hand-updated in lockstep with a path move). Instead
// the counts are STORED, at `<topologyPath>/counts.json` = `{"nodes": N, "edges": E}`,
// written once by whichever operation changes the node/edge set (today: none does — the
// topology is editor-authored by hand, so this file is hand-maintained alongside it). TS
// reads two numbers and stops knowing the layout entirely.
//
// `nodes` means the ROW COUNT (the largest node id in the tree, ROW ID = NODE ID - 1 —
// nodes/Wiring's `topoSpec.RowCount`), NOT how many node directories exist. A deleted node
// leaves a gap row rather than shrinking the row space, and that gap row still needs its
// own dedicated node/interior/drive fds allocated, so sizing this from a live-node count
// would under-allocate. `edges` has no id space to gap and stays a plain count.
//
// Unlike the old countNodes/countEdges, which returned 0 on any read/parse failure (a
// SILENT failure: 0 edges/nodes just meant no dedicated streams got allocated, degrading
// the bridge invisibly), a missing or malformed counts.json now THROWS. There is no correct
// fallback count to guess, and guessing 0 is exactly the bug this rewrite removes.
export function readCounts(topologyPath: string): { nodes: number; edges: number } {
  const countsPath = path.join(topologyPath, "counts.json");
  let raw: string;
  try {
    raw = fs.readFileSync(countsPath, "utf8");
  } catch (err) {
    throw new Error(`cannot read ${countsPath}: ${err instanceof Error ? err.message : String(err)}`);
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    throw new Error(`cannot parse ${countsPath}: ${err instanceof Error ? err.message : String(err)}`);
  }
  const obj = parsed as { nodes?: unknown; edges?: unknown };
  if (
    !parsed || typeof parsed !== "object" ||
    typeof obj.nodes !== "number" || !Number.isInteger(obj.nodes) || obj.nodes < 0 ||
    typeof obj.edges !== "number" || !Number.isInteger(obj.edges) || obj.edges < 0
  ) {
    throw new Error(`${countsPath} must be {"nodes": <non-negative integer>, "edges": <non-negative integer>}, got: ${raw}`);
  }
  return { nodes: obj.nodes, edges: obj.edges };
}
