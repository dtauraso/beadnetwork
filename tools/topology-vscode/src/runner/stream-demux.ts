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

/** Everything ONE spawned Go process's bytes need in order to be demultiplexed: which
 *  probe files this run writes, whether trace lines are written at all, how many
 *  edge/node rows exist, which generation the frames belong to, and where a decoded frame
 *  / a stdout line / an error goes. */
export interface StreamDemuxConfig {
  paths: ProbePaths | undefined;
  probeTrace: boolean;
  edgeCount: number;
  nodeCount: number;
  gen: number;
  onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void;
  /** A plain stdout line from Go — relayed to the output channel verbatim. */
  onLine: (line: string) => void;
  /** An operational problem (a bad frame length on some stream) — reported to the output
   *  channel AND the go-errors probe log by the runner, which owns both. */
  onError: (msg: string) => void;
}

/**
 * StreamDemux is ONE spawned Go process's read side: the per-fd frame reassembly, the
 * per-owner probe-log decode, the last-frame replay cache, and the relay to the webview.
 * The runner constructs one per spawn and delegates every "data" listener to it.
 *
 * Every field here is INSTANCE state, deliberately: it is that process's bytes. None of it
 * is module-level, because a second runner (or a second spawn) must not see the first
 * one's partial frames or cached keyframes — that is the exact bug freshStreamState exists
 * for, and hoisting any of it to module scope would reintroduce it and make the ext host a
 * holder of shared mutable state besides.
 */
export class StreamDemux {
  // Per-process partial-frame parse state (stdout line + each dedicated stream's binary
  // frame). Minted with this demux, so its lifetime tracks the Go process, not the
  // long-lived runner — see freshStreamState for why that reset is at the spawn.
  private stream: StreamParseState;

  private probeFile: string | undefined;
  private probeNodeFile: string | undefined;
  private probeEdgeFile: string | undefined;
  private probeInteriorFile: string | undefined;
  // Read once per run() from wirefold.probe.trace. Gates only which DECODED LINES are
  // appended to the four Go trace files at handleViewFd/handleEdgeFd/handleNodeFd/
  // handleInteriorFd — breadcrumb rows (kind==="breadcrumb") are appended regardless (see
  // decodeStreamFrameEvents/decodeBufferLog's breadcrumbsOnly filtering in buffer-log.ts),
  // since CLAUDE.md designates them as the always-on Go debug channel.
  private readonly probeTrace: boolean;

  // Last VIEW-stream frame (camera+overlay+scene), kept so a REMOUNTED webview (which holds
  // no state) can be handed a full frame instantly on "ready" without round-tripping to Go —
  // see getLastViewFrame(). This is a keyframe cache of Go's own bytes, not authored state.
  // MUST be a COPY, not the ArrayBuffer instance handed to onSnapshot/postMessage: VS Code's
  // webview.postMessage TRANSFERS (does not clone) ArrayBuffers to the webview process on
  // engines >=1.57 (see the @types/vscode postMessage doc comment — "will be more
  // efficiently transferred to the webview"), which DETACHES the source buffer (byteLength
  // -> 0) once posted. Caching the same reference would silently hand a later "ready" an
  // empty buffer. See runCommand.test.ts for the byteLength assertion.
  private lastViewFrame: ArrayBuffer | undefined;

  // Last frame PER EDGE ROW from the dedicated per-edge streams (see StreamParseState.
  // edgeBufs) — the per-edge analogue of lastViewFrame, keyed by edge row instead of a
  // singleton. Same COPY-before-cache reasoning as lastViewFrame (postMessage transfers/
  // detaches ArrayBuffers). Cleared on every spawn.
  private lastEdgeFrames: Map<number, ArrayBuffer> = new Map();
  // Current spawn's edge-fd count. Recomputed at every run() call from the topology spec.
  readonly edgeCount: number;

  // Last frame PER NODE ROW from the dedicated per-node NODE streams, and separately from
  // the dedicated per-node INTERIOR streams — the per-node analogues of lastEdgeFrames, one
  // map per stream kind since a node's geometry/ports/label and its interior beads are
  // written by two DIFFERENT goroutines (memory/feedback_no_single_writer_bridge.md) onto
  // two different fds. Same COPY-before-cache reasoning as lastViewFrame. Cleared on every
  // spawn.
  private lastNodeFrames: Map<number, ArrayBuffer> = new Map();
  private lastInteriorFrames: Map<number, ArrayBuffer> = new Map();
  // Current spawn's node-fd count. Recomputed at every run() call from the topology spec.
  readonly nodeCount: number;

