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
  readOverlayDoubleLinks,
  readOverlayDragNodeRow,
  readNodeGotDragMsg,
  readNodeDragDeltaA,
  readNodeDragDeltaB,
  readNodeDragDeltaC,
  readNodeDragRequantCount,
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
    a.overlays === b.overlays &&
    a.doubleLinks === b.doubleLinks
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
    doubleLinks: !!readOverlayDoubleLinks(v),
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

/** Decode the running time-node abc-drag "received" total by SUMMING each node row's
 *  OWN cumulative DragRequantCount column (Node block) — the per-recipient count that
 *  replaced the old central Overlay.AbcDragCount. That central counter lived on a
 *  cross-goroutine channel (abcDragCh) a fast drag's pointer-input load could starve,
 *  silently dropping ticks (the node-7-target-only-drag bug); DragRequantCount is state
 *  on each recipient's OWN reliable node stream, so nothing can be dropped — this just
 *  sums what's already there. Returns 0 if no node frame decoded yet. */
export function readDragReceivedCount(): number {
  const decoded = getNodeFrame();
  if (!decoded) return 0;
  let total = 0;
  for (let row = 0; row < decoded.nodeCount; row++) {
    total += readNodeDragRequantCount(decoded.nodeView, row);
  }
  return total;
}

/** React hook: re-renders the caller as time.abc-drag events accumulate (Go-owned,
 *  summed across each recipient's own node-stream count; affirms the drag-log is
 *  happening live). Subscribes to the NODE stream (DragRequantCount lives in the Node
 *  block), not the VIEW stream. */
export function useDragReceivedCount(): number {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readDragReceivedCount, readDragReceivedCount);
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

/** One current-drag recipient: its display name plus the DRAGGED node's own
 *  quantized-triple delta (dA,dB,dC) that rode the message this recipient received
 *  (Node block DragDeltaA/B/C columns). */
export interface AbcDragRow {
  name: string;
  dA: number;
  dB: number;
  dC: number;
}

let cachedRowsKey = "\0";
let cachedRows: AbcDragRow[] = [];

/** Decode the current drag's recipient ROWS (name + received delta triple) from the
 *  Node block's per-row GotDragMsg flag + DragDeltaA/B/C columns. Go-owned and
 *  drag-scoped (cleared on KindAbcDragReset at drag start, which emits the cleared
 *  state — so an empty result is meaningful). Stable identity while unchanged —
 *  including a (0,0,0) delta row, which is real information ("got the message, didn't
 *  move"), not absence. */
export function readAbcDragRows(): AbcDragRow[] {
  const decoded = getNodeFrame();
  if (!decoded) return cachedRows;
  const rows: AbcDragRow[] = [];
  for (let row = 0; row < decoded.nodeCount; row++) {
    if (!readNodeGotDragMsg(decoded.nodeView, row)) continue;
    rows.push({
      name: nodeLabel(decoded, row),
      dA: readNodeDragDeltaA(decoded.nodeView, row),
      dB: readNodeDragDeltaB(decoded.nodeView, row),
      dC: readNodeDragDeltaC(decoded.nodeView, row),
    });
  }
  const key = rows.map((r) => `${r.name}\0${r.dA},${r.dB},${r.dC}`).join("\0");
  if (key === cachedRowsKey) return cachedRows;
  cachedRowsKey = key;
  cachedRows = rows;
  return cachedRows;
}

/** React hook: re-renders the caller when the current drag's recipient rows change —
 *  INCLUDING when cleared to empty at drag start. The GotDragMsg/DragDeltaA-C columns
 *  live in the Node block (node-owner-group frame), so this subscribes to node-frame
 *  arrivals, not scene-frame arrivals. */
export function useAbcDragRows(): AbcDragRow[] {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readAbcDragRows, readAbcDragRows);
}

