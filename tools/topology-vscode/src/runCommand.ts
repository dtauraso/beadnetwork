import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";
import type { HostToWebviewMsg } from "./messages";
import { killOrphanedSims } from "./goBuild";
import { frameRecord } from "./schema/input-layout";
import { isProbeTraceEnabled } from "./probe-files";
import {
  VIEW_FD,
  EDGE_BASE_FD,
  MAX_EDGE_STREAMS,
  MAX_NODE_STREAMS,
  DRIVE_SLOTS_PER_NODE,
} from "./runner/stream-fds";
import { readCounts } from "./runner/counts";
import { appendGoError } from "./runner/go-errors";
import { probePathsFor, type ProbePaths } from "./runner/probe-paths";
import { ensureBinaryBuilt } from "./runner/ensure-binary";
import { StreamDemux } from "./runner/stream-demux";

// The jobs this file used to do inline now live beside it under ./runner/, one concern per
// module: the fd-allocation contract and the ROW ID = NODE ID - 1 arithmetic
// (runner/stream-fds.ts), the stored topology counts (runner/counts.ts), the Go-error probe
// line (runner/go-errors.ts), the probe-log rotation and path layout (runner/probe-paths.ts),
// the two pure framing steps (runner/framing.ts), the per-spawn parse state
// (runner/parse-state.ts), and the whole read side — per-fd frame reassembly, per-owner
// probe decode, the last-frame replay cache — as ONE object the runner owns per spawn
// (runner/stream-demux.ts; the state is that process's bytes, held on the INSTANCE, never
// module state). What stays here is the runner itself: spawn + env, the listener wiring,
// lifecycle, and writeStdin.
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
  private goErrorsFile: string | undefined;
  // The read side of ONE spawned process: parse state, probe paths, probeTrace, dead
  // streams, and the four last-frame replay caches all live on it (runner/stream-demux.ts).
  // A FRESH one is minted at every spawn, which is the single reset point every restart
  // path funnels through — a prior process's leftover partial frame or cached keyframe can
  // never reach the next process's stream, because it is not on the same object. The one
  // built here (counts 0, no probe paths, gen 0) is what an unspawned runner reads, so the
  // getLast* accessors answer before the first run() exactly as they did as fields.
  private demux: StreamDemux = this.newDemux(undefined, false, 0, 0, 0);

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

  /** One spawn's read side. The callbacks are the two things the demux cannot own: the
   *  output channel and the go-errors probe log are the RUNNER's (the stderr and exit
   *  handlers write the same log), and onSnapshot belongs to whoever constructed the
   *  runner. Read through `this` lazily so this can also be the field initializer. */
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
    this.goErrorsFile = probePaths.goErrorsFile;
    const probeTrace = isProbeTraceEnabled();
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
    const edgeCount = edgeCountRaw > MAX_EDGE_STREAMS ? 0 : edgeCountRaw;
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
    const nodeCount = nodeCountRaw > MAX_NODE_STREAMS ? 0 : nodeCountRaw;
    if (nodeCountRaw > MAX_NODE_STREAMS) {
      // Same reasoning as the edge-count case above.
      const msg = `node count ${nodeCountRaw} exceeds MAX_NODE_STREAMS (${MAX_NODE_STREAMS}); disabling ALL dedicated per-node NODE/INTERIOR streams for this run`;
      this.channel.appendLine(`\n[${msg}]`);
      appendGoError(probePaths.goErrorsFile, msg);
    }
    const nodeBaseFd = EDGE_BASE_FD + edgeCount;
    const interiorBaseFd = nodeBaseFd + nodeCount;
    // driveBaseFd sits right after the interior range: nodeCount * DRIVE_SLOTS_PER_NODE
    // dedicated fds, one PER (node row, drive slot) — see DRIVE_SLOTS_PER_NODE's doc
    // comment and Buffer/stream_fds.go's StreamKindDrive. Required in lockstep with
    // "node"/"interior" (see the streamFDsEnvParts push below and main.go's matching
    // three-way check) — Go falls back to a loud stderr message and unwired streams
    // rather than a startup panic if this ever drifts from what Go expects (never a
    // crash-loop; see the panic-avoidance note on that fallback in main.go).
    const driveBaseFd = interiorBaseFd + nodeCount;
    // A FRESH read side for this spawn: a prior process's leftover partial frame must
    // never prefix this one's stream (see freshStreamState). This is the single reset
    // point every restart path funnels through, including the looping respawn.
    //
    // It also drops the cached keyframes, which belong to the PRIOR process. Without that,
    // a webview remounting in the window between "ready" and the new process's first
    // frame would be replayed the previous process's frames via getLastViewFrame()/
    // getLastEdgeFrames()/etc. The freshly spawned Go emits its full state again, so
    // continuity is preserved by that emit — not by re-serving one process's bytes as
    // another's.
    const demux = this.newDemux(probePaths, probeTrace, edgeCount, nodeCount, this.spawnGen);
    this.demux = demux;
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
    for (let i = 0; i < edgeCount; i++) stdio.push("pipe");
    for (let i = 0; i < nodeCount; i++) stdio.push("pipe");
    for (let i = 0; i < nodeCount; i++) stdio.push("pipe");
    for (let i = 0; i < nodeCount * DRIVE_SLOTS_PER_NODE; i++) stdio.push("pipe");
    const streamFDsEnvParts = [`view:${VIEW_FD}`];
    if (edgeCount > 0) streamFDsEnvParts.push(`edge:${EDGE_BASE_FD}`);
    // Go's stream_fds.go / main.go only wires the per-node node+interior+drive streams
    // when "node", "interior", AND "drive" env entries ALL resolve — always emit all
    // three together (main.go's three-way check treats a partial set the same as none).
    if (nodeCount > 0) {
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
        WIREFOLD_EDGE_BEAD_TRACE: probeTrace ? "1" : "0",
      },
    });
    // Flush any framed binary records buffered before this spawn (writeStdin queued them).
    if (this.pendingStdin.length > 0) {
      for (const rec of this.pendingStdin) this.proc.stdin?.write(rec);
      this.pendingStdin = [];
    }
    this.proc.stdout?.on("data", (d: Buffer) => demux.handleStdout(d.toString()));
    // stdio index 3 is a reserved, unused pipe (see the stdio comment above) — nothing
    // reads it; Go writes nothing to it.
    // VIEW_FD: the dedicated view-stream pipe. Cast needed because Node's ChildProcess
    // types only narrow stdio[0..2]; higher indices are typed as Readable|null via the
    // array form.
    const viewFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[VIEW_FD];
    if (viewFd) {
      viewFd.on("data", (d: Buffer) => demux.handleViewFd(d));
    }
    // Per-edge dedicated pipes: EDGE_BASE_FD..EDGE_BASE_FD+edgeCount-1, one per edge row.
    for (let row = 0; row < edgeCount; row++) {
      const fdIdx = EDGE_BASE_FD + row;
      const edgeFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[fdIdx];
      if (edgeFd) {
        edgeFd.on("data", (d: Buffer) => demux.handleEdgeFd(row, d));
      }
    }
    // Per-node dedicated pipes: nodeBaseFd..nodeBaseFd+nodeCount-1 (NODE stream, geometry+
    // ports+label) and interiorBaseFd..interiorBaseFd+nodeCount-1 (INTERIOR stream, that
    // node's own interior beads — a separate goroutine's fd, see NODE_BASE_FD's doc comment).
    for (let row = 0; row < nodeCount; row++) {
      const nodeFdIdx = nodeBaseFd + row;
      const nodeFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[nodeFdIdx];
      if (nodeFd) {
        nodeFd.on("data", (d: Buffer) => demux.handleNodeFd(row, d));
      }
      const interiorFdIdx = interiorBaseFd + row;
      const interiorFd = (this.proc.stdio as (NodeJS.ReadableStream | null)[])[interiorFdIdx];
      if (interiorFd) {
        interiorFd.on("data", (d: Buffer) => demux.handleInteriorFd(row, d));
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
          driveFd.on("data", (d: Buffer) => demux.handleDriveFd(row, slot, d));
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

  // The replay cache the "ready" handler serves a remounted webview from. It is the
  // current spawn's demux that holds it (see StreamDemux) — there is no `resend` command
  // on the bridge, so these four accessors ARE how a webview that lost its state gets a
  // full scene back (.claude/rules/bridge-surface.md).

  /** The most recent VIEW-stream frame (camera+overlay+scene), or undefined if none has
   *  arrived yet. */
  getLastViewFrame(): ArrayBuffer | undefined {
    return this.demux.getLastViewFrame();
  }

  /** The most recent frame for EVERY cached edge row, or an empty array if none has
   *  arrived yet. */
  getLastEdgeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastEdgeFrames();
  }

  /** The most recent frame for EVERY cached node row from the dedicated NODE stream. */
  getLastNodeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastNodeFrames();
  }

  /** The most recent frame for EVERY cached node row from the dedicated INTERIOR stream. */
  getLastInteriorFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return this.demux.getLastInteriorFrames();
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
