// snapshot-buffer.ts — module-level sinks for the latest binary frames from Go's dedicated
// per-owner stream fds (memory/feedback_no_single_writer_bridge.md). There is no combined
// SCENE/BEAD/NODE/EDGE fallback anymore — WIREFOLD_STREAM_FDS is mandatory
// (memory/feedback_no_single_writer_bridge.md's final step deleted the central
// accumulator that used to write those combined fallback frames).
//
// Separated from main.tsx so that buffer-scene.tsx / BeadInstances.tsx can read them
// without creating a circular import (main.tsx → ThreeView → buffer-scene → main).
// Not a Zustand store: these are plain module-level cells, identical in role to a
// global register — each holds exactly one pointer and never notifies beyond its own
// listener set.
//
// One singleton cell exists (VIEW), plus keyed maps for the per-row streams (edge/node/
// interior, below):
//   - the VIEW cell (setLatestViewFrame/getLatestViewFrame/subscribeViewFrame) — the
//     camera+overlay+scene frame arriving on its own dedicated pipe (VIEW_FD in
//     runCommand.ts), the first stream migrated off fd 3 per the no-single-writer-bridge
//     rule. Null until the first frame has landed.

import { decodeViewFrame } from "./three/buffer-decode";

let latestViewFrame: ArrayBuffer | null = null;

// Listeners are notified after each new snapshot lands. This is NOT a domain store —
// it is the same one-pointer cell as before; the listener set only lets a widget that
// must re-render on Go-owned change (e.g. the overlay toggle control, whose displayed
// state reflects the buffer's Go-owned overlay columns) subscribe via
// useSyncExternalStore instead of polling every frame. Nothing here authors state.
type SnapshotListener = () => void;
const viewListeners = new Set<SnapshotListener>();

/** Called by main.tsx whenever a new VIEW buffer-snapshot message arrives (the dedicated
 * view-fd stream — see this file's header comment). */
export function setLatestViewFrame(buf: ArrayBuffer, gen = 0): void {
  latestViewFrame = buf;
  noteGen(gen);
  for (const fn of viewListeners) fn();
}

/**
 * ONE ROW TABLE PER PROCESS.
 *
 * The per-row maps below used to be single module-level Maps keyed by ROW ALONE, so the
 * ring's row 3 and the pair's row 3 were the same cell: two tabs sharing one keyspace, each
 * able to overwrite the other. Switching from a 9-node scene to a 2-node one left rows 2..8
 * holding the previous scene's frames and the renderer drew both diagrams at once.
 *
 * An earlier fix CLEARED the maps on a scene change. That is fine for node geometry, re-sent
 * every tick and repopulated within a frame, and wrong for INTERIOR beads, which a node
 * writes only when its held value CHANGES — a cleared interior row has nothing to re-send it,
 * so a Pulse holding a steady value loses its bead until that value happens to change, which
 * may be never.
 *
 * So the table is chosen by the frame's own GENERATION, stamped by the extension host at
 * spawn (runCommand.ts's spawnGen) and carried on the message. A generation is decided
 * before its process writes a byte, so routing is WIRED rather than inferred: a frame from a
 * dead process cannot land in a live process's table no matter what order things arrive in.
 * Since a tab switch restarts Go, that is also what keeps one tab's rows unreachable from
 * another's.
 *
 * Per PROCESS, not per scene name, and deliberately so: ring → pair → ring gives the second
 * ring visit its own table rather than reusing the first visit's. That is correct — it is a
 * different process, and a fresh process re-emits what it has. Nothing stale survives a
 * restart, which is the property the clear was clumsily reaching for.
 *
 * Old generations are dropped as new ones appear (see reap below), so this cannot grow
 * without bound across a long session of restarts.
 */
function genTable<T>(tables: Map<number, Map<number, T>>, gen: number): Map<number, T> {
  let t = tables.get(gen);
  if (!t) {
    t = new Map<number, T>();
    tables.set(gen, t);
    // Keep only the newest generation's table plus the one before it: a frame from the
    // outgoing process can still be in flight while the new one starts, and dropping its
    // table mid-relay would throw away a row that is about to be replaced anyway. Anything
    // older cannot receive another write.
    for (const key of tables.keys()) {
      if (key < gen - 1) tables.delete(key);
    }
  }
  return t;
}

// latestGen is the newest generation any frame has arrived under — which table READS come
// from. Writes address their own frame's generation, never this.
let latestGen = 0;

function noteGen(gen: number): void {
  if (gen > latestGen) latestGen = gen;
}

/** TEST-ONLY: forget the current generation AND every generation's table, so a fresh test starts from
 *  "nothing has arrived yet". Production has exactly one webview lifetime per page load and
 *  never needs this. */
export function resetSceneIdentityForTest(): void {
  latestGen = 0;
  edgeStreamTables.clear();
  nodeStreamTables.clear();
  interiorStreamTables.clear();
}

/** Called by three/view-blocks.ts (and tests) to read the most-recent VIEW frame. Null
 * until the first frame has landed. */
export function getLatestViewFrame(): ArrayBuffer | null {
  return latestViewFrame;
}

