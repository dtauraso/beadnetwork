import type { HostToWebviewMsg } from "../../../Input/messages";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM, BUF_BLOCK_TAG_BEAD_STREAM } from "../../../Buffer/frame-tags";
import { appendViewProbe, appendEdgeProbe, appendNodeProbe, appendBeadProbe, appendInteriorProbe } from "./probe-append";
import { splitFrames } from "../framing";

export interface FrameDispatchContext {
  probeTrace: boolean;
  gen: number;
  onSnapshot?: (msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void;
  dispatch(
    key: string,
    carry: Buffer,
    chunk: Buffer,
    storeRest: (rest: Buffer) => void,
    errorContext: string,
    onFrames: (frames: ArrayBuffer[]) => void,
  ): void;
}

export function makeFrameDispatchContext(
  probeTrace: boolean,
  gen: number,
  onSnapshot: ((msg: HostToWebviewMsg & { type: "buffer-snapshot" }) => void) | undefined,
  onError: (msg: string) => void,
): FrameDispatchContext {
  const deadStreams: Set<string> = new Set();
  return {
    probeTrace,
    gen,
    onSnapshot,
    dispatch(key, carry, chunk, storeRest, errorContext, onFrames) {
      if (deadStreams.has(key)) return;
      const { frames, rest, error } = splitFrames(carry, chunk);
      storeRest(rest);
      if (error) {
        deadStreams.add(key);
        onError(`${errorContext}: ${error}`);
      }
      onFrames(frames);
    },
  };
}

export function dispatchViewFrames(
  ctx: FrameDispatchContext,
  carry: Buffer,
  chunk: Buffer,
  storeRest: (rest: Buffer) => void,
  probeFile: string | undefined,
  setLast: (ab: ArrayBuffer) => void,
): void {
  ctx.dispatch("view", carry, chunk, storeRest, "handleViewFd", (frames) => {
    for (const ab of frames) {
      appendViewProbe(probeFile, ab, ctx.probeTrace);
      setLast(ab.slice(0));
      if (ctx.onSnapshot) {
        ctx.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_VIEW, gen: ctx.gen });
      }
    }
  });
}

export function dispatchEdgeFrames(
  ctx: FrameDispatchContext,
  row: number,
  carry: Buffer,
  chunk: Buffer,
  storeRest: (rest: Buffer) => void,
  probeFile: string | undefined,
  setLast: (row: number, ab: ArrayBuffer) => void,
): void {
  ctx.dispatch(`edge:${row}`, carry, chunk, storeRest, `handleEdgeFd(row=${row})`, (frames) => {
    for (const ab of frames) {
      appendEdgeProbe(probeFile, row, ab, ctx.probeTrace);
      setLast(row, ab.slice(0));
      if (ctx.onSnapshot) {
        ctx.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_EDGE_STREAM, row, gen: ctx.gen });
      }
    }
  });
}

export function dispatchNodeFrames(
  ctx: FrameDispatchContext,
  row: number,
  errorContextLabel: string,
  carry: Buffer,
  chunk: Buffer,
  storeRest: (rest: Buffer) => void,
  probeFile: string | undefined,
  setLast: (row: number, ab: ArrayBuffer) => void,
): void {
  ctx.dispatch(`node:${row}`, carry, chunk, storeRest, errorContextLabel, (frames) => {
    for (const ab of frames) {
      appendNodeProbe(probeFile, row, ab, ctx.probeTrace);
      setLast(row, ab.slice(0));
      if (ctx.onSnapshot) {
        ctx.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_NODE_STREAM, row, gen: ctx.gen });
      }
    }
  });
}

export function dispatchBeadFrames(
  ctx: FrameDispatchContext,
  row: number,
  carry: Buffer,
  chunk: Buffer,
  storeRest: (rest: Buffer) => void,
  probeFile: string | undefined,
  setLast: (row: number, ab: ArrayBuffer) => void,
): void {
  ctx.dispatch(`bead:${row}`, carry, chunk, storeRest, `handleBeadFd(row=${row})`, (frames) => {
    for (const ab of frames) {
      appendBeadProbe(probeFile, row, ab, ctx.probeTrace);
      setLast(row, ab.slice(0));
      if (ctx.onSnapshot) {
        ctx.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_BEAD_STREAM, row, gen: ctx.gen });
      }
    }
  });
}

export function dispatchInteriorLikeFrames(
  ctx: FrameDispatchContext,
  key: string,
  row: number,
  errorContextLabel: string,
  carry: Buffer,
  chunk: Buffer,
  storeRest: (rest: Buffer) => void,
  probeFile: string | undefined,
  assertsSlots: boolean,
  setLast: (row: number, ab: ArrayBuffer) => void,
): void {
  ctx.dispatch(key, carry, chunk, storeRest, errorContextLabel, (frames) => {
    for (const ab of frames) {
      appendInteriorProbe(probeFile, row, ab, ctx.probeTrace);
      if (!assertsSlots) continue;
      setLast(row, ab.slice(0));
      if (ctx.onSnapshot) {
        ctx.onSnapshot({ type: "buffer-snapshot", buffer: ab, tag: BUF_BLOCK_TAG_INTERIOR_STREAM, row, gen: ctx.gen });
      }
    }
  });
}
