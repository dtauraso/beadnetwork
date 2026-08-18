import * as vscode from "vscode";
import * as fs from "fs";
import * as path from "path";
import { PROBE_DIR, PROBE_FILES, PROBE_TRACE_FILES, PROBE_OWNER_DIRS } from "../../probe-files";
import { resolveRepoRoot } from "../../repo-root";

export interface ProbePaths {

  probeFile: string;

  probeDir: string;
  goErrorsFile: string;

}

function rotateProbeLog(p: string): void {
  try {
    if (!fs.existsSync(p)) return;
    fs.rmSync(`${p}.prev`, { force: true });
    fs.renameSync(p, `${p}.prev`);
  } catch { /* eslint-disable-line no-empty */ }
}

function rotateProbeOwnerDir(dir: string): void {
  try {
    fs.rmSync(`${dir}.prev`, { recursive: true, force: true });
    if (fs.existsSync(dir)) fs.renameSync(dir, `${dir}.prev`);
  } catch { /* eslint-disable-line no-empty */ }
  fs.mkdirSync(dir, { recursive: true });
}

export function probePathsFor(folder: vscode.WorkspaceFolder): ProbePaths {
  const root = resolveRepoRoot(folder.uri.fsPath) ?? folder.uri.fsPath;
  const probeDir = path.join(root, PROBE_DIR);
  fs.mkdirSync(probeDir, { recursive: true });

  rotateProbeLog(path.join(probeDir, PROBE_FILES.goErrors));
  rotateProbeLog(path.join(probeDir, PROBE_FILES.tsErrors));
  rotateProbeLog(path.join(probeDir, PROBE_FILES.handlerErrorLast));
  for (const name of PROBE_TRACE_FILES) {
    rotateProbeLog(path.join(probeDir, name));
  }
  for (const owner of PROBE_OWNER_DIRS) {
    rotateProbeOwnerDir(path.join(probeDir, owner));
  }
  return {
    probeFile: path.join(probeDir, PROBE_FILES.go),
    probeDir,
    goErrorsFile: path.join(probeDir, PROBE_FILES.goErrors),
  };
}
