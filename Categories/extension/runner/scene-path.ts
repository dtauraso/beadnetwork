import * as fs from "fs";
import * as path from "path";
import { SCENES } from "../../Scene/scenes-gen";

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
