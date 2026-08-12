import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../../../scene/view-blocks";

export interface SceneTabsState {
  names: string[];
  selected: number;
}

let cached: SceneTabsState = { names: [], selected: 0 };

function sameTabs(a: SceneTabsState, b: SceneTabsState): boolean {
  return (
    a.selected === b.selected &&
    a.names.length === b.names.length &&
    a.names.every((n, i) => n === b.names[i])
  );
}

function getSceneTabs(): SceneTabsState {
  const blocks = getViewBlocks();
  if (!blocks) return cached;
  const next: SceneTabsState = { names: blocks.sceneTabs, selected: blocks.sceneTabSelected };
  if (!sameTabs(cached, next)) cached = next;
  return cached;
}

export function useSceneTabs(): SceneTabsState {
  return useSyncExternalStore(subscribeViewBlocks, getSceneTabs);
}

