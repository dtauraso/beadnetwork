import * as fs from "fs";
import type { HostToWebviewMsg } from "../messages";
import { decodeBufferLog, decodeStreamFrameEvents } from "../buffer-log";
import { decodeNodeStreamFrame } from "../webview/three/decode/buffer-decode-node";
import { decodeEdgeStreamFrame } from "../webview/three/decode/buffer-decode-edge";
import { decodeInteriorStreamFrame } from "../webview/three/decode/buffer-decode-interior";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM } from "../schema/frame-tags";
import { nodeIdForRow } from "./stream-fds";
import { splitJsonlLines, splitFrames } from "./framing";
import { freshStreamState, type StreamParseState } from "./parse-state";
import type { ProbePaths } from "./probe-paths";

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

  private readonly probeTrace: boolean;

  private lastViewFrame: ArrayBuffer | undefined;

  private lastEdgeFrames: Map<number, ArrayBuffer> = new Map();

  readonly edgeCount: number;

  private lastNodeFrames: Map<number, ArrayBuffer> = new Map();
  private lastInteriorFrames: Map<number, ArrayBuffer> = new Map();

  readonly nodeCount: number;

  private deadStreams: Set<string> = new Set();

  private readonly gen: number;
  private readonly onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void;
  private readonly onLine: (line: string) => void;
  private readonly onError: (msg: string) => void;

  constructor(cfg: StreamDemuxConfig) {
    this.stream = freshStreamState(cfg.edgeCount, cfg.nodeCount);
    this.probeFile = cfg.paths?.probeFile;
    this.probeNodeFile = cfg.paths?.probeNodeFile;
    this.probeEdgeFile = cfg.paths?.probeEdgeFile;
    this.probeInteriorFile = cfg.paths?.probeInteriorFile;
    this.probeTrace = cfg.probeTrace;
    this.edgeCount = cfg.edgeCount;
    this.nodeCount = cfg.nodeCount;
    this.gen = cfg.gen;
    this.onSnapshot = cfg.onSnapshot;
    this.onLine = cfg.onLine;
    this.onError = cfg.onError;
  }

  handleStdout(chunk: string) {
    const { lines, rest } = splitJsonlLines(this.stream.stdoutBuf, chunk);
    this.stream.stdoutBuf = rest;
    for (const line of lines) {

      this.onLine(line);
    }
  }

  private dispatchFrames(
    key: string,
    carry: Buffer,
    chunk: Buffer,
    storeRest: (rest: Buffer) => void,
    errorContext: string,
    onFrames: (frames: ArrayBuffer[]) => void,
  ) {
    if (this.deadStreams.has(key)) return;
    const { frames, rest, error } = splitFrames(carry, chunk);
    storeRest(rest);
    if (error) {
      this.deadStreams.add(key);
      this.onError(`${errorContext}: ${error}`);
    }
    onFrames(frames);
  }

  handleViewFd(chunk: Buffer) {
    this.dispatchFrames(
      "view",
      this.stream.viewBuf,
      chunk,
      (rest) => { this.stream.viewBuf = rest; },
      "handleViewFd",
      (frames) => {
        for (const ab of frames) {

          if (this.probeFile) {
            const lines = decodeBufferLog(ab, !this.probeTrace);
            if (lines.length > 0) {
              try {
                fs.appendFileSync(this.probeFile, lines, "utf8");
              } catch { /* eslint-disable-line no-empty */ }
            }
          }

          this.lastViewFrame = ab.slice(0);
          if (this.onSnapshot) {
            this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_VIEW, gen: this.gen });
          }
        }
      },
    );
  }

  handleEdgeFd(row: number, chunk: Buffer) {
    this.dispatchFrames(
      `edge:${row}`,
      this.stream.edgeBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.edgeBufs[row] = rest; },
      `handleEdgeFd(row=${row})`,
      (frames) => {
        for (const ab of frames) {

          if (this.probeEdgeFile) {
            const decoded = decodeEdgeStreamFrame(row, ab);
            if (decoded && decoded.eventCount > 0) {
              const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
              if (lines.length > 0) {
                try {
                  fs.appendFileSync(this.probeEdgeFile, lines, "utf8");
                } catch { /* eslint-disable-line no-empty */ }
              }
            }
          }

          this.lastEdgeFrames.set(row, ab.slice(0));
          if (this.onSnapshot) {
            this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_EDGE_STREAM, row, gen: this.gen });
          }
        }
      },
    );
  }

  handleNodeFd(row: number, chunk: Buffer) {
    this.dispatchFrames(
      `node:${row}`,
      this.stream.nodeBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.nodeBufs[row] = rest; },
      `handleNodeFd(node=${nodeIdForRow(row)})`,
      (frames) => {
        for (const ab of frames) {

          if (this.probeNodeFile) {
            const decoded = decodeNodeStreamFrame(row, ab);
            if (decoded && decoded.eventCount > 0) {
              const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
              if (lines.length > 0) {
                try {
                  fs.appendFileSync(this.probeNodeFile, lines, "utf8");
                } catch { /* eslint-disable-line no-empty */ }
              }
            }
          }

          this.lastNodeFrames.set(row, ab.slice(0));
          if (this.onSnapshot) {
            this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_NODE_STREAM, row, gen: this.gen });
          }
        }
      },
    );
  }

  handleInteriorFd(row: number, chunk: Buffer) {
    this.dispatchFrames(
      `interior:${row}`,
      this.stream.interiorBufs[row] ?? Buffer.alloc(0),
      chunk,
      (rest) => { this.stream.interiorBufs[row] = rest; },
      `handleInteriorFd(node=${nodeIdForRow(row)})`,

      (frames) => this.processInteriorLikeFrames(row, frames, true),
    );
  }

  handleDriveFd(row: number, slot: number, chunk: Buffer) {
    this.dispatchFrames(
      `drive:${row}:${slot}`,
      this.stream.driveBufs[row]?.[slot] ?? Buffer.alloc(0),
      chunk,
      (rest) => {
        if (!this.stream.driveBufs[row]) this.stream.driveBufs[row] = [];
        this.stream.driveBufs[row][slot] = rest;
      },
      `handleDriveFd(node=${nodeIdForRow(row)}, slot=${slot})`,

      (frames) => this.processInteriorLikeFrames(row, frames, false),
    );
  }

  private processInteriorLikeFrames(row: number, frames: ArrayBuffer[], assertsSlots: boolean) {
    for (const ab of frames) {

      if (this.probeInteriorFile) {
        const decoded = decodeInteriorStreamFrame(row, ab);
        if (decoded && decoded.eventCount > 0) {
          const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
          if (lines.length > 0) {
            try {
              fs.appendFileSync(this.probeInteriorFile, lines, "utf8");
            } catch { /* eslint-disable-line no-empty */ }
          }
        }
      }
      if (!assertsSlots) continue; 
      this.lastInteriorFrames.set(row, ab.slice(0));
      if (this.onSnapshot) {
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_INTERIOR_STREAM, row, gen: this.gen });
      }
    }
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
