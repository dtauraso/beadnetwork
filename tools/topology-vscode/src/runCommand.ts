import * as vscode from "vscode";
import * as cp from "child_process";
import * as fs from "fs";
import * as path from "path";
import type { HostToWebviewMsg } from "./messages";
import { killOrphanedSims } from "./goBuild";
import { frameRecord } from "./schema/input-layout";
import { isProbeTraceEnabled } from "./probe-files";
import { decodeBufferLog, decodeStreamFrameEvents } from "./buffer-log";
import { decodeNodeStreamFrame, decodeEdgeStreamFrame, decodeInteriorStreamFrame } from "./webview/three/buffer-decode";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM } from "./schema/frame-tags";
import {
  VIEW_FD,
  EDGE_BASE_FD,
  MAX_EDGE_STREAMS,
  MAX_NODE_STREAMS,
  DRIVE_SLOTS_PER_NODE,
  nodeIdForRow,
} from "./runner/stream-fds";
import { readCounts } from "./runner/counts";
import { appendGoError } from "./runner/go-errors";
import { probePathsFor } from "./runner/probe-paths";
import { splitJsonlLines, splitFrames } from "./runner/framing";
import { ensureBinaryBuilt } from "./runner/ensure-binary";
import { freshStreamState, type StreamParseState } from "./runner/parse-state";

// The jobs this file used to do inline now live beside it under ./runner/, one concern per
// module: the fd-allocation contract and the ROW ID = NODE ID - 1 arithmetic
// (runner/stream-fds.ts), the stored topology counts (runner/counts.ts), the Go-error probe
// line (runner/go-errors.ts), the probe-log rotation and path layout (runner/probe-paths.ts),
// the two pure framing steps (runner/framing.ts), and the per-spawn parse state
// (runner/parse-state.ts). What stays here is the runner itself: spawn + env, the per-fd
// demux, the last-frame replay cache (owned by the INSTANCE — it is that process's bytes,
// never module state), lifecycle, and writeStdin.
//
// Re-exported so importers (extension.ts, extension/handle-message.ts, and the tests that
// import these from "../src/runCommand") keep working unchanged.
export { nodeIdForRow, rowForNodeId } from "./runner/stream-fds";
export { readCounts } from "./runner/counts";
export { splitJsonlLines, splitFrames, MAX_FRAME_BYTES } from "./runner/framing";

// Go stdout relay: errors (stderr, non-zero exit, spawn failure) are written to
// .probe/go-errors.jsonl. Trace events are no longer emitted on stdout at all (see
// handleStdout below) — the .probe trace logs are now the DECODE of each per-owner
// stream's own trailing EVENTS section (decodeBufferLog/decodeStreamFrameEvents, in
// handleViewFd/handleNodeFd/handleEdgeFd/handleInteriorFd).

export class BuildAndRunRunner {
  private proc: cp.ChildProcess | undefined;
  // Explicit cancel flag — distinguishing cancellation by signal name races
  // against natural exits, since a process that happened to die from SIGTERM
  // on its own would be misreported as "cancelled".
  private cancelled = false;
  // looping: when true, respawn automatically on natural exit. Set by run(). A
  // deliberate teardown (cancel(), from dispose) does not clear this — it sets
  // `cancelled`, and the close handler's cancelled branch suppresses the respawn.
  private looping = false;
  private channel: vscode.OutputChannel | undefined;
  // Per-process partial-frame parse state (stdout line + each dedicated stream's binary
  // frame). Rebuilt at every spawn by run() (freshStreamState), so its lifetime tracks the
  // Go process, not this long-lived runner — see freshStreamState for why that reset is at
  // the spawn.
  private stream: StreamParseState = freshStreamState(0, 0);
  private probeFile: string | undefined;
  private probeNodeFile: string | undefined;
  private probeEdgeFile: string | undefined;
  private probeInteriorFile: string | undefined;
  private goErrorsFile: string | undefined;
  // Read once per run() from wirefold.probe.trace. Gates only which DECODED LINES are
  // appended to the four Go trace files at handleViewFd/handleEdgeFd/handleNodeFd/
  // handleInteriorFd — breadcrumb rows (kind==="breadcrumb") are appended regardless (see
  // decodeStreamFrameEvents/decodeBufferLog's breadcrumbsOnly filtering in buffer-log.ts),
  // since CLAUDE.md designates them as the always-on Go debug channel.
  private probeTrace = false;
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
  private edgeCount = 0;