  // deadStreams marks per-fd streams that splitFrames reported a bad (over-MAX_FRAME_BYTES)
  // frame length on. Keyed "view" / "edge:<row>" / "node:<row>" / "interior:<row>". Once a
  // stream is dead its handleXFd short-circuits and never calls splitFrames again for it —
  // mirroring Go's stdin_reader.go, which logs and stops that reader goroutine rather than
  // trying to resynchronize on a corrupt length. A fresh demux per spawn is what resets it,
  // exactly like the *Bufs fields.
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
      // Trace events (and, since task/breadcrumbs-binary-buffer, DEBUG BREADCRUMBs too)
      // are NO LONGER emitted on stdout: Go's JSON-trace emitter was removed and the
      // .probe log is now the DECODE of each per-owner stream's own trailing EVENTS
      // section (see handleViewFd/handleNodeFd/handleEdgeFd/handleInteriorFd below,
      // and buffer-log.ts's "breadcrumb" case). The ext host therefore no longer parses
      // any structured line from stdout; any stdout line here is just process output.
      this.onLine(line);
    }
  }

  // handleViewFd parses the dedicated VIEW-stream pipe (VIEW_FD): frames are
  // [len:u32][payload] with NO tag byte (the fd position already identifies the stream —
  // see Buffer/stream_fds.go / frame-tags.ts). Relayed to the webview under the
  // "buffer-snapshot" message shape, tagged BUF_BLOCK_TAG_VIEW (a synthetic ext-host-side
  // tag, never a wire byte).
  handleViewFd(chunk: Buffer) {
    if (this.deadStreams.has("view")) return;
    const { frames, rest, error } = splitFrames(this.stream.viewBuf, chunk);
    this.stream.viewBuf = rest;
    if (error) {
      this.deadStreams.add("view");
      this.onError(`handleViewFd: ${error}`);
    }
    for (const ab of frames) {
      // Decode this frame's OWN trailing EVENTS section (camera/overlay/scene events —
      // every other trace kind is decentralized to its own owner fd; memory/
      // feedback_no_single_writer_bridge.md). Written to its OWN .probe file (go.jsonl) —
      // N separate logs, never merged on write.
      if (this.probeFile) {
        const lines = decodeBufferLog(ab, !this.probeTrace);
        if (lines.length > 0) {
          try {
            fs.appendFileSync(this.probeFile, lines, "utf8");
          } catch { /* swallow */ }
        }
      }
      // Cache a COPY before handing `ab` off — see the lastViewFrame field comment for why
      // the reference itself cannot be cached (postMessage may transfer/detach it).
      this.lastViewFrame = ab.slice(0);
      if (this.onSnapshot) {
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_VIEW, gen: this.gen });
      }
    }
  }

  // handleEdgeFd parses ONE dedicated per-edge stream pipe (fd = EDGE_BASE_FD + row):
  // frames are [len:u32][payload] with NO tag byte (the fd position already identifies
  // WHICH edge — see Buffer/stream_fds.go / Buffer/edge_stream_frame.go). splitFrames is
  // reused as-is, same as handleViewFd. Each decoded frame is relayed to the webview under
  // the SAME "buffer-snapshot" shape as the other tags, tagged BUF_BLOCK_TAG_EDGE_STREAM
  // (synthetic, never a wire byte) PLUS `row` so the webview routes it to the right
  // per-edge cell (there are many edge streams, unlike view's singleton).
  handleEdgeFd(row: number, chunk: Buffer) {
    const key = `edge:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.edgeBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.edgeBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      this.onError(`handleEdgeFd(row=${row}): ${error}`);
    }
    for (const ab of frames) {
      // Decode this edge's OWN trailing EVENTS section (Geometry/Position/Arrive — this
      // goroutine's own row-resolved events; memory/feedback_no_single_writer_bridge.md).
      // Written to its OWN .probe file (go-edge.jsonl) — N separate logs, never merged.
      if (this.probeEdgeFile) {
        const decoded = decodeEdgeStreamFrame(row, ab);
        if (decoded && decoded.eventCount > 0) {
          const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
          if (lines.length > 0) {
            try {
              fs.appendFileSync(this.probeEdgeFile, lines, "utf8");
            } catch { /* swallow */ }
          }
        }
      }
      // Cache under this edge row (same copy-before-hand-off reasoning as lastViewFrame).
      this.lastEdgeFrames.set(row, ab.slice(0));
      if (this.onSnapshot) {
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_EDGE_STREAM, row, gen: this.gen });
      }
    }
  }

  // handleNodeFd parses ONE dedicated per-node NODE stream pipe (fd = nodeBaseFd + row):
  // frames are [len:u32][payload] with NO tag byte (the fd position already identifies
  // WHICH node — see Buffer/stream_fds.go / Buffer/node_stream_frame.go). splitFrames is
  // reused as-is, same as handleEdgeFd. Each decoded frame is relayed to the webview under
  // the SAME "buffer-snapshot" shape, tagged BUF_BLOCK_TAG_NODE_STREAM (synthetic, never a
  // wire byte) PLUS `row` so the webview routes it to the right per-node cell.
  handleNodeFd(row: number, chunk: Buffer) {
    const key = `node:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.nodeBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.nodeBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      this.onError(`handleNodeFd(node=${nodeIdForRow(row)}): ${error}`);
    }
    for (const ab of frames) {
      // Decode this node's OWN trailing EVENTS section (NodeGeometry — this nodeMover
      // goroutine's own row-resolved event; memory/feedback_no_single_writer_bridge.md).
      // Written to its OWN .probe file (go-node.jsonl) — N separate logs, never merged.
      if (this.probeNodeFile) {
        const decoded = decodeNodeStreamFrame(row, ab);
        if (decoded && decoded.eventCount > 0) {
          const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
          if (lines.length > 0) {
            try {
              fs.appendFileSync(this.probeNodeFile, lines, "utf8");
            } catch { /* swallow */ }
          }
        }
      }
      // Cache under this node row (same copy-before-hand-off reasoning as lastViewFrame).
      this.lastNodeFrames.set(row, ab.slice(0));
      if (this.onSnapshot) {
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_NODE_STREAM, row, gen: this.gen });
      }
    }
  }

  // handleInteriorFd parses ONE dedicated per-node INTERIOR stream pipe (fd =
  // interiorBaseFd + row) — that node's OWN Update goroutine (a SEPARATE goroutine from
  // its nodeMover), same framing/relay shape as handleNodeFd, tagged
  // BUF_BLOCK_TAG_INTERIOR_STREAM. Delegates decode/cache/relay to
  // processInteriorLikeFrames, shared with handleDriveFd — see that method's doc comment
  // for why the CARRY BUFFER and dead-stream key stay separate per fd even though the
  // decoded output lands in the same place.
  handleInteriorFd(row: number, chunk: Buffer) {
    const key = `interior:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.interiorBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.interiorBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      this.onError(`handleInteriorFd(node=${nodeIdForRow(row)}): ${error}`);
    }
    this.processInteriorLikeFrames(row, frames, true); // this node own interior stream: the sole author of slot state
  }

  // handleDriveFd parses ONE dedicated per-(node row, drive slot) DRIVE stream pipe (fd =
  // driveBaseFd + row*DRIVE_SLOTS_PER_NODE + slot) — one gatecommon.DriveHeld goroutine's
  // OWN fd (docs/interior-stream-framing.md's fix: this goroutine used to share the
  // node's INTERIOR fd with its Update-loop goroutine, desyncing the frame reader — see
  // that doc for the mechanism). Its OWN carry buffer and dead-stream key
  // (`drive:{row}:{slot}`) are kept separate per (row, slot) — see driveBufs' doc
  // comment: merging two physically distinct pipes' partial-frame state would reintroduce
  // the exact desync this fd split exists to remove, just on the read side instead of the
  // write side. The DECODED frames, though, are a drive-slot frame is an interior-shaped
  // frame for this node row (Buffer.StreamKindDrive's doc comment) — so once split and
  // reassembled, they're handed to the SAME processInteriorLikeFrames as
  // handleInteriorFd: same probe file, same lastInteriorFrames cache (last writer wins,
  // matching this node's own single most-recent bead-state snapshot regardless of which
  // goroutine produced it), same BUF_BLOCK_TAG_INTERIOR_STREAM tag to the webview.
  handleDriveFd(row: number, slot: number, chunk: Buffer) {
    const key = `drive:${row}:${slot}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.driveBufs[row]?.[slot] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    if (!this.stream.driveBufs[row]) this.stream.driveBufs[row] = [];
    this.stream.driveBufs[row][slot] = rest;
    if (error) {
      this.deadStreams.add(key);
      this.onError(`handleDriveFd(node=${nodeIdForRow(row)}, slot=${slot}): ${error}`);
    }
    this.processInteriorLikeFrames(row, frames, false); // drive slot: events only
  }

  // processInteriorLikeFrames is the shared decode/probe-log tail of handleInteriorFd and
  // handleDriveFd, once each has independently reassembled its OWN fd's frames (see
  // handleDriveFd's doc comment for why reassembly itself is NOT shared).
  //
  // assertsSlots says whether this frame's four interior SLOTS mean anything. Only the
  // node's own interior stream authors slot state (emitHeldBead / emitNodeBeads /
  // emitInputBeads, all on that node's own Update goroutine). A DRIVE-slot stream is
  // interior-SHAPED but authors none of it: its WriteEvents reuses ITS OWN stream's
  // lastPresent, which nothing ever sets, so every drive frame carries an all-absent
  // snapshot. Relaying that as though it were a statement about the node erased the held
  // bead a fraction of a second after it appeared — the node still held it, and nothing
  // took its place, because the drive frame had nothing to put there.
  //
  // Drive frames are still decoded and probe-logged here (their EVENTS are the point of
  // them); they simply no longer reach the webview's interior state. One writer for what is
  // inside a node.
  private processInteriorLikeFrames(row: number, frames: ArrayBuffer[], assertsSlots: boolean) {
    for (const ab of frames) {
      // Decode this node's OWN trailing EVENTS section (NodeBead — this node's own
      // Update-loop goroutine's row-resolved events; memory/feedback_no_single_writer_bridge.md).
      // Written to its OWN .probe file (go-interior.jsonl) — N separate logs, never merged.
      if (this.probeInteriorFile) {
        const decoded = decodeInteriorStreamFrame(row, ab);
        if (decoded && decoded.eventCount > 0) {
          const lines = decodeStreamFrameEvents(decoded.eventCount, decoded.eventView, decoded.eventTextView, undefined, undefined, !this.probeTrace);
          if (lines.length > 0) {
            try {
              fs.appendFileSync(this.probeInteriorFile, lines, "utf8");
            } catch { /* swallow */ }
          }
        }
      }
      if (!assertsSlots) continue; // a drive frame: logged above, never interior state
      this.lastInteriorFrames.set(row, ab.slice(0));
      if (this.onSnapshot) {
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_INTERIOR_STREAM, row, gen: this.gen });
      }
    }
  }

  /** The most recent VIEW-stream frame (camera+overlay+scene), or undefined if none has
   *  arrived yet. Used by the "ready" handler to hand a remounted webview the cached frame
   *  instantly (see the lastViewFrame field comment).
   *
   *  The returned buffer is a FRESH COPY, because the caller posts what it gets and
   *  webview.postMessage TRANSFERS ArrayBuffers — handing out the cached reference
   *  would detach our own cache on the first serve. That breaks the exact case this
   *  cache exists for: while PAUSED no new frame ever arrives to repopulate it, so a
   *  second remount would be served a zero-length buffer. The copy is one per remount. */
  getLastViewFrame(): ArrayBuffer | undefined {
    return this.lastViewFrame?.slice(0);
  }

  /** The most recent frame for EVERY cached edge row (see lastEdgeFrames), or an empty
   *  array if none has arrived yet. Used by the "ready" handler to hand a remounted webview
   *  every edge's last frame instantly, the per-edge analogue of getLastViewFrame(). Same
   *  fresh-copy-per-remount reasoning. */
  getLastEdgeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastEdgeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  /** The most recent frame for EVERY cached node row from the dedicated NODE stream, the
   *  per-node analogue of getLastEdgeFrames. */
  getLastNodeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastNodeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  /** The most recent frame for EVERY cached node row from the dedicated INTERIOR stream. */
  getLastInteriorFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastInteriorFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }
}
