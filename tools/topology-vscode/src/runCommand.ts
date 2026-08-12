import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";
import type { HostToWebviewMsg } from "./messages";
import { killOrphanedSims } from "./goBuild";
import { frameRecord } from "./schema/input-encode";
import { isProbeTraceEnabled } from "./probe-files";
import { readCounts } from "./runner/counts";
import { appendGoError } from "./runner/go-errors";
import { probePathsFor, type ProbePaths } from "./runner/probe-paths";
import { ensureBinaryBuilt } from "./runner/ensure-binary";
import { StreamDemux } from "./runner/stream-demux";
import { computeSpawnLayout, type SpawnLayout } from "./runner/spawn-layout";
import { attachStreamListeners } from "./runner/attach-listeners";

export { nodeIdForRow, rowForNodeId } from "./runner/stream-fds";
export { readCounts } from "./runner/counts";
export { splitJsonlLines, splitFrames, MAX_FRAME_BYTES } from "./runner/framing";

export class BuildAndRunRunner {
  private proc: cp.ChildProcess | undefined;

  private cancelled = false;

  private looping = false;
  private channel: vscode.OutputChannel | undefined;
  private goErrorsFile: string | undefined;

  private demux: StreamDemux = this.newDemux(undefined, false, 0, 0, 0);

  private spawnGen = 0;

  private topologyPath: string | undefined;

  private restartPending = false;

  constructor(
    private readonly onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void,
  ) {}

  currentGen(): number {
    return this.spawnGen;
  }

  private newDemux(paths: ProbePaths | undefined, probeTrace: boolean, edgeCount: number, nodeCount: number, gen: number): StreamDemux {
    return new StreamDemux({
      paths,
      probeTrace,
      edgeCount,
      nodeCount,
      gen,
      onSnapshot: (msg) => this.onSnapshot?.(msg),
      onLine: (line) => this.channel!.appendLine(line),
      onError: (msg) => {
        this.channel?.appendLine(`\n[${msg}]`);
        appendGoError(this.goErrorsFile, msg);
      },
    });
  }

  run(topologyPath?: string) {
    if (this.proc) {

      return;
    }
    if (topologyPath) this.topologyPath = topologyPath;
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) return;

    this.ensureOutputChannel();
    const repoRoot = folder.uri.fsPath;
    const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
    const topArgs = this.topologyPath ? ["-topology", this.topologyPath] : [];

    const probePaths = probePathsFor(folder);
    if (!this.buildBinary(repoRoot, binPath, probePaths)) return;
    const killed = this.reapOrphans(binPath);

    const prepared = this.armFieldsAndPrepareLayout(probePaths, killed, binPath, topArgs, repoRoot);
    if (!prepared) return;
    const { layout, demux } = prepared;

    this.spawnProcess(binPath, topArgs, repoRoot, layout);
    this.attachStreamHandlers(demux, layout);
    this.wireExitHandlers();
  }

  private ensureOutputChannel(): void {
    if (!this.channel) this.channel = vscode.window.createOutputChannel("topology run");
    this.channel.clear();
  }

  private buildBinary(repoRoot: string, binPath: string, probePaths: ProbePaths): boolean {
    const built = ensureBinaryBuilt(repoRoot, binPath);
    if (!built.ok) {
      this.channel!.appendLine(`\n[build error: ${built.error}]`);

      this.channel!.show(true);
      appendGoError(probePaths.goErrorsFile, built.error);
      this.looping = false;
      return false;
    }
    return true;
  }

  private reapOrphans(binPath: string): number {
    const activePid: number | undefined = this.proc?.pid;
    const { killed } = killOrphanedSims(binPath, activePid);
    return killed;
  }

  private armFieldsAndPrepareLayout(
    probePaths: ProbePaths,
    killed: number,
    binPath: string,
    topArgs: string[],
    repoRoot: string,
  ): { layout: SpawnLayout; demux: StreamDemux } | undefined {
    this.goErrorsFile = probePaths.goErrorsFile;
    const probeTrace = isProbeTraceEnabled();
    if (killed > 0) {
      this.channel!.appendLine(`[cleanup] killed ${killed} orphaned sim process(es)`);
    }
    this.channel!.appendLine("$ " + binPath + " " + topArgs.join(" "));
    this.cancelled = false;
    this.looping = true;

    this.spawnGen++;

    let counts: { nodes: number; edges: number };
    try {
      counts = readCounts(this.topologyPath ?? path.join(repoRoot, "topology"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      this.channel!.appendLine(`\n[counts.json error: ${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
      this.looping = false;
      return undefined;
    }

    const layout = computeSpawnLayout(counts);
    const { edgeCount, nodeCount } = layout;
    for (const msg of layout.warnings) {
      this.channel!.appendLine(`\n[${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
    }

    const demux = this.newDemux(probePaths, probeTrace, edgeCount, nodeCount, this.spawnGen);
    this.demux = demux;
    return { layout, demux };
  }

  private spawnProcess(binPath: string, topArgs: string[], repoRoot: string, layout: SpawnLayout): void {
    const { stdio, streamFDsEnv } = layout;

    const probeTrace = isProbeTraceEnabled();
    this.proc = cp.spawn(binPath, [...topArgs], {
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

    if (this.pendingStdin.length > 0) {
      for (const rec of this.pendingStdin) this.proc.stdin?.write(rec);
      this.pendingStdin = [];
    }
  }

  private attachStreamHandlers(demux: StreamDemux, layout: SpawnLayout): void {
    attachStreamListeners(this.proc!, demux, layout);
    this.proc!.stderr?.on("data", (d: Buffer) => {
      const msg = d.toString();
      this.channel!.append(msg);
      appendGoError(this.goErrorsFile, msg);
    });
  }

  private wireExitHandlers(): void {
    this.proc!.on("close", (code) => {
      const cancelled = this.cancelled;
      const looping = this.looping;
      this.proc = undefined;
      this.cancelled = false;
      if (cancelled) {
        this.channel!.appendLine("\n[cancelled]");
        if (this.restartPending) {

          this.restartPending = false;
          this.run();
        }
      } else if (looping) {

        this.channel!.appendLine(code === 0 ? "\n[ok — restarting]" : `\n[exit ${code} — restarting]`);
        this.run();
      } else if (code === 0) {
        this.channel!.appendLine("\n[ok]");
      } else {
        const message = `exit code ${code}`;
        this.channel!.appendLine(`\n[${message}]`);
        appendGoError(this.goErrorsFile, message);
      }
    });
    this.proc!.on("error", (err) => {
      this.proc = undefined;
      this.cancelled = false;
      this.channel!.appendLine(`\n[spawn error: ${err.message}]`);
      appendGoError(this.goErrorsFile, err.message);
    });
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

  private pendingStdin: Uint8Array[] = [];

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
