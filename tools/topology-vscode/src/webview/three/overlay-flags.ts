// overlay-flags.ts — a row-keyed READ resource over the buffer's Overlay columns.
//
// The overlay on/off state is Go-owned: Go flips it on the
// `edit op=update kind=overlays` command and streams the updated flags into the
// dedicated VIEW stream's Overlay block (nodes/Wiring's MoveDispatch, Buffer/view_stream_frame.go).
// This module REFLECTS those Go-owned
// columns for widgets that must re-render when a flag flips (the overlay toggle
// control, NavGuides gating). It is NOT a domain store — it authors nothing; it only
// decodes the latest snapshot's Overlay row and subscribes to snapshot arrivals so a
// toggle round-trips to the displayed state.

import { useSyncExternalStore } from "react";
import type { OverlayFlag } from "../../messages";
import { getNodeFrame, subscribeNodeStreamBlocks } from "./node-stream-blocks";
import { getViewBlocks, subscribeViewBlocks } from "./view-blocks";
import {
  readOverlaySceneTori,
  readOverlayScenePoles,
  readOverlayNodePoles,
  readOverlaySelSpherePoles,
  readOverlayHandholds,
  readOverlayLabelsGlobal,
  readOverlayOverlaysVis,
  readOverlayDragNodeRow,
  readOverlayGroupLenTime,
  readOverlayGroupLenInput,
  readOverlayGroupLenGate,
} from "../../schema/buffer-layout";
import { nodeLabel } from "./buffer-decode";

// Keyed by OverlayFlag. Polarity is MIXED — a historical wart worth stating plainly, since
// the ViewerState key names it mirrored are gone (that state island was deleted once Go
// owned scene persistence):
//   • most flags are visible-sense (true = shown) — <x>Visible
//   • labelsGlobal is HIDDEN-sense (true = hidden) — labelsGlobalHidden. The buffer
//     stores visible-sense, so we invert that one here.
export type OverlayFlagVals = Record<OverlayFlag, boolean>;

// Cache so getSnapshot returns a STABLE object identity while the flags are unchanged
// (useSyncExternalStore compares by identity; a fresh object every 60fps snapshot would
// re-render every frame). We recompute the flags each call — cheap — and only mint a
// new OverlayFlagVals when a flag actually flips, detected by VALUE equality.
let cachedVals: OverlayFlagVals | null = null;

function overlayFlagsEqual(a: OverlayFlagVals, b: OverlayFlagVals): boolean {
  return (
    a.tori === b.tori &&
    a.scenePoles === b.scenePoles &&
    a.nodePoles === b.nodePoles &&
    a.selSpherePoles === b.selSpherePoles &&
    a.handholds === b.handholds &&
    a.labelsGlobal === b.labelsGlobal &&
    a.overlays === b.overlays
  );
}

/** Decode the latest snapshot's Overlay row into store-polarity booleans, or null if no
 *  snapshot / decode failure. Stable identity while unchanged. Each flag is read ONCE into
 *  its named field (no bit-packing intermediary), and change is detected by value equality —
 *  so there is a single ordering (the field assignments), not three parallel ones. */
export function readOverlayFlags(): OverlayFlagVals | null {
  const blocks = getViewBlocks();
  if (!blocks) return cachedVals;
  const v = blocks.overlayView;
  const next: OverlayFlagVals = {
    tori: !!readOverlaySceneTori(v),
    scenePoles: !!readOverlayScenePoles(v),
    nodePoles: !!readOverlayNodePoles(v),
    selSpherePoles: !!readOverlaySelSpherePoles(v),
    handholds: !!readOverlayHandholds(v),
    // hidden-sense: buffer stores VISIBLE, store field is *Hidden → invert this one.
    labelsGlobal: !readOverlayLabelsGlobal(v),
    overlays: !!readOverlayOverlaysVis(v),
  };
  if (cachedVals && overlayFlagsEqual(cachedVals, next)) return cachedVals;
  cachedVals = next;
  return cachedVals;
}

/** React hook: re-renders the caller when any overlay flag flips (Go-owned). Returns
 *  null until the first snapshot lands. */
