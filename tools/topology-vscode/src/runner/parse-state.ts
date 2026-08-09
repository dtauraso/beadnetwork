import { DRIVE_SLOTS_PER_NODE } from "./stream-fds";

// Per-process incremental parse state for Go's output byte-streams. Each field holds a
// partial fragment straddling a chunk boundary — a partial newline-delimited stdout line,
// and a partial length-prefixed binary frame per dedicated stream fd — and is meaningful
// ONLY within a single Go process's stream: a leftover tail is a fragment of THAT process's
// output. Its lifetime is therefore the process's, not the runner's. fd 3 itself carries no
// frames anymore (WIREFOLD_STREAM_FDS is mandatory — the old central accumulator and
// its fallback frames were deleted, memory/feedback_no_single_writer_bridge.md's final step); the pipe slot
// stays allocated (see run()) purely to keep the remaining fd numbering unchanged.
export interface StreamParseState {
  stdoutBuf: string;
  // Partial-frame carry-over for the dedicated VIEW fd (VIEW_FD).
  viewBuf: Buffer;
  // Partial-frame carry-over PER EDGE fd (index = edge row), same role as viewBuf but one
  // per dedicated edge pipe — each is its OWN pipe with its own chunk boundaries.
  edgeBufs: Buffer[];
  // Partial-frame carry-over PER NODE fd / PER INTERIOR fd (index = node row), same role
  // as edgeBufs — one per dedicated node/interior pipe.
  nodeBufs: Buffer[];
  interiorBufs: Buffer[];
  // Partial-frame carry-over PER DRIVE fd (index = node row, inner index = drive slot,
  // 0..DRIVE_SLOTS_PER_NODE-1) — one per dedicated per-DriveHeld-goroutine pipe (see
  // DRIVE_SLOTS_PER_NODE's doc comment). Kept SEPARATE from interiorBufs even though
  // handleDriveFd feeds the same decode/cache/relay path as handleInteriorFd: each drive
  // fd is its OWN pipe with its OWN chunk boundaries, and merging two physically distinct
  // byte streams into one carry buffer would reintroduce the exact framing desync this
  // whole per-goroutine-fd change exists to remove, just moved from Go's write side to
  // this file's read side.
  driveBufs: Buffer[][];
}

// freshStreamState mints empty parse state for a newly spawned process. run() calls it
// at every spawn so no respawn/restart path can carry a dead process's tail bytes into
// the next process — concatenating them would make splitFrames read a frame length from
// inside stale bytes and freeze (or silently starve) the scene. Binding the reset to the
// spawn, not to each exit handler, makes "start a process with leftover bytes" impossible
// to express rather than a rule every exit path must remember. edgeCount/nodeCount size
// edgeBufs/nodeBufs+interiorBufs to this spawn's fd ranges (0 when that dedicated path is
// off).
export function freshStreamState(edgeCount: number, nodeCount: number): StreamParseState {
  return {
    stdoutBuf: "",
    viewBuf: Buffer.alloc(0),
    edgeBufs: Array.from({ length: edgeCount }, () => Buffer.alloc(0)),
    nodeBufs: Array.from({ length: nodeCount }, () => Buffer.alloc(0)),
    interiorBufs: Array.from({ length: nodeCount }, () => Buffer.alloc(0)),
    driveBufs: Array.from({ length: nodeCount }, () => Array.from({ length: DRIVE_SLOTS_PER_NODE }, () => Buffer.alloc(0))),
  };
}
