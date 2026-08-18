import { useSyncExternalStore } from "react";
import { columnF32 } from "../../../../../Buffer/column-values";
import { subscribeFrame } from "../../../frame-tick";
import { COL_STREAM_OVERLAY_SPEED } from "../../../../../Buffer/column-streams-gen";

export function readPlaybackSpeed(): number | null {
  return columnF32(COL_STREAM_OVERLAY_SPEED);
}

export function usePlaybackSpeed(): number | null {
  return useSyncExternalStore(subscribeFrame, readPlaybackSpeed, readPlaybackSpeed);
}
