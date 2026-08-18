import { useSyncExternalStore } from "react";
import { columnU8, columnU32, subscribeColumns } from "../../../../../Buffer/column-values";
import {
  COL_STREAM_OVERLAY_SCENE_EDITABLE,
  COL_STREAM_OVERLAY_SCENE_KINDS,
} from "../../../../../Buffer/column-streams-gen";

export function readSceneEditable(): boolean {
  return columnU8(COL_STREAM_OVERLAY_SCENE_EDITABLE) !== 0;
}

export function useSceneEditable(): boolean {
  return useSyncExternalStore(subscribeColumns, readSceneEditable, readSceneEditable);
}

export function readSceneKinds(): number {
  return columnU32(COL_STREAM_OVERLAY_SCENE_KINDS);
}

export function useSceneKinds(): number {
  return useSyncExternalStore(subscribeColumns, readSceneKinds, readSceneKinds);
}
