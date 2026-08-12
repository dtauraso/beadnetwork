import * as fs from "fs";
import * as path from "path";

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
