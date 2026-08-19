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

