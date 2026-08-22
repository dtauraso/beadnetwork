import * as path from "path";
import { SCENES } from "../../Categories/Scene/scenes-gen";

export function sceneRoots(anchorPath: string): string[] {
  const container = path.dirname(anchorPath);
  return SCENES.map((s) => path.join(container, s.dir));
}
