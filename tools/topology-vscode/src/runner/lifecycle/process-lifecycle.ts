import * as vscode from "vscode";
import * as cp from "child_process";
import { killOrphanedSims } from "../../goBuild";
import { appendGoError } from "../probe/go-errors";
import { ensureBinaryBuilt } from "./ensure-binary";
import { attachStreamListeners } from "./attach-listeners";
import type { StreamDemux } from "../stream-demux";
import type { SpawnLayout } from "../spawn-layout";

export function buildBinary(
  channel: vscode.OutputChannel,
  repoRoot: string,
  binPath: string,
  goErrorsFile: string | undefined,
): boolean {
  const built = ensureBinaryBuilt(repoRoot, binPath);
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
  layout: SpawnLayout,
  probeTrace: boolean,
): cp.ChildProcess {
  const { stdio, streamFDsEnv } = layout;
  return cp.spawn(binPath, [...topArgs], {
    cwd: repoRoot,
    detached: true,
    stdio,
    env: {
      ...process.env,
      WIREFOLD_BUF_OUT_FD: "3",
      WIREFOLD_STREAM_FDS: streamFDsEnv,
      WIREFOLD_EDGE_BEAD_TRACE: probeTrace ? "1" : "0",
    },
  });
}

export function attachStreamHandlers(
  proc: cp.ChildProcess,
  demux: StreamDemux,
  layout: SpawnLayout,
  channel: vscode.OutputChannel,
  goErrorsFile: string | undefined,
): void {
  attachStreamListeners(proc, demux, layout);
  proc.stderr?.on("data", (d: Buffer) => {
    const msg = d.toString();
    channel.append(msg);
    appendGoError(goErrorsFile, msg);
  });
}
