import * as fs from "fs";
import * as path from "path";

// resolveRepoRoot walks UP from a starting directory to the enclosing git root.
//
// The workspace folder is not the repo root and must not be used as one: open
// the window on tools/topology-vscode — a perfectly ordinary thing to do while
// working on the extension — and `go build .` runs there and fails with
// "no Go files in .../tools/topology-vscode", leaving the editor with no Go
// backend at all. The same wrong root also scattered a second .probe/ log dir
// inside the subdirectory.
//
// A worktree's .git is a FILE, not a directory, so this tests existence rather
// than directory-ness.
export function resolveRepoRoot(startDir: string | undefined): string | undefined {
  if (!startDir) return undefined;

  let dir = path.resolve(startDir);
  for (;;) {
    if (fs.existsSync(path.join(dir, ".git"))) return dir;
    const parent = path.dirname(dir);
    // path.dirname("/") === "/": we are at the filesystem root.
    if (parent === dir) break;
    dir = parent;
  }

  // No git root above it — fall back to the folder we were given rather than
  // returning nothing, so a non-git checkout behaves as it did before.
  return path.resolve(startDir);
}