export function useOverlayFlags(): OverlayFlagVals | null {
  return useSyncExternalStore(subscribeViewBlocks, readOverlayFlags, readOverlayFlags);
}

/** Decode the row index of the node currently being dragged (Overlay block
 *  DragNodeRow column, Go's gesture FSM g.dragNode resolved via NodeRowFor), or -1
 *  when idle. Returns -1 if no snapshot / decode failure yet. */
export function readDragNodeRow(): number {
  const blocks = getViewBlocks();
  if (!blocks) return -1;
  return readOverlayDragNodeRow(blocks.overlayView);
}

/** React hook: re-renders the caller when the dragged node's row changes (drag
 *  start/end). Returns -1 when no drag is in progress. */
export function useDragNodeRow(): number {
  return useSyncExternalStore(subscribeViewBlocks, readDragNodeRow, readDragNodeRow);
}

/** Decode the human name of the node currently being dragged, by resolving
 *  DragNodeRow against the Node block's own Label section — identity rides row
 *  index, the name is never sidecar'd (same plumbing readAbcDragRows uses per-row).
 *  Returns "" when idle or the row can't be resolved yet (no node frame decoded). */
export function readDraggedNodeName(): string {
  const row = readDragNodeRow();
  if (row < 0) return "";
  const decoded = getNodeFrame();
  if (!decoded || row >= decoded.nodeCount) return "";
  return nodeLabel(decoded, row);
}

/** subscribeDraggedNodeName subscribes to BOTH the VIEW stream (DragNodeRow lives in
 *  the Overlay block) and the node stream (the Label section lives in the Node block)
 *  — either arrival can change the resolved name (drag start/end flips the row; a
 *  node-frame arrival while dragging can (re)decode its label). */
function subscribeDraggedNodeName(fn: () => void): () => void {
  const unsubView = subscribeViewBlocks(fn);
  const unsubNode = subscribeNodeStreamBlocks(fn);
  return () => {
    unsubView();
    unsubNode();
  };
}

/** React hook: re-renders the caller when the dragged node's name changes (drag
 *  start/end, or a node-frame arrival while dragging). Returns "" when idle. */
export function useDraggedNodeName(): string {
  return useSyncExternalStore(subscribeDraggedNodeName, readDraggedNodeName, readDraggedNodeName);
}

/** The "distance home button" toolbar panel's 3 group max-pair-lengths, in Go's
 *  distanceGroupOrder (nodes/Wiring/distance_groups.go): time, input, gate. Read-only
 *  reflect of the Overlay block's GroupLenTime/GroupLenInput/GroupLenGate columns — Go
 *  computes these fresh every VIEW-frame emit; TS holds no group definitions. */
export interface DistanceGroupLens {
  time: number;
  input: number;
  gate: number;
}

let cachedGroupLens: DistanceGroupLens | null = null;

function distanceGroupLensEqual(a: DistanceGroupLens, b: DistanceGroupLens): boolean {
  return a.time === b.time && a.input === b.input && a.gate === b.gate;
}

/** Decode the current 3 group max-pair-lengths, or null if no snapshot yet. Stable
 *  identity while unchanged (useSyncExternalStore compares by identity). */
export function readDistanceGroupLens(): DistanceGroupLens | null {
  const blocks = getViewBlocks();
  if (!blocks) return cachedGroupLens;
  const overlayView = blocks.overlayView;
  const next: DistanceGroupLens = {
    time: readOverlayGroupLenTime(overlayView),
    input: readOverlayGroupLenInput(overlayView),
    gate: readOverlayGroupLenGate(overlayView),
  };
  if (cachedGroupLens && distanceGroupLensEqual(cachedGroupLens, next)) return cachedGroupLens;
  cachedGroupLens = next;
  return cachedGroupLens;
}

/** React hook: re-renders the caller when any of the 3 group max-pair-lengths change. */
export function useDistanceGroupLens(): DistanceGroupLens | null {
  return useSyncExternalStore(subscribeViewBlocks, readDistanceGroupLens, readDistanceGroupLens);
}

