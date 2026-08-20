import * as fs from "fs";
import * as path from "path";
import { SCENES } from "../../Scene/scenes-gen";

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

export function resolveScenePath(anchorPath: string): string {
  let selected: string;
  try {
    const raw = fs.readFileSync(path.join(anchorPath, "view", "scene", "selected.json"), "utf8");
    const parsed: unknown = JSON.parse(raw);
    selected = typeof parsed === "string" ? parsed : "";
  } catch {
    selected = "";
  }

  const scene = SCENES.find((s) => s.name === selected) ?? SCENES[0];
  if (!scene) return anchorPath;

  const container = path.dirname(anchorPath);
  return path.join(container, scene.dir);
}

export function readCounts(anchorPath: string): { nodes: number; edges: number } {
  const scenePath = resolveScenePath(anchorPath);
  return { nodes: readCount(scenePath, "nodes"), edges: readCount(scenePath, "edges") };
}
