import type { HostToWebviewMsg } from "../messages";
import { nodeIdForRow } from "./stream-fds";
import { splitJsonlLines } from "./framing";
import { freshStreamState, type StreamParseState } from "./parse-state";
import type { ProbePaths } from "./probe-paths";
import {
  dispatchViewFrames,
  dispatchEdgeFrames,
  dispatchNodeFrames,
  dispatchInteriorLikeFrames,
  makeFrameDispatchContext,
  type FrameDispatchContext,
} from "./frame-dispatch";

export interface StreamDemuxConfig {
  paths: ProbePaths | undefined;
  probeTrace: boolean;
  edgeCount: number;
  nodeCount: number;
  gen: number;
  onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void;

  onLine: (line: string) => void;

  onError: (msg: string) => void;
}

export class StreamDemux {

  private stream: StreamParseState;

  private probeFile: string | undefined;
  private probeNodeFile: string | undefined;
  private probeEdgeFile: string | undefined;
  private probeInteriorFile: string | undefined;

  private lastViewFrame: ArrayBuffer | undefined;

  private lastEdgeFrames: Map<number, ArrayBuffer> = new Map();

  readonly edgeCount: number;

  private lastNodeFrames: Map<number, ArrayBuffer> = new Map();
  private lastInteriorFrames: Map<number, ArrayBuffer> = new Map();

  readonly nodeCount: number;

  private readonly frameCtx: FrameDispatchContext;
  private readonly onLine: (line: string) => void;

  constructor(cfg: StreamDemuxConfig) {
    this.stream = freshStreamState(cfg.edgeCount, cfg.nodeCount);
    this.probeFile = cfg.paths?.probeFile;
    this.probeNodeFile = cfg.paths?.probeNodeFile;
    this.probeEdgeFile = cfg.paths?.probeEdgeFile;
    this.probeInteriorFile = cfg.paths?.probeInteriorFile;
    this.edgeCount = cfg.edgeCount;
    this.nodeCount = cfg.nodeCount;
    this.frameCtx = makeFrameDispatchContext(cfg.probeTrace, cfg.gen, cfg.onSnapshot, cfg.onError);
    this.onLine = cfg.onLine;
  }

  handleStdout(chunk: string) {
    const { lines, rest } = splitJsonlLines(this.stream.stdoutBuf, chunk);
    this.stream.stdoutBuf = rest;
    for (const line of lines) {

      this.onLine(line);
    }
  }

  handleViewFd(chunk: Buffer) {
    dispatchViewFrames(
      this.frameCtx,
      this.stream.viewBuf,
      chunk,
      (rest) => { this.stream.viewBuf = rest; },
      this.probeFile,
      (ab) => { this.lastViewFrame = ab; },
    );
  }

  handleEdgeFd(row: number, chunk: Buffer) {
    dispatchEdgeFrames(
      this.frameCtx,
      row,
      this.stream.edgeBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.edgeBufs[row] = rest; },
      this.probeEdgeFile,
      (row, ab) => { this.lastEdgeFrames.set(row, ab); },
    );
  }

  handleNodeFd(row: number, chunk: Buffer) {
    dispatchNodeFrames(
      this.frameCtx,
      row,
      `handleNodeFd(node=${nodeIdForRow(row)})`,
      this.stream.nodeBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.nodeBufs[row] = rest; },
      this.probeNodeFile,
      (row, ab) => { this.lastNodeFrames.set(row, ab); },
    );
  }

  handleInteriorFd(row: number, chunk: Buffer) {
    dispatchInteriorLikeFrames(
      this.frameCtx,
      `interior:${row}`,
      row,
      `handleInteriorFd(node=${nodeIdForRow(row)})`,
      this.stream.interiorBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.interiorBufs[row] = rest; },
      this.probeInteriorFile,
      true,
      (row, ab) => { this.lastInteriorFrames.set(row, ab); },
    );
  }

  handleDriveFd(row: number, slot: number, chunk: Buffer) {
    dispatchInteriorLikeFrames(
      this.frameCtx,
      `drive:${row}:${slot}`,
      row,
      `handleDriveFd(node=${nodeIdForRow(row)}, slot=${slot})`,
      this.stream.driveBufs[row]?.[slot] ?? Buffer.alloc(0),
      chunk,
      (rest) => {
        if (!this.stream.driveBufs[row]) this.stream.driveBufs[row] = [];
        this.stream.driveBufs[row][slot] = rest;
      },
      this.probeInteriorFile,
      false,
      (row, ab) => { this.lastInteriorFrames.set(row, ab); },
    );
  }

  getLastViewFrame(): ArrayBuffer | undefined {
    return this.lastViewFrame?.slice(0);
  }

  getLastEdgeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastEdgeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  getLastNodeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastNodeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  getLastInteriorFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastInteriorFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }
}
