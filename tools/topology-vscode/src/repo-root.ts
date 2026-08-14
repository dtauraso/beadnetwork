import * as fs from "fs";
import * as path from "path";

export function resolveRepoRoot(startDir: string | undefined): string | undefined {
  if (!startDir) return undefined;

  let dir = path.resolve(startDir);
  for (;;) {
    if (fs.existsSync(path.join(dir, ".git"))) return dir;
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }

  return path.resolve(startDir);
}