  // Last frame PER NODE ROW from the dedicated per-node NODE streams, and separately from
  // the dedicated per-node INTERIOR streams — the per-node analogues of lastEdgeFrames, one
  // map per stream kind since a node's geometry/ports/label and its interior beads are
  // written by two DIFFERENT goroutines (memory/feedback_no_single_writer_bridge.md) onto
  // two different fds. Same COPY-before-cache reasoning as lastViewFrame. Cleared on every
  // spawn.
  private lastNodeFrames: Map<number, ArrayBuffer> = new Map();
  private lastInteriorFrames: Map<number, ArrayBuffer> = new Map();
  // Current spawn's node-fd count. Recomputed at every run() call from the topology spec.
  private nodeCount = 0;

  // deadStreams marks per-fd streams that splitFrames reported a bad (over-MAX_FRAME_BYTES)
  // frame length on. Keyed "view" / "edge:<row>" / "node:<row>" / "interior:<row>". Once a
  // stream is dead its handleXFd short-circuits and never calls splitFrames again for it —
  // mirroring Go's stdin_reader.go, which logs and stops that reader goroutine rather than
  // trying to resynchronize on a corrupt length. Reset on every spawn (see freshStreamState's
  // call site in run()) since it's this-process's-bytes state, exactly like the *Bufs fields.
  private deadStreams: Set<string> = new Set();

  // spawnGen counts spawns. It is bumped in run(), BEFORE cp.spawn, so every frame this
  // runner relays carries the generation of the process that produced it — decided before
  // that process exists rather than inferred from what has arrived. The webview files rows
  // per generation, which is what makes a scene switch (a respawn) unable to mix the old
  // process's rows into the new one's tables.
  private spawnGen = 0;

  private topologyPath: string | undefined;

  // Set by restart(), consumed by the close handler's cancelled branch. cancel() alone
  // just stops the sim — it never respawns (that is the whole point of the cancelled/
  // looping split above). restart() needs a respawn to follow ITS OWN cancel, but the
  // close event is async (SIGTERM takes a beat to land), so the "please run() again once
  // this process is actually gone" intent has to survive from restart()'s call to the
  // close handler that fires later. A boolean is enough: restart() only ever needs to run
  // against this.topologyPath (never a different one), so there's nothing else to carry.
  private restartPending = false;

  constructor(
    private readonly onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void,
  ) {}

  // No spawn reveals the output panel. This was once a `reveal` option, true for the one
  // spawn a user started by opening the editor and false for every automatic one (the
  // hot-restart after a .go rebuild, the looping respawn, the webview's "ready" re-arm),
  // because an automatic spawn yanking the panel over whatever was showing made a crash
  // loop unreadable. The user-started case turned out to be no better: opening the editor
  // is exactly when someone wants to SEE the editor, and the panel came back at whatever
  // height VS Code had remembered — maximized, if it was maximized when the window was
  // last reloaded — covering the tab that had just opened. With its last true case gone the
  // option would only be a knob nothing sets, so it is gone too. A build FAILURE still
  // reveals: that is the one event with nothing else on screen to announce it.
  /** The generation of the process currently running (or last run). The webview-remount
   *  replay in handle-message.ts stamps the cached frames with THIS, because those frames
   *  are that process's own — replaying them under a fresh generation would file them in a
   *  table nothing reads. */
  currentGen(): number {
    return this.spawnGen;
  }

