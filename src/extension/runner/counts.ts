import * as fs from "fs";
import * as path from "path";
import { SCENES } from "../../Scene/scenes-gen";

function readIntLeaf(leafPath: string): number {
  let raw: Buffer;
  try {
    raw = fs.readFileSync(leafPath);
  } catch (err) {
    throw new Error(`cannot read ${leafPath}: ${err instanceof Error ? err.message : String(err)}`);
  }
  if (raw.length !== 8) {
    throw new Error(`${leafPath} must be an 8-byte integer leaf, got ${raw.length} bytes`);
  }
  const value = Number(new DataView(raw.buffer, raw.byteOffset, raw.length).getBigInt64(0, true));
  if (!Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${leafPath} must be a non-negative integer, got: ${value}`);
  }
  return value;
}

function readCount(topologyPath: string, name: "nodes" | "edges"): number {
  return readIntLeaf(path.join(topologyPath, "counts", `${name}.bin`));
}

export function resolveScenePath(anchorPath: string): string {
  let selected: string;
  try {
    selected = fs.readFileSync(path.join(anchorPath, "view", "scene", "selected.bin"), "utf8");
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
