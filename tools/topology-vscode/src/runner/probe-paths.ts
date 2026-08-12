import * as vscode from "vscode";
import * as fs from "fs";
import * as path from "path";
import { PROBE_DIR, PROBE_FILES, PROBE_TRACE_FILES } from "../probe-files";


export interface ProbePaths {






  probeFile: string;
  probeNodeFile: string;
  probeEdgeFile: string;
  probeInteriorFile: string;
  goErrorsFile: string;



}



function rotateProbeLog(p: string): void {
  try {
    if (!fs.existsSync(p)) return;
    fs.rmSync(`${p}.prev`, { force: true });
    fs.renameSync(p, `${p}.prev`);
  } catch {

  }
}

export function probePathsFor(folder: vscode.WorkspaceFolder): ProbePaths {
  const probeDir = path.join(folder.uri.fsPath, PROBE_DIR);
  fs.mkdirSync(probeDir, { recursive: true });



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
