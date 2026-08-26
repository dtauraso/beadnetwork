import * as path from "path";

export function sceneRoots(anchorPath: string): string[] {
  return [path.dirname(anchorPath)];
}