  run(topologyPath?: string) {
    if (this.proc) {
      // Already spawned: return silently, posting nothing. A webview that remounts
      // (reopened file, hot reload) re-learns liveness via the "ready" handler, which
      // replays the ext host's cached snapshot instead (see handle-message.ts's
      // wasRunning branch / BuildAndRunRunner.getLastSnapshot).
      return;
    }
    if (topologyPath) this.topologyPath = topologyPath;
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) return;
    // Channel setup happens here (not folded into the post-build block below) because the
    // build-failure branch needs a visible place to report the error to the user, same as
    // before this refactor.
    if (!this.channel) this.channel = vscode.window.createOutputChannel("topology run");
    this.channel.clear();
    const repoRoot = folder.uri.fsPath;
    const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
    const topArgs = this.topologyPath ? ["-topology", this.topologyPath] : [];
    // probePathsFor computes the probe-directory paths as PLAIN LOCALS (not yet written to
    // `this.*`) so the build-failure branch below can log to goErrorsFile without arming
    // any of the runner's own fields — see the ProbePaths/probePathsFor doc comment.
    const probePaths = probePathsFor(folder);
    // Build the binary once (and only rebuild when a .go source changed) instead of
    // relinking a throwaway binary via `go run .` on every start/restart.
    const built = ensureBinaryBuilt(repoRoot, binPath);
    if (!built.ok) {
      this.channel.appendLine(`\n[build error: ${built.error}]`);
      // Reveal even on an automatic spawn: a broken build is the one thing the user has to
      // see, and it is precisely the case where nothing else on screen will say so.
      this.channel.show(true);
      appendGoError(probePaths.goErrorsFile, built.error);
      this.looping = false;
      return;
    }
    // Reap orphaned sims left by prior/crashed editor sessions before spawning a
    // new one. exceptPid spares the proc this runner legitimately manages (the
    // stop/respawn logic still owns that). Single-panel assumption documented in
    // killOrphanedSims: this kills ALL matching sims except the active one.
    // this.proc is guaranteed undefined here (run() returns early at the top if a
    // proc exists), so there is no active sim to spare — exceptPid is undefined.
    // Passing it explicitly keeps the contract honest if that guard ever changes.
    // this.proc is undefined here (guarded by the early return above), so activePid
    // is always undefined — the cast overrides TypeScript's control-flow narrowing.
    const activePid: number | undefined = (this.proc as cp.ChildProcess | undefined)?.pid;
    const { killed } = killOrphanedSims(binPath, activePid);

