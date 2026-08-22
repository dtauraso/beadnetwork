import * as vscode from "vscode";
import * as cp from "child_process";
import { appendGoError } from "../probe/go-errors";
import { probePathsFor, type ProbePaths } from "../probe/probe-paths";

export interface PreparedRun {
  goErrorsFile: string;
  spawnGen: number;
}

export function prepareRunLayout(args: {
  channel: vscode.OutputChannel;
  probePaths: ProbePaths;
  killed: number;
  binPath: string;
  topArgs: string[];
  spawnGenBefore: number;
}): PreparedRun | undefined {
  const { channel, probePaths, killed, binPath, topArgs, spawnGenBefore } = args;

  if (killed > 0) {
    channel.appendLine(`[cleanup] killed ${killed} orphaned sim process(es)`);
  }
  channel.appendLine("$ " + binPath + " " + topArgs.join(" "));

  return { goErrorsFile: probePaths.goErrorsFile, spawnGen: spawnGenBefore + 1 };
}

export function probePathsForFolder(folder: vscode.WorkspaceFolder): ProbePaths {
  return probePathsFor(folder);
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
