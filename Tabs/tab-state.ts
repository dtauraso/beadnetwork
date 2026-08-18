import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "../tools/topology-vscode/src/webview/three/scene/view-blocks";

export interface TabsState {
  names: string[];
  selected: number;
}

let cached: TabsState = { names: [], selected: 0 };

function sameTabs(a: TabsState, b: TabsState): boolean {
  return (
    a.selected === b.selected &&
    a.names.length === b.names.length &&
    a.names.every((n, i) => n === b.names[i])
  );
}

function getTabs(): TabsState {
  const blocks = getViewBlocks();
  if (!blocks) return cached;
  const next: TabsState = { names: blocks.sceneTabs, selected: blocks.sceneTabSelected };
  if (!sameTabs(cached, next)) cached = next;
  return cached;
}

export function useTabs(): TabsState {
  return useSyncExternalStore(subscribeViewBlocks, getTabs);
}

