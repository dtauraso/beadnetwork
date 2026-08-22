import * as cp from "child_process";
import * as fs from "fs";
import * as path from "path";

const GO_WALK_EXCLUDE = new Set([
  "node_modules", ".git", "out", ".probe", ".wirefold-cache", "handoff-archive",
]);

export function maxGoMtime(dir: string): number {
  let max = 0;
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return max;
  }
  for (const ent of entries) {
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) {
      if (GO_WALK_EXCLUDE.has(ent.name)) continue;
      const sub = maxGoMtime(full);
      if (sub > max) max = sub;
    } else if (ent.isFile() && ent.name.endsWith(".go")) {
      try {
        const m = fs.statSync(full).mtimeMs;
        if (m > max) max = m;
      } catch { /* eslint-disable-line no-empty */ }
    }
  }
  return max;
}

export type BuildResult =
  | { ok: true; busy?: boolean }
  | { ok: false; error: string };

let building = false;

export function killOrphanedSims(binPath: string, exceptPid?: number): { killed: number } {
  if (process.platform !== "darwin" && process.platform !== "linux") {
    return { killed: 0 };
  }
  let out: string;
  try {
    const res = cp.spawnSync("ps", ["-axo", "pid=,command="], { encoding: "utf8" });
    if (res.status !== 0 || typeof res.stdout !== "string") return { killed: 0 };
    out = res.stdout;
  } catch {
    return { killed: 0 };
  }
  const self = process.pid;
  let killed = 0;
  for (const rawLine of out.split("\n")) {
    const line = rawLine.trim();
    if (!line) continue;
    const sp = line.indexOf(" ");
    if (sp < 0) continue;
    const pid = Number(line.slice(0, sp));
    if (!Number.isInteger(pid) || pid <= 0) continue;
    const command = line.slice(sp + 1);

    if (pid === self || (exceptPid !== undefined && pid === exceptPid)) continue;

    if (!command.includes("wirefold")) continue;
    if (!command.includes(binPath) && !command.includes("-topology")) continue;
    try {
      process.kill(pid, "SIGKILL");
      killed++;
    } catch { /* eslint-disable-line no-empty */ }
  }
  return { killed };
}

export function buildBinary(repoRoot: string, binPath: string): BuildResult {
  if (building) return { ok: true, busy: true };
  building = true;
  try {
    try {
      fs.mkdirSync(path.dirname(binPath), { recursive: true });
    } catch (e) {
      return { ok: false, error: (e as Error).message };
    }
    const res = cp.spawnSync("go", ["build", "-o", binPath, "./Start"], { cwd: repoRoot, encoding: "utf8" });
    if (res.error) return { ok: false, error: res.error.message };
    if (res.status !== 0) return { ok: false, error: res.stderr || `go build exited ${res.status}` };
    return { ok: true };
  } finally {
    building = false;
  }
}
