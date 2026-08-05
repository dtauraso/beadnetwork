// scene-tabs.ts — a READ resource over the VIEW frame's Go-owned scene tab strip.
//
// Sibling of overlay-flags.ts, and allowed by check-no-webview-state.sh for the same
// reason: it REFLECTS Go (decodes the latest VIEW frame's tab section) and authors
// nothing. The strip's rendering lives in SceneTabs.tsx; keeping the two apart is what
// makes "this file only reflects" checkable per-file rather than per-function.
//
// Everything about a tab is GO-OWNED (nodes/Wiring/scene_tabs.go): the list, the labels,
// which one is selected, what each one loads, and how the switch is performed. This file
// reflects the strip off the VIEW frame and forwards a click as ONE addressed edit
// (edit-update kind="scene" attr="selected", carrying the tab INDEX — no name, no path).
// It is the same shape as the overlay toggles: read-only reflection via
// useSyncExternalStore, fire-and-forget send, no local selection state.
//
// Deliberately NO optimistic highlight. Clicking a tab does not move the highlight here —
// the highlight follows the buffer, so it moves when the newly-loaded scene's first VIEW
// frame arrives. An optimistic one would be TS authoring the selection, and it would lie
// for the whole reload if the switch failed (a write error leaves Go on the old scene,
// scene_tabs.go's SelectScene).

import { useSyncExternalStore } from "react";
import { getViewBlocks, subscribeViewBlocks } from "./view-blocks";

export interface SceneTabsState {
  names: string[];
  selected: number;
}

// Stable-identity cache: useSyncExternalStore compares snapshots by identity, and the VIEW
// frame arrives every tick, so minting a fresh object each call would re-render the strip
// at frame rate. The tabs are constant for a process's lifetime (switching ends the
// process), so this only ever mints twice: once empty, once when the first frame lands.
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

/** Reflect the Go-owned tab strip. Empty names until the first VIEW frame lands, and empty
 *  forever for an untabbed anchor — both render nothing. */
export function useSceneTabs(): SceneTabsState {
  return useSyncExternalStore(subscribeViewBlocks, getSceneTabs);
}

