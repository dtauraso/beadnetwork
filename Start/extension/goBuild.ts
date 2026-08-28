import * as cp from "child_process";
import * as fs from "fs";
import * as path from "path";

const GO_WALK_EXCLUDE = new Set([
  "node_modules", ".git", "out", ".probe", ".beadnetwork-cache", "handoff-archive",
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

    if (!command.includes("beadnetwork")) continue;
    if (!command.includes(binPath) && !command.includes("-topology")) continue;
    try {
      process.kill(pid, "SIGKILL");
      killed++;
    } catch { /* eslint-disable-line no-empty */ }
  }
  return { killed };
}

export function buildBinary(repoRoot: string, binPath: string): Promise<BuildResult> {
  if (building) return Promise.resolve({ ok: true, busy: true });
  building = true;

  try {
    fs.mkdirSync(path.dirname(binPath), { recursive: true });
  } catch (e) {
    building = false;
    return Promise.resolve({ ok: false, error: (e as Error).message });
  }

  return new Promise<BuildResult>((resolve) => {
    const proc = cp.spawn("go", ["build", "-o", binPath, "./Start"], { cwd: repoRoot });
    let stderr = "";
    proc.stderr?.on("data", (chunk: Buffer) => { stderr += chunk.toString(); });
    proc.on("error", (err) => {
      building = false;
      resolve({ ok: false, error: err.message });
    });
    proc.on("close", (code) => {
      building = false;
      if (code === 0) resolve({ ok: true });
      else resolve({ ok: false, error: stderr || `go build exited ${String(code)}` });
    });
  });
}
