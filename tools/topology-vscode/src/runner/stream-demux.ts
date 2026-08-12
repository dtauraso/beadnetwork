import type { HostToWebviewMsg } from "../messages";
import { nodeIdForRow } from "./stream-fds";
import { splitJsonlLines } from "./framing";
import { freshStreamState, type StreamParseState } from "./parse-state";
import type { ProbePaths } from "./probe/probe-paths";
import { LastFrameStore } from "./last-frame-store";
import {
  dispatchViewFrames,
  dispatchEdgeFrames,
  dispatchNodeFrames,
  makeFrameDispatchContext,
  type FrameDispatchContext,
} from "./probe/frame-dispatch";
import { handleInteriorFdImpl, handleDriveFdImpl } from "./stream-demux-interior";

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

export class StreamDemux extends LastFrameStore {

  private stream: StreamParseState;

  private probeFile: string | undefined;
  private probeNodeFile: string | undefined;
  private probeEdgeFile: string | undefined;
  private probeInteriorFile: string | undefined;

  readonly edgeCount: number;

  readonly nodeCount: number;

  private readonly frameCtx: FrameDispatchContext;
  private readonly onLine: (line: string) => void;

  constructor(cfg: StreamDemuxConfig) {
    super();
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
    handleInteriorFdImpl(this.frameCtx, this.stream, this.probeInteriorFile, row, chunk,
      (row, ab) => { this.lastInteriorFrames.set(row, ab); });
  }

  handleDriveFd(row: number, slot: number, chunk: Buffer) {
    handleDriveFdImpl(this.frameCtx, this.stream, this.probeInteriorFile, row, slot, chunk,
      (row, ab) => { this.lastInteriorFrames.set(row, ab); });
  }
}
