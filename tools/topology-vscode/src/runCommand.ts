import * as vscode from "vscode";
import * as path from "path";
import { isProbeTraceEnabled } from "./probe-files";
import { appendGoError } from "./runner/probe/go-errors";
import { probePathsForFolder, prepareRunLayout, wireExitHandlers } from "./runner/lifecycle/run-lifecycle";
import { buildBinary, reapOrphans, spawnProcess, attachStreamHandlers } from "./runner/lifecycle/process-lifecycle";
import { RunnerLifecycle } from "./runner/lifecycle/runner-base";
import { resolveRepoRoot } from "./repo-root";

export { nodeIdForRow, rowForNodeId } from "./runner/stream-fds";
export { readCounts } from "./runner/counts";

export { splitJsonlLines, splitFrames, MAX_FRAME_BYTES } from "./runner/framing";

export class BuildAndRunRunner extends RunnerLifecycle {
  run(topologyPath?: string) {
    if (this.proc) return;
    if (topologyPath) this.topologyPath = topologyPath;
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) return;

    this.ensureOutputChannel();
    // The git root, NOT the workspace folder: `go build .` has to run where the
    // Go module is, whichever subdirectory the window happens to be open on.
    const repoRoot = resolveRepoRoot(folder.uri.fsPath);
    if (!repoRoot) return;
    const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
    const topArgs = this.topologyPath ? ["-topology", this.topologyPath] : [];

    const probePaths = probePathsForFolder(folder);
    if (!buildBinary(this.channel!, repoRoot, binPath, probePaths.goErrorsFile)) {
      this.looping = false;
      return;
    }
    const killed = reapOrphans(binPath, undefined);

    this.goErrorsFile = probePaths.goErrorsFile;
    this.cancelled = false;
    this.looping = true;
    const probeTrace = isProbeTraceEnabled();

    const prepared = prepareRunLayout({
      channel: this.channel!,
      probePaths,
      killed,
      binPath,
      topArgs,
      repoRoot,
      topologyPath: this.topologyPath,
      spawnGenBefore: this.spawnGen,
      probeTrace,
      newDemux: (paths, trace, edgeCount, nodeCount, gen) => this.newDemux(paths, trace, edgeCount, nodeCount, gen),
    });
    if (!prepared) {
      this.looping = false;
      return;
    }
    this.spawnGen = prepared.spawnGen;
    this.demux = prepared.demux;
    const { layout } = prepared;

    this.proc = spawnProcess(binPath, topArgs, repoRoot, layout, probeTrace);
    if (this.pendingStdin.length > 0) {
      for (const rec of this.pendingStdin) this.proc.stdin?.write(rec);
      this.pendingStdin = [];
    }
    attachStreamHandlers(this.proc, prepared.demux, layout, this.channel!, this.goErrorsFile);

    wireExitHandlers(this.proc, () => { this.proc = undefined; }, {
      isCancelled: () => this.cancelled,
      clearCancelled: () => { this.cancelled = false; },
      isLooping: () => this.looping,
      isRestartPending: () => this.restartPending,
      clearRestartPending: () => { this.restartPending = false; },
      appendLine: (line) => this.channel!.appendLine(line),
      reportError: (msg) => appendGoError(this.goErrorsFile, msg),
      restart: () => this.run(),
    });
  }
}
