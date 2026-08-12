import * as vscode from "vscode";
import * as cp from "child_process";
import type { HostToWebviewMsg } from "../messages";
import { frameRecord } from "../schema/input-encode";
import type { StreamDemux } from "./stream-demux";
import { makeDemuxFactory } from "./run-lifecycle";
import { appendGoError } from "./go-errors";

export abstract class RunnerLifecycle {
  protected proc: cp.ChildProcess | undefined;
  protected cancelled = false;
  protected looping = false;
  protected channel: vscode.OutputChannel | undefined;
  protected goErrorsFile: string | undefined;

  protected readonly newDemux = makeDemuxFactory({
    onSnapshot: (msg) => this.onSnapshot?.(msg),
    appendLine: (line) => this.channel?.appendLine(line),
    reportError: (msg) => {
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
    },
  });

  protected demux: StreamDemux = this.newDemux(undefined, false, 0, 0, 0);
  protected spawnGen = 0;
  protected topologyPath: string | undefined;
  protected restartPending = false;
  protected pendingStdin: Uint8Array[] = [];

  constructor(
    protected readonly onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void,
  ) {}

  currentGen(): number {
    return this.spawnGen;
  }

  protected ensureOutputChannel(): void {
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
