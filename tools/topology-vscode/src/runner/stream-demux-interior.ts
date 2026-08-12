import { nodeIdForRow } from "./stream-fds";
import { dispatchInteriorLikeFrames, type FrameDispatchContext } from "./probe/frame-dispatch";
import type { StreamParseState } from "./parse-state";

export function handleInteriorFdImpl(
  frameCtx: FrameDispatchContext,
  stream: StreamParseState,
  probeInteriorFile: string | undefined,
  row: number,
  chunk: Buffer,
  onFrame: (row: number, ab: ArrayBuffer) => void,
) {
  dispatchInteriorLikeFrames(
    frameCtx,
    `interior:${row}`,
    row,
    `handleInteriorFd(node=${nodeIdForRow(row)})`,
    stream.interiorBufs[row] ?? Buffer.alloc(0),
    chunk,
    (rest) => { stream.interiorBufs[row] = rest; },
    probeInteriorFile,
    true,
    onFrame,
  );
}

export function handleDriveFdImpl(
  frameCtx: FrameDispatchContext,
  stream: StreamParseState,
  probeInteriorFile: string | undefined,
  row: number,
  slot: number,
  chunk: Buffer,
  onFrame: (row: number, ab: ArrayBuffer) => void,
) {
  dispatchInteriorLikeFrames(
    frameCtx,
    `drive:${row}:${slot}`,
    row,
    `handleDriveFd(node=${nodeIdForRow(row)}, slot=${slot})`,
    stream.driveBufs[row]?.[slot] ?? Buffer.alloc(0),
    chunk,
    (rest) => {
      if (!stream.driveBufs[row]) stream.driveBufs[row] = [];
      stream.driveBufs[row][slot] = rest;
    },
    probeInteriorFile,
    false,
    onFrame,
  );
}
