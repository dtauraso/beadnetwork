import { useSyncExternalStore } from "react";
import { columnU32 } from "../../../../../Buffer/column-values";
import { subscribeFrame } from "../../../frame-tick";
import { COL_STREAM_OVERLAY_EDIT_REFUSED } from "../../../../../Buffer/column-streams-gen";

export function readEditRefused(): number {
  return columnU32(COL_STREAM_OVERLAY_EDIT_REFUSED);
}

export function useEditRefused(): number {
  return useSyncExternalStore(subscribeFrame, readEditRefused, readEditRefused);
}
