






import type * as cp from "child_process";
import { VIEW_FD, EDGE_BASE_FD, DRIVE_SLOTS_PER_NODE } from "./stream-fds";
import type { StreamDemux } from "./stream-demux";
import type { SpawnLayout } from "./spawn-layout";

export function attachStreamListeners(proc: cp.ChildProcess, demux: StreamDemux, layout: SpawnLayout): void {
  const { edgeCount, nodeCount, nodeBaseFd, interiorBaseFd, driveBaseFd } = layout;
  proc.stdout?.on("data", (d: Buffer) => demux.handleStdout(d.toString()));




  const viewFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[VIEW_FD];
  if (viewFd) {
    viewFd.on("data", (d: Buffer) => demux.handleViewFd(d));
  }

  for (let row = 0; row < edgeCount; row++) {
    const fdIdx = EDGE_BASE_FD + row;
    const edgeFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[fdIdx];
    if (edgeFd) {
      edgeFd.on("data", (d: Buffer) => demux.handleEdgeFd(row, d));
    }
  }



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






    for (let slot = 0; slot < DRIVE_SLOTS_PER_NODE; slot++) {
      const driveFdIdx = driveBaseFd + row * DRIVE_SLOTS_PER_NODE + slot;
      const driveFd = (proc.stdio as (NodeJS.ReadableStream | null)[])[driveFdIdx];
      if (driveFd) {
        driveFd.on("data", (d: Buffer) => demux.handleDriveFd(row, slot, d));
      }
    }
  }
}
