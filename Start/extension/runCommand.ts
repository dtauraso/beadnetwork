import * as vscode from "vscode";
import * as path from "path";
import { appendGoError } from "./runner/probe/go-errors";
import { probePathsForFolder, prepareRunLayout, wireExitHandlers } from "./runner/lifecycle/run-lifecycle";
import { buildBinary, reapOrphans, spawnProcess, attachStreamHandlers } from "./runner/lifecycle/process-lifecycle";
import { RunnerLifecycle } from "./runner/lifecycle/runner-base";
import { resolveRepoRoot } from "./repo-root";

export class BuildAndRunRunner extends RunnerLifecycle {
  run(topologyPath?: string) {
    if (this.proc) return;
    if (topologyPath) this.topologyPath = topologyPath;
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) return;

    this.ensureOutputChannel();
    const repoRoot = resolveRepoRoot(folder.uri.fsPath);
    if (!repoRoot) return;
    const binPath = path.join(repoRoot, ".beadnetwork-cache", "beadnetwork");
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

    const prepared = prepareRunLayout({
      channel: this.channel!,
      probePaths,
      killed,
      binPath,
      topArgs,
      spawnGenBefore: this.spawnGen,
    });
    if (!prepared) {
      this.looping = false;
      return;
    }
    this.spawnGen = prepared.spawnGen;

    this.proc = spawnProcess(binPath, topArgs, repoRoot);
    if (this.pendingStdin.length > 0) {
      for (const rec of this.pendingStdin) this.proc.stdin?.write(rec);
      this.pendingStdin = [];
    }
    attachStreamHandlers(this.proc, this.channel!, this.goErrorsFile);

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
