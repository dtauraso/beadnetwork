import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";
import type { HostToWebviewMsg } from "./messages";
import { frameRecord } from "./schema/input-encode";
import { isProbeTraceEnabled } from "./probe-files";
import { appendGoError } from "./runner/go-errors";
import { probePathsForFolder, prepareRunLayout, wireExitHandlers, makeDemuxFactory } from "./runner/run-lifecycle";
import type { StreamDemux } from "./runner/stream-demux";
import { buildBinary, reapOrphans, spawnProcess, attachStreamHandlers } from "./runner/process-lifecycle";

export { nodeIdForRow, rowForNodeId } from "./runner/stream-fds";
export { readCounts } from "./runner/counts";
export { splitJsonlLines, splitFrames, MAX_FRAME_BYTES } from "./runner/framing";

export class BuildAndRunRunner {
  private proc: cp.ChildProcess | undefined;
  private cancelled = false;
  private looping = false;
  private channel: vscode.OutputChannel | undefined;
  private goErrorsFile: string | undefined;

  private readonly newDemux = makeDemuxFactory({
    onSnapshot: (msg) => this.onSnapshot?.(msg),
    appendLine: (line) => this.channel?.appendLine(line),
    reportError: (msg) => {
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
    },
  });

  private demux: StreamDemux = this.newDemux(undefined, false, 0, 0, 0);
  private spawnGen = 0;
  private topologyPath: string | undefined;
  private restartPending = false;
  private pendingStdin: Uint8Array[] = [];

  constructor(
    private readonly onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void,
  ) {}

  currentGen(): number {
    return this.spawnGen;
  }

  run(topologyPath?: string) {
    if (this.proc) return;
    if (topologyPath) this.topologyPath = topologyPath;
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) return;

    this.ensureOutputChannel();
    const repoRoot = folder.uri.fsPath;
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

  private ensureOutputChannel(): void {
    if (!this.channel) this.channel = vscode.window.createOutputChannel("topology run");
    this.channel.clear();
  }

  cancel() {
    this.pendingStdin = [];
    if (!this.proc || this.proc.pid === undefined) return;
    this.cancelled = true;
    try {
      process.kill(-this.proc.pid, "SIGTERM");
    } catch {
      this.proc.kill("SIGTERM");
    }
  }

  isRunning(): boolean {
    return this.proc !== undefined;
  }

  restart(): boolean {
    if (!this.proc) return false;
    this.restartPending = true;
    this.cancel();
    return true;
  }

  getLastViewFrame(): ArrayBuffer | undefined {
    return this.demux.getLastViewFrame();
  }

  getLastEdgeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastEdgeFrames();
  }

  getLastNodeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastNodeFrames();
  }

  getLastInteriorFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastInteriorFrames();
  }

  writeStdin(record: ArrayBuffer | Uint8Array): void {
    const framed = record instanceof Uint8Array ? record : frameRecord(record);
    if (!this.proc?.stdin) {
      this.pendingStdin.push(framed);
      return;
    }
    this.proc.stdin.write(framed);
  }

  dispose() {
    this.cancel();
    this.channel?.dispose();
  }
}
