import * as vscode from "vscode";
import * as cp from "child_process";
import { killOrphanedSims } from "../../goBuild";
import { appendGoError } from "../probe/go-errors";
import { ensureBinaryBuilt } from "./ensure-binary";
import { isProbeTraceEnabled } from "../../probe-files";

export async function buildBinary(
  channel: vscode.OutputChannel,
  repoRoot: string,
  binPath: string,
  goErrorsFile: string | undefined,
): Promise<boolean> {
  const built = await ensureBinaryBuilt(repoRoot, binPath);
  if (!built.ok) {
    channel.appendLine(`\n[build error: ${built.error}]`);
    channel.show(true);
    appendGoError(goErrorsFile, built.error);
    return false;
  }
  return true;
}

export function reapOrphans(binPath: string, activePid: number | undefined): number {
  const { killed } = killOrphanedSims(binPath, activePid);
  return killed;
}

export function spawnProcess(
  binPath: string,
  topArgs: string[],
  repoRoot: string,
): cp.ChildProcess {
  return cp.spawn(binPath, [...topArgs], {
    cwd: repoRoot,
    detached: true,
    stdio: ["pipe", "pipe", "pipe"],
    env: {
      ...process.env,
      BEADNETWORK_PROBE_TRACE: isProbeTraceEnabled() ? "1" : "0",
    },
  });
}

export function attachStreamHandlers(
  proc: cp.ChildProcess,
  channel: vscode.OutputChannel,
  goErrorsFile: string | undefined,
): void {
  proc.stderr?.on("data", (d: Buffer) => {
    const msg = d.toString();
    channel.append(msg);
    appendGoError(goErrorsFile, msg);
  });
}
