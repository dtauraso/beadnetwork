import * as vscode from "vscode";
import * as fs from "fs";
import * as path from "path";
import { PROBE_DIR, PROBE_FILES, PROBE_TRACE_FILES } from "../probe-files";

/** The probe-path set derived once per run() from the workspace folder — see
 *  probePathsFor. */
export interface ProbePaths {
  // All paths are ALWAYS armed, regardless of wirefold.probe.trace: the four Go trace
  // files (probeFile/probeNodeFile/probeEdgeFile/probeInteriorFile) double as the DEBUG
  // BREADCRUMB channel's storage (.claude/rules/go-debugging.md — breadcrumb rows
  // must survive with tracing off), so they must exist and be writable either way. What
  // the setting gates is which DECODED LINES get appended at each write site — see
  // handleViewFd/handleEdgeFd/handleNodeFd/handleInteriorFd's probeTrace-gated filtering.
  probeFile: string;
  probeNodeFile: string;
  probeEdgeFile: string;
  probeInteriorFile: string;
  goErrorsFile: string;
  // NOTE: ts.jsonl / ts-errors.jsonl are deliberately absent. The runner never writes
  // them — extension/webview-log.ts is the sole writer and resolves its own paths from
  // the webview's workspace folder. Fields for them were armed here and read by nobody.
}

/** Computes (and ensures on disk) the probe-directory file paths for one run. Pure w.r.t.
 *  the runner — returns a plain object rather than writing `this.*` fields, so a caller can
 *  use goErrorsFile to report a build failure BEFORE deciding whether to arm the runner's
 *  own fields (see run()). */
/** Rotates one probe log: `<f>` -> `<f>.prev` (overwriting any older .prev), leaving no
 *  `<f>` so the run's first appendFileSync starts it fresh.
 *
 *  Why rotate rather than truncate: the probe writers are append-only and nothing ever
 *  reset them, so `.probe/` grew without bound across every run AND every reload-window
 *  (measured at 1.2 GB, 1.1 GB of it go-edge.jsonl). But plain truncate-on-start would
 *  destroy the log of the run you just did — which is the exact moment you want it, since
 *  the first move on an editor hang is to read go-errors.jsonl
 *  (memory/feedback_runner_errors_probe_first). One generation keeps that evidence alive
 *  across a single reload while bounding growth at two runs.
 *
 *  Best-effort: a rotation failure must never stop a run from starting, so errors are
 *  swallowed. The worst case is the old behavior (this run appends to the existing file).
 *  tools/probe-merge.sh reads the live files and is unaffected. */
function rotateProbeLog(p: string): void {
  try {
    if (!fs.existsSync(p)) return;
    fs.rmSync(`${p}.prev`, { force: true });
    fs.renameSync(p, `${p}.prev`);
  } catch {
    /* best effort — never block a run on log rotation */
  }
}

export function probePathsFor(folder: vscode.WorkspaceFolder): ProbePaths {
  const probeDir = path.join(folder.uri.fsPath, PROBE_DIR);
  fs.mkdirSync(probeDir, { recursive: true });
  // All paths rotate unconditionally: the four Go trace files always receive breadcrumb
  // rows regardless of wirefold.probe.trace (see ProbePaths), so they are live logs either
  // way and must rotate like the error logs.
  rotateProbeLog(path.join(probeDir, PROBE_FILES.goErrors));
  rotateProbeLog(path.join(probeDir, PROBE_FILES.tsErrors));
  rotateProbeLog(path.join(probeDir, PROBE_FILES.handlerErrorLast));
  for (const name of PROBE_TRACE_FILES) {
    rotateProbeLog(path.join(probeDir, name));
  }
  return {
    probeFile: path.join(probeDir, PROBE_FILES.go),
    probeNodeFile: path.join(probeDir, PROBE_FILES.goNode),
    probeEdgeFile: path.join(probeDir, PROBE_FILES.goEdge),
    probeInteriorFile: path.join(probeDir, PROBE_FILES.goInterior),
    goErrorsFile: path.join(probeDir, PROBE_FILES.goErrors),
  };
}
