// Wires ONE spawned process's dedicated stdio pipes to its demux. Pure plumbing: every
// dedicated fd this spawn opened (view, one per edge row, one NODE + one INTERIOR per node
// row, one DRIVE per (node row, drive slot)) gets exactly one `data` listener that hands the
// bytes to the matching StreamDemux method — no framing, no decode, no state, that all lives
// in StreamDemux itself. Split out of runCommand.ts's run() because this block never touches
// runner lifecycle fields (this.channel/this.cancelled/...); it only needs the process, the
// demux, and the fd layout that spawned it.
import type * as cp from "child_process";
import { VIEW_FD, EDGE_BASE_FD, DRIVE_SLOTS_PER_NODE } from "./stream-fds";
import type { StreamDemux } from "./stream-demux";
import type { SpawnLayout } from "./spawn-layout";

export function attachStreamListeners(proc: cp.ChildProcess, demux: StreamDemux, layout: SpawnLayout): void {
  const { edgeCount, nodeCount, nodeBaseFd, interiorBaseFd, driveBaseFd } = layout;
  proc.stdout?.on("data", (d: Buffer) => demux.handleStdout(d.toString()));
  // stdio index 3 is a reserved, unused pipe (see run()'s stdio comment) — nothing reads
  // it; Go writes nothing to it.
  // VIEW_FD: the dedicated view-stream pipe. Cast needed because Node's ChildProcess types
  // only narrow stdio[0..2]; higher indices are typed as Readable|null via the array form.
  const viewFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[VIEW_FD];
  if (viewFd) {
    viewFd.on("data", (d: Buffer) => demux.handleViewFd(d));
  }
  // Per-edge dedicated pipes: EDGE_BASE_FD..EDGE_BASE_FD+edgeCount-1, one per edge row.
  for (let row = 0; row < edgeCount; row++) {
    const fdIdx = EDGE_BASE_FD + row;
    const edgeFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[fdIdx];
    if (edgeFd) {
      edgeFd.on("data", (d: Buffer) => demux.handleEdgeFd(row, d));
    }
  }
  // Per-node dedicated pipes: nodeBaseFd..nodeBaseFd+nodeCount-1 (NODE stream, geometry+
  // ports+label) and interiorBaseFd..interiorBaseFd+nodeCount-1 (INTERIOR stream, that
  // node's own interior beads — a separate goroutine's fd, see NODE_BASE_FD's doc comment).
  for (let row = 0; row < nodeCount; row++) {
    const nodeFdIdx = nodeBaseFd + row;
    const nodeFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[nodeFdIdx];
    if (nodeFd) {
      nodeFd.on("data", (d: Buffer) => demux.handleNodeFd(row, d));
    }
    const interiorFdIdx = interiorBaseFd + row;
    const interiorFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[interiorFdIdx];
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
      const driveFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[driveFdIdx];
      if (driveFd) {
        driveFd.on("data", (d: Buffer) => demux.handleDriveFd(row, slot, d));
      }
    }
  }
}
