import * as fs from "fs";
import * as path from "path";

function readCount(topologyPath: string, name: "nodes" | "edges"): number {
  const countPath = path.join(topologyPath, "counts", `${name}.json`);
  let raw: string;
  try {
    raw = fs.readFileSync(countPath, "utf8");
  } catch (err) {
    throw new Error(`cannot read ${countPath}: ${err instanceof Error ? err.message : String(err)}`);
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    throw new Error(`cannot parse ${countPath}: ${err instanceof Error ? err.message : String(err)}`);
  }
  if (typeof parsed !== "number" || !Number.isInteger(parsed) || parsed < 0) {
    throw new Error(`${countPath} must be a single non-negative integer, got: ${raw}`);
  }
  return parsed;
}

export function readCounts(topologyPath: string): { nodes: number; edges: number } {
  return { nodes: readCount(topologyPath, "nodes"), edges: readCount(topologyPath, "edges") };
}
