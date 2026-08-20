import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";
import { readCounts } from "../counts";
import { appendGoError } from "../probe/go-errors";
import { probePathsFor, type ProbePaths } from "../probe/probe-paths";
import { computeSpawnLayout, type SpawnLayout } from "../spawn-layout";
import { StreamDemux } from "../stream-demux";
import type { HostToWebviewMsg } from "../../../schema/messages";

export interface PreparedRun {
  goErrorsFile: string;
  spawnGen: number;
  layout: SpawnLayout;
  demux: StreamDemux;
}

export function prepareRunLayout(args: {
  channel: vscode.OutputChannel;
  probePaths: ProbePaths;
  killed: number;
  binPath: string;
  topArgs: string[];
  repoRoot: string;
  topologyPath: string | undefined;
  spawnGenBefore: number;
  probeTrace: boolean;
  newDemux: (paths: ProbePaths, probeTrace: boolean, edgeCount: number, nodeCount: number, gen: number) => StreamDemux;
}): PreparedRun | undefined {
  const { channel, probePaths, killed, binPath, topArgs, repoRoot, topologyPath, spawnGenBefore, probeTrace, newDemux } = args;

  if (killed > 0) {
    channel.appendLine(`[cleanup] killed ${killed} orphaned sim process(es)`);
  }
  channel.appendLine("$ " + binPath + " " + topArgs.join(" "));

  const spawnGen = spawnGenBefore + 1;

  let counts: { nodes: number; edges: number };
  try {
    counts = readCounts(topologyPath ?? path.join(repoRoot, "topology"));
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    channel.appendLine(`\n[counts.json error: ${msg}]`);
    appendGoError(probePaths.goErrorsFile, msg);
    return undefined;
  }

  const layout = computeSpawnLayout(counts);
  for (const msg of layout.warnings) {
    channel.appendLine(`\n[${msg}]`);
    appendGoError(probePaths.goErrorsFile, msg);
  }

  const demux = newDemux(probePaths, probeTrace, layout.edgeCount, layout.nodeCount, spawnGen);

  return { goErrorsFile: probePaths.goErrorsFile, spawnGen, layout, demux };
}

export function probePathsForFolder(folder: vscode.WorkspaceFolder): ProbePaths {
  return probePathsFor(folder);
}

export function makeDemuxFactory(hooks: {
  onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void;
  appendLine: (line: string) => void;
  reportError: (msg: string) => void;
}): (paths: ProbePaths | undefined, probeTrace: boolean, edgeCount: number, nodeCount: number, gen: number) => StreamDemux {
  return (paths, probeTrace, edgeCount, nodeCount, gen) => {
    const demux = new StreamDemux({
      paths,
      probeTrace,
      edgeCount,
      nodeCount,
      gen,
      onSnapshot: (msg) => hooks.onSnapshot?.(msg),
      onLine: (line) => hooks.appendLine(line),
      onError: (msg) => hooks.reportError(msg),
    });

    demux.seedOwnerCounts(nodeCount, edgeCount);
    return demux;
  };
}

export interface ExitHandlerHooks {
  isCancelled(): boolean;
  clearCancelled(): void;
  isLooping(): boolean;
  isRestartPending(): boolean;
  clearRestartPending(): void;
  appendLine(line: string): void;
  reportError(msg: string): void;
  restart(): void;
}

export function wireExitHandlers(proc: cp.ChildProcess, clearProc: () => void, hooks: ExitHandlerHooks): void {
  proc.on("close", (code) => {
    const cancelled = hooks.isCancelled();
    const looping = hooks.isLooping();
    clearProc();
    hooks.clearCancelled();
    if (cancelled) {
      hooks.appendLine("\n[cancelled]");
      if (hooks.isRestartPending()) {
        hooks.clearRestartPending();
        hooks.restart();
      }
    } else if (looping) {
      hooks.appendLine(code === 0 ? "\n[ok — restarting]" : `\n[exit ${code} — restarting]`);
      hooks.restart();
    } else if (code === 0) {
      hooks.appendLine("\n[ok]");
    } else {
      const message = `exit code ${code}`;
      hooks.appendLine(`\n[${message}]`);
      hooks.reportError(message);
    }
  });
  proc.on("error", (err) => {
    clearProc();
    hooks.clearCancelled();
    hooks.appendLine(`\n[spawn error: ${err.message}]`);
    hooks.reportError(err.message);
  });
}