/** Subscribe to VIEW-frame arrivals; returns an unsubscribe fn (useSyncExternalStore shape). */
export function subscribeViewFrame(fn: SnapshotListener): () => void {
  viewListeners.add(fn);
  return () => {
    viewListeners.delete(fn);
  };
}

// --- per-edge dedicated streams (memory/feedback_no_single_writer_bridge.md) ---
//
// Unlike the singleton VIEW stream, there are MANY per-edge streams (one per edge row —
// see Buffer/stream_fds.go's StreamKindEdge / runCommand.ts's edge-fd range). Keyed by
// edge row (int) rather than one bare cell. Still a plain module-level Map, not a store —
// same one-pointer-per-key role as the cells above, just keyed.
const edgeStreamTables: Map<number, Map<number, ArrayBuffer>> = new Map();
const edgeStreamListeners = new Set<SnapshotListener>();

/** Called by main.tsx whenever a new per-edge dedicated-stream frame arrives (tag
 *  BUF_BLOCK_TAG_EDGE_STREAM, carrying `row`). */
export function setLatestEdgeStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  noteGen(gen);
  genTable(edgeStreamTables, gen).set(row, buf);
  for (const fn of edgeStreamListeners) fn();
}

/** Called by EdgeTube.tsx/BeadInstances.tsx to read every edge row's most-recent dedicated
 *  per-edge frame. Empty when the dedicated edge-fd path is not active (fallback launches
 *  never populate this map) — callers fall back to the single EDGE/BEAD cells above in
 *  that case. */
export function getLatestEdgeStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return genTable(edgeStreamTables, latestGen);
}

/** Subscribe to per-edge dedicated-stream arrivals (any row); returns an unsubscribe fn
 *  (useSyncExternalStore shape). */
export function subscribeEdgeStreamFrame(fn: SnapshotListener): () => void {
  edgeStreamListeners.add(fn);
  return () => {
    edgeStreamListeners.delete(fn);
  };
}

// --- per-node dedicated streams (memory/feedback_no_single_writer_bridge.md) ---
//
// TWO keyed maps, mirroring edgeStreamFrames' role but split by stream KIND: one node's
// own nodeMover (geometry+ports+label) and its OWN Update goroutine (interior beads) write
// to two DIFFERENT fds (see Buffer/stream_fds.go's StreamKindNode/StreamKindInterior), so
// they get two separate cells here, both keyed by node row. version counters let
// node-stream-blocks.ts's aggregator memoize its (necessarily-copying) rebuild instead of
// re-concatenating every consumer's per-frame read.
const nodeStreamTables: Map<number, Map<number, ArrayBuffer>> = new Map();
const interiorStreamTables: Map<number, Map<number, ArrayBuffer>> = new Map();
const nodeStreamListeners = new Set<SnapshotListener>();
const interiorStreamListeners = new Set<SnapshotListener>();
let nodeStreamVersion = 0;
let interiorStreamVersion = 0;

/** Called by main.tsx whenever a new per-node dedicated NODE-stream frame arrives (tag
 *  BUF_BLOCK_TAG_NODE_STREAM, carrying `row`). */
export function setLatestNodeStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  noteGen(gen);
  genTable(nodeStreamTables, gen).set(row, buf);
  nodeStreamVersion++;
  for (const fn of nodeStreamListeners) fn();
}

/** Called by node-stream-blocks.ts (and tests) to read every node row's most-recent
 *  dedicated NODE-stream frame. Empty when the dedicated node-fd path is not active
 *  (fallback launches never populate this map). */
export function getLatestNodeStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return genTable(nodeStreamTables, latestGen);
}

/** Monotonic counter bumped on every setLatestNodeStreamFrame call — a cheap memo key for
 *  the aggregator (rebuilding the aggregate DecodedNodeFrame is a full copy; this avoids
 *  redoing it when nothing changed since the last read). */
export function getNodeStreamVersion(): number {
  return nodeStreamVersion;
}

/** Subscribe to per-node NODE-stream arrivals (any row); returns an unsubscribe fn. */
export function subscribeNodeStreamFrame(fn: SnapshotListener): () => void {
  nodeStreamListeners.add(fn);
  return () => {
    nodeStreamListeners.delete(fn);
  };
}

/** Called by main.tsx whenever a new per-node dedicated INTERIOR-stream frame arrives (tag
 *  BUF_BLOCK_TAG_INTERIOR_STREAM, carrying `row`). */
export function setLatestInteriorStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  noteGen(gen);
  genTable(interiorStreamTables, gen).set(row, buf);
  interiorStreamVersion++;
  for (const fn of interiorStreamListeners) fn();
}

/** Called by node-stream-blocks.ts (and tests) to read every node row's most-recent
 *  dedicated INTERIOR-stream frame. */
export function getLatestInteriorStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return genTable(interiorStreamTables, latestGen);
}

/** Monotonic counter mirroring getNodeStreamVersion for the interior stream. */
export function getInteriorStreamVersion(): number {
  return interiorStreamVersion;
}

/** Subscribe to per-node INTERIOR-stream arrivals (any row); returns an unsubscribe fn. */
export function subscribeInteriorStreamFrame(fn: SnapshotListener): () => void {
  interiorStreamListeners.add(fn);
  return () => {
    interiorStreamListeners.delete(fn);
  };
}
