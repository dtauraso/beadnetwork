import type { HostToWebviewMsg } from "../../Input/messages";
import { nodeIdForRow } from "./stream-fds";
import { splitJsonlLines } from "./framing";
import { freshStreamState, type StreamParseState } from "./parse-state";
import type { ProbePaths } from "./probe/probe-paths";
import { LastFrameStore } from "./last-frame-store";
import { ColumnStore } from "./column-store";
import { BUF_BLOCK_TAG_COLUMN } from "../../Buffer/frame-tags";
import {
  COL_STREAM_SCENE_NODE_COUNT, COL_STREAM_SCENE_EDGE_COUNT,
} from "../../Scene/columns-gen";
import {
  dispatchViewFrames,
  dispatchEdgeFrames,
  dispatchNodeFrames,
  dispatchBeadFrames,
  makeFrameDispatchContext,
  type FrameDispatchContext,
} from "./probe/frame-dispatch";
import { handleInteriorFdImpl } from "./stream-demux-interior";

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
  private probeDir: string | undefined;

  readonly edgeCount: number;

  readonly nodeCount: number;

  readonly columns = new ColumnStore();

  private readonly onSnapshot: ((msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void) | undefined;
  private readonly gen: number;

  private readonly frameCtx: FrameDispatchContext;
  private readonly onLine: (line: string) => void;

  constructor(cfg: StreamDemuxConfig) {
    super();
    this.stream = freshStreamState(cfg.edgeCount, cfg.nodeCount);
    this.probeFile = cfg.paths?.probeFile;
    this.probeDir = cfg.paths?.probeDir;
    this.edgeCount = cfg.edgeCount;
    this.nodeCount = cfg.nodeCount;
    this.frameCtx = makeFrameDispatchContext(cfg.probeTrace, cfg.gen, cfg.onSnapshot, cfg.onError);
    this.onLine = cfg.onLine;
    this.onSnapshot = cfg.onSnapshot;
    this.gen = cfg.gen;
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
      this.probeDir,
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
      this.probeDir,
      (row, ab) => { this.lastNodeFrames.set(row, ab); },
    );
  }

  handleBeadFd(row: number, chunk: Buffer) {
    dispatchBeadFrames(
      this.frameCtx,
      row,
      this.stream.beadBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.beadBufs[row] = rest; },
      this.probeDir,
      (row, ab) => { this.lastBeadFrames.set(row, ab); },
    );
  }

  seedOwnerCounts(nodes: number, edges: number): void {
    if (!this.onSnapshot) return;
    for (const [col, value] of [[COL_STREAM_SCENE_NODE_COUNT, nodes], [COL_STREAM_SCENE_EDGE_COUNT, edges]] as const) {
      const buf = Buffer.alloc(4);
      buf.writeInt32LE(value, 0);
      const ab = buf.buffer.slice(buf.byteOffset, buf.byteOffset + 4);
      this.columns.seed(col, buf);
      this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_COLUMN, row: col, gen: this.gen });
    }
  }

  handleColFd(col: number, chunk: Buffer) {
    if (!this.columns.handle(col, chunk)) return;
    const value = this.columns.get(col);
    if (!value || !this.onSnapshot) return;

    const ab = value.buffer.slice(value.byteOffset, value.byteOffset + value.byteLength) as ArrayBuffer;
    this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_COLUMN, row: col, gen: this.gen });
  }

  handleInteriorFd(row: number, chunk: Buffer) {
    handleInteriorFdImpl(this.frameCtx, this.stream, this.probeDir, row, chunk,
      (row, ab) => { this.lastInteriorFrames.set(row, ab); });
  }
}