    // Build (and orphan-reap) succeeded: from here on, arm every receiver field this run()
    // call touches in ONE uninterrupted run immediately before cp.spawn, so no early return
    // inserted above this point can ever leave probeFile/.../stream/lastSnapshot half-set —
    // see the ProbePaths doc comment for why probePaths was computed as locals above.
    this.probeFile = probePaths.probeFile;
    this.probeNodeFile = probePaths.probeNodeFile;
    this.probeEdgeFile = probePaths.probeEdgeFile;
    this.probeInteriorFile = probePaths.probeInteriorFile;
    this.goErrorsFile = probePaths.goErrorsFile;
    this.probeTrace = isProbeTraceEnabled();
    if (killed > 0) {
      this.channel.appendLine(`[cleanup] killed ${killed} orphaned sim process(es)`);
    }
    this.channel.appendLine("$ " + binPath + " " + topArgs.join(" "));
    this.cancelled = false;
    this.looping = true;
    // A new process: everything it emits belongs to a new generation.
    this.spawnGen++;
    // Size the dedicated per-edge/per-node fd ranges from the stored counts BEFORE spawning
    // (the ext host must know the range up front — see readCounts' doc comment). A missing
    // or malformed counts.json throws; that is a real configuration error, so it is reported
    // the same way a build failure is (channel + goErrorsFile + abort this run without
    // spawning), not silently treated as 0.
    let counts: { nodes: number; edges: number };
    try {
      counts = readCounts(this.topologyPath ?? path.join(repoRoot, "topology"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      this.channel.appendLine(`\n[counts.json error: ${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
      this.looping = false;
      return;
    }
    // Clamped to MAX_EDGE_STREAMS; a count above the bound omits the dedicated per-edge
    // streams entirely — see MAX_EDGE_STREAMS's doc comment.
    const edgeCountRaw = counts.edges;
    this.edgeCount = edgeCountRaw > MAX_EDGE_STREAMS ? 0 : edgeCountRaw;
    if (edgeCountRaw > MAX_EDGE_STREAMS) {
      // Capacity limit reached by legitimate input (a large topology), not a code bug — a
      // panic/throw here would be wrong. But silently zeroing edgeCount disables ALL
      // dedicated per-edge streams with no signal, which is the quietest possible failure
      // for the loudest consequence (this is the same path that used to strand the
      // `pending` leak, fixed in 93d2e9b6). Report it loudly through the same error
      // channel as every other operational problem in this file.
      const msg = `edge count ${edgeCountRaw} exceeds MAX_EDGE_STREAMS (${MAX_EDGE_STREAMS}); disabling ALL dedicated per-edge streams for this run`;
      this.channel.appendLine(`\n[${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
    }
    // Size the dedicated per-node NODE + INTERIOR fd ranges the same way, right after the
    // edge range (nodeBase = EDGE_BASE_FD + edgeCount, interiorBase = nodeBase + nodeCount —
    // see NODE_BASE_FD's doc comment). Clamped to MAX_NODE_STREAMS; 0 omits the dedicated
    // per-node NODE/INTERIOR/Port streams entirely.
    const nodeCountRaw = counts.nodes;
    this.nodeCount = nodeCountRaw > MAX_NODE_STREAMS ? 0 : nodeCountRaw;
    if (nodeCountRaw > MAX_NODE_STREAMS) {
      // Same reasoning as the edge-count case above.
      const msg = `node count ${nodeCountRaw} exceeds MAX_NODE_STREAMS (${MAX_NODE_STREAMS}); disabling ALL dedicated per-node NODE/INTERIOR streams for this run`;
      this.channel.appendLine(`\n[${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
    }
    const nodeBaseFd = EDGE_BASE_FD + this.edgeCount;
    const interiorBaseFd = nodeBaseFd + this.nodeCount;
    // driveBaseFd sits right after the interior range: nodeCount * DRIVE_SLOTS_PER_NODE
    // dedicated fds, one PER (node row, drive slot) — see DRIVE_SLOTS_PER_NODE's doc
    // comment and Buffer/stream_fds.go's StreamKindDrive. Required in lockstep with
    // "node"/"interior" (see the streamFDsEnvParts push below and main.go's matching
    // three-way check) — Go falls back to a loud stderr message and unwired streams
    // rather than a startup panic if this ever drifts from what Go expects (never a
    // crash-loop; see the panic-avoidance note on that fallback in main.go).
    const driveBaseFd = interiorBaseFd + this.nodeCount;
    // Fresh parse state for this spawn: a prior process's leftover partial frame must
    // never prefix this one's stream (see freshStreamState). This is the single reset
    // point every restart path funnels through, including the looping respawn.
    this.stream = freshStreamState(this.edgeCount, this.nodeCount);
    this.deadStreams.clear();
    // Also drop the cached keyframes: they belong to the PRIOR process. Without this,
    // a webview remounting in the window between "ready" and the new process's first
    // frame would be replayed the previous process's frames via getLastViewFrame()/
    // getLastEdgeFrames()/etc. The freshly spawned Go emits its full state again, so
    // continuity is preserved by that emit — not by re-serving one process's bytes as
    // another's.
    this.lastViewFrame = undefined;
    this.lastEdgeFrames.clear();
    this.lastNodeFrames.clear();
    this.lastInteriorFrames.clear();
    // detached: true makes the child the leader of a new process group; the
    // prebuilt binary is the sole group member, so kill(-pid) reaches it
    // directly. Without this, SIGTERM could leave it orphaned on macOS.
    // stdio index 3 is a RESERVED, UNUSED pipe slot: Go no longer writes anything to fd 3
    // (the central accumulator that used to write it, plus its fallback frames, was
    // deleted entirely; memory/feedback_no_single_writer_bridge.md's final step). The
    // slot stays allocated purely so the remaining fd numbering (VIEW_FD=4, edge/node/
    // interior ranges after it) matches this file's existing constants unchanged. stdio
    // index VIEW_FD (4) = the dedicated VIEW-stream pipe (WIREFOLD_STREAM_FDS=
    // "view:<VIEW_FD>"). stdio indices EDGE_BASE_FD..EDGE_BASE_FD+edgeCount-1 are one
    // dedicated pipe PER EDGE (see EDGE_BASE_FD's doc comment); the next nodeCount indices
    // are one dedicated pipe PER NODE (the "node" stream — geometry + ports + label); the
    // FOLLOWING nodeCount indices are one dedicated pipe PER NODE again (the "interior"
    // stream — that node's own interior beads, a SEPARATE goroutine's fd — see
    // NODE_BASE_FD's doc comment). The FOLLOWING nodeCount*DRIVE_SLOTS_PER_NODE indices
    // are one dedicated pipe PER (NODE, DRIVE SLOT) — the "drive" stream, one per
    // gatecommon.DriveHeld goroutine a node kind may spawn (docs/interior-stream-
    // framing.md's fix; see driveBaseFd's doc comment). Any of these ranges is omitted
    // (and its kind left out of WIREFOLD_STREAM_FDS) when its count is 0 (e.g. a topology
    // with no edges) — Go simply never streams that kind. "pipe" opens a readable pipe at
    // each index; the existing stdin(0)/stdout(1)/stderr(2) are unchanged.
    const stdio: Array<"pipe"> = ["pipe", "pipe", "pipe", "pipe", "pipe"];
    for (let i = 0; i < this.edgeCount; i++) stdio.push("pipe");
    for (let i = 0; i < this.nodeCount; i++) stdio.push("pipe");
    for (let i = 0; i < this.nodeCount; i++) stdio.push("pipe");
    for (let i = 0; i < this.nodeCount * DRIVE_SLOTS_PER_NODE; i++) stdio.push("pipe");
    const streamFDsEnvParts = [`view:${VIEW_FD}`];
    if (this.edgeCount > 0) streamFDsEnvParts.push(`edge:${EDGE_BASE_FD}`);
    // Go's stream_fds.go / main.go only wires the per-node node+interior+drive streams
    // when "node", "interior", AND "drive" env entries ALL resolve — always emit all
    // three together (main.go's three-way check treats a partial set the same as none).
    if (this.nodeCount > 0) {
      streamFDsEnvParts.push(`node:${nodeBaseFd}`, `interior:${interiorBaseFd}`, `drive:${driveBaseFd}`);
    }
    const streamFDsEnv = streamFDsEnvParts.join(",");
    this.proc = cp.spawn(binPath, [...topArgs], {
      cwd: repoRoot,
      detached: true,
      stdio,
      env: {
        ...process.env,
        WIREFOLD_BUF_OUT_FD: "3",
        WIREFOLD_STREAM_FDS: streamFDsEnv,
        // Gates whether nodes/wire's stepAll emits the per-tick, per-in-flight-bead
        // KindEdgeBead trace event at all -- read ONCE by Go at startup (see
        // nodes/wire/paced_wire.go's edgeBeadTraceEnabled), same shape as
        // WIREFOLD_STREAM_FDS above. Derived from the SAME isProbeTraceEnabled() that
        // gates whether the decoded .probe trace lines are WRITTEN on the TS side (see
        // this.probeTrace above) -- one source of truth for the wirefold.probe.trace
        // setting. Off by default: with tracing off, Go now stops emitting the event at
        // the source instead of TS decoding and discarding it every tick.
        WIREFOLD_EDGE_BEAD_TRACE: this.probeTrace ? "1" : "0",
      },
    });
    // Flush any framed binary records buffered before this spawn (writeStdin queued them).
    if (this.pendingStdin.length > 0) {
      for (const rec of this.pendingStdin) this.proc.stdin?.write(rec);
      this.pendingStdin = [];
    }
    this.proc.stdout?.on("data", (d: Buffer) => this.handleStdout(d.toString()));
    // stdio index 3 is a reserved, unused pipe (see the stdio comment above) — nothing
    // reads it; Go writes nothing to it.
    // VIEW_FD: the dedicated view-stream pipe. Cast needed because Node's ChildProcess
    // types only narrow stdio[0..2]; higher indices are typed as Readable|null via the
    // array form.
    const viewFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[VIEW_FD];
    if (viewFd) {
      viewFd.on("data", (d: Buffer) => this.handleViewFd(d));
    }
    // Per-edge dedicated pipes: EDGE_BASE_FD..EDGE_BASE_FD+edgeCount-1, one per edge row.
    for (let row = 0; row < this.edgeCount; row++) {
      const fdIdx = EDGE_BASE_FD + row;
      const edgeFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[fdIdx];
      if (edgeFd) {
        edgeFd.on("data", (d: Buffer) => this.handleEdgeFd(row, d));
      }
    }
    // Per-node dedicated pipes: nodeBaseFd..nodeBaseFd+nodeCount-1 (NODE stream, geometry+
    // ports+label) and interiorBaseFd..interiorBaseFd+nodeCount-1 (INTERIOR stream, that
    // node's own interior beads — a separate goroutine's fd, see NODE_BASE_FD's doc comment).
    for (let row = 0; row < this.nodeCount; row++) {
      const nodeFdIdx = nodeBaseFd + row;
      const nodeFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[nodeFdIdx];
      if (nodeFd) {
        nodeFd.on("data", (d: Buffer) => this.handleNodeFd(row, d));
      }
      const interiorFdIdx = interiorBaseFd + row;
      const interiorFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[interiorFdIdx];
      if (interiorFd) {
        interiorFd.on("data", (d: Buffer) => this.handleInteriorFd(row, d));
      }
      // Per-drive dedicated pipes: driveBaseFd + row*DRIVE_SLOTS_PER_NODE + slot, one PER
      // (node row, drive slot) — see driveBaseFd's doc comment. Each is its OWN pipe
      // (handleDriveFd keeps its carry buffer and dead-stream key separate per slot; see
      // driveBufs' doc comment) but decodes/relays through the SAME per-node interior
      // state as handleInteriorFd, since a drive-slot frame IS an interior-shaped frame
      // for this node row (Buffer.StreamKindDrive's doc comment).
      for (let slot = 0; slot < DRIVE_SLOTS_PER_NODE; slot++) {
        const driveFdIdx = driveBaseFd + row * DRIVE_SLOTS_PER_NODE + slot;
        const driveFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[driveFdIdx];
        if (driveFd) {
          driveFd.on("data", (d: Buffer) => this.handleDriveFd(row, slot, d));
        }
      }
    }
    this.proc.stderr?.on("data", (d: Buffer) => {
      const msg = d.toString();
      this.channel!.append(msg);
      appendGoError(this.goErrorsFile, msg);
    });
    this.proc.on("close", (code) => {
      const cancelled = this.cancelled;
      const looping = this.looping;
      this.proc = undefined;
      this.cancelled = false;
      if (cancelled) {
        this.channel!.appendLine("\n[cancelled]");
        if (this.restartPending) {
          // The respawn restart() asked for — funnels through the SAME run() every other
          // spawn path uses (freshStreamState reset, cache clear, orphan reap), rather than
          // a second spawn path that would have to duplicate all of it. run() with no
          // argument reuses this.topologyPath, which restart() never touched — same
          // topology the live run was started with.
          this.restartPending = false;
          this.run();
        }
      } else if (looping) {
        // Natural exit while looping — respawn immediately.
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
    this.proc.on("error", (err) => {
      this.proc = undefined;
      this.cancelled = false;
      this.channel!.appendLine(`\n[spawn error: ${err.message}]`);
      appendGoError(this.goErrorsFile, err.message);
    });
  }

  private handleStdout(chunk: string) {
    const { lines, rest } = splitJsonlLines(this.stream.stdoutBuf, chunk);
    this.stream.stdoutBuf = rest;
    for (const line of lines) {
      // Trace events (and, since task/breadcrumbs-binary-buffer, DEBUG BREADCRUMBs too)
      // are NO LONGER emitted on stdout: Go's JSON-trace emitter was removed and the
      // .probe log is now the DECODE of each per-owner stream's own trailing EVENTS
      // section (see handleViewFd/handleNodeFd/handleEdgeFd/handleInteriorFd below,
      // and buffer-log.ts's "breadcrumb" case). The ext host therefore no longer parses
      // any structured line from stdout; any stdout line here is just process output.
      this.channel!.appendLine(line);
    }
  }

  // handleViewFd parses the dedicated VIEW-stream pipe (VIEW_FD): frames are
  // [len:u32][payload] with NO tag byte (the fd position already identifies the stream —
  // see Buffer/stream_fds.go / frame-tags.ts). Relayed to the webview under the
  // "buffer-snapshot" message shape, tagged BUF_BLOCK_TAG_VIEW (a synthetic ext-host-side
  // tag, never a wire byte).
  private handleViewFd(chunk: Buffer) {
    if (this.deadStreams.has("view")) return;
    const { frames, rest, error } = splitFrames(this.stream.viewBuf, chunk);
    this.stream.viewBuf = rest;
    if (error) {
      this.deadStreams.add("view");
      const msg = `handleViewFd: ${error}`;
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
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
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_VIEW, gen: this.spawnGen });
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
  private handleEdgeFd(row: number, chunk: Buffer) {
    const key = `edge:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.edgeBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.edgeBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      const msg = `handleEdgeFd(row=${row}): ${error}`;
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
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
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_EDGE_STREAM, row, gen: this.spawnGen });
      }
    }
  }

  // handleNodeFd parses ONE dedicated per-node NODE stream pipe (fd = nodeBaseFd + row):
  // frames are [len:u32][payload] with NO tag byte (the fd position already identifies
  // WHICH node — see Buffer/stream_fds.go / Buffer/node_stream_frame.go). splitFrames is
  // reused as-is, same as handleEdgeFd. Each decoded frame is relayed to the webview under
  // the SAME "buffer-snapshot" shape, tagged BUF_BLOCK_TAG_NODE_STREAM (synthetic, never a
  // wire byte) PLUS `row` so the webview routes it to the right per-node cell.
  private handleNodeFd(row: number, chunk: Buffer) {
    const key = `node:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.nodeBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.nodeBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      const msg = `handleNodeFd(node=${nodeIdForRow(row)}): ${error}`;
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
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
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_NODE_STREAM, row, gen: this.spawnGen });
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
  private handleInteriorFd(row: number, chunk: Buffer) {
    const key = `interior:${row}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.interiorBufs[row] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    this.stream.interiorBufs[row] = rest;
    if (error) {
      this.deadStreams.add(key);
      const msg = `handleInteriorFd(node=${nodeIdForRow(row)}): ${error}`;
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
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
  private handleDriveFd(row: number, slot: number, chunk: Buffer) {
    const key = `drive:${row}:${slot}`;
    if (this.deadStreams.has(key)) return;
    const carry = this.stream.driveBufs[row]?.[slot] ?? Buffer.alloc(0);
    const { frames, rest, error } = splitFrames(carry, chunk);
    if (!this.stream.driveBufs[row]) this.stream.driveBufs[row] = [];
    this.stream.driveBufs[row][slot] = rest;
    if (error) {
      this.deadStreams.add(key);
      const msg = `handleDriveFd(node=${nodeIdForRow(row)}, slot=${slot}): ${error}`;
      this.channel?.appendLine(`\n[${msg}]`);
      appendGoError(this.goErrorsFile, msg);
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
        this.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_INTERIOR_STREAM, row, gen: this.spawnGen });
      }
    }
  }

  cancel() {
    // Drop any stdin lines buffered while proc was null — they belong to the
    // stopped session and must NOT replay onto the next spawned Go process (which
    // re-reads the graph from disk); stale replay would double-apply edits.
    this.pendingStdin = [];
    if (!this.proc || this.proc.pid === undefined) return;
    this.cancelled = true;
    try {
      // Negative pid → kill the whole process group (the leader created by
      // detached: true plus any descendants like the compiled binary).
      process.kill(-this.proc.pid, "SIGTERM");
    } catch {
      // Process already exited or no permission — the close handler will
      // clean up either way.
      this.proc.kill("SIGTERM");
    }
  }

  isRunning(): boolean {
    return this.proc !== undefined;
  }

  /**
   * restart() re-spawns Go against the SAME topologyPath the live run was started with,
   * for the hot-restart-on-.go-change feature (extension.ts's goWatcher). Returns false
   * and does nothing if no sim is live — a caller must not spawn one as a side effect of
   * editing a file (requirement 1 of that feature); starting a fresh sim is run()'s job,
   * not restart()'s.
   *
   * Deliberately funnels through cancel() + the close handler's run() call (see
   * restartPending) rather than introducing a second stop/spawn path: cancel() is already
   * the one place that tears a live proc down cleanly (process-group SIGTERM, pendingStdin
   * drop), and run() is already "the single reset point every restart path funnels
   * through" (see its freshStreamState comment) — this reuses both instead of duplicating
   * either.
   */
  restart(): boolean {
    if (!this.proc) return false;
    this.restartPending = true;
    this.cancel();
    return true;
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

  /** Framed binary records written before Go's stdin exists, flushed on spawn (see writeStdin/run). */
  private pendingStdin: Uint8Array[] = [];

  /**
   * Write a BINARY editor→Go record to Go's stdin, FRAMED as [len:u32-LE][record]
   * (symmetric with the per-goroutine content-buffer streams). Accepts either a bare record ArrayBuffer
   * (framed here) or an already-framed Uint8Array. If the process is not yet spawned,
   * BUFFER the framed bytes and flush once stdin exists (in run()) — early writes must not
   * be dropped (that lost the load-time guide-vis push, which races the spawn).
   *
   * Returns void: the TS→Go send is FIRE-AND-FORGET — no await, no request/response
   * (guard: check-no-await-on-bridge.sh).
   */
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
