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
  readOverlayCascadeLinks,
  readOverlayDragNodeRow,
  readOverlayGroupLenTime,
  readOverlayGroupLenInput,
  readOverlayGroupLenGate,
  readNodeGotDragMsg,
  readNodeDragDeltaA,
  readNodeDragDeltaB,
  readNodeDragDeltaC,
  readNodeDragRequantCount,
  readNodeGotForwardMsg,
  readNodeForwardDeltaA,
  readNodeForwardDeltaB,
  readNodeForwardDeltaC,
  readNodeForwardFromRow,
  readNodeCascadeRelay,
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
    a.cascadeLinks === b.cascadeLinks
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
    cascadeLinks: !!readOverlayCascadeLinks(v),
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

/** The dragged node's cascade relay behavior, as the word Go's Node.CascadeRelay column
 *  encodes (0 = flood, 1 = routed, 2 = terminus — see nodes/Wiring/node_mover.go's
 *  cascadeRelayClass, which is where the classification is DECIDED; this only names the
 *  value it streams). "" when nothing is dragged, or when the streamed value is one this
 *  build has no word for — an unnamed number is not rendered as if it were understood. */
export function readDraggedNodeRelay(): string {
  const row = readDragNodeRow();
  if (row < 0) return "";
  const decoded = getNodeFrame();
  if (!decoded || row >= decoded.nodeCount) return "";
  switch (readNodeCascadeRelay(decoded.nodeView, row)) {
    case 0:
      return "flood";
    case 1:
      return "routed";
    case 2:
      return "terminus";
    default:
      return "";
  }
}

/** React hook: re-renders the caller when the dragged node's relay word changes. Same
 *  two-stream subscription as useDraggedNodeName — the row comes from the Overlay block
 *  (VIEW stream), the column from the Node block (node stream). */
export function useDraggedNodeRelay(): string {
  return useSyncExternalStore(subscribeDraggedNodeName, readDraggedNodeRelay, readDraggedNodeRelay);
}

/** One current-drag recipient: its display name, its OWN cumulative
 *  DragRequantCount (Node block) — the per-recipient count that replaced the old
 *  central Overlay.AbcDragCount (that central counter lived on a cross-goroutine
 *  channel, abcDragCh, a fast drag's pointer-input load could starve, silently
 *  dropping ticks; DragRequantCount is state on each recipient's OWN reliable node
 *  stream, so nothing can be dropped) — plus the DRAGGED node's own quantized-triple
 *  delta (dA,dB,dC) that rode the message this recipient received (Node block
 *  DragDeltaA/B/C columns). */
export interface AbcDragRow {
  name: string;
  count: number;
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
      count: readNodeDragRequantCount(decoded.nodeView, row),
      dA: readNodeDragDeltaA(decoded.nodeView, row),
      dB: readNodeDragDeltaB(decoded.nodeView, row),
      dC: readNodeDragDeltaC(decoded.nodeView, row),
    });
  }
  const key = rows.map((r) => `${r.name}\0${r.count}\0${r.dA},${r.dB},${r.dC}`).join("\0");
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

/** One current-drag delta-FORWARD recipient: its display name, the FORWARDER's display
 *  name (resolved from ForwardFromRow, the buffer row of whichever neighbor's own hop
 *  reached this node FIRST this drag — see nodes/Wiring/node_mover.go's
 *  forwardDeltaOnce), and the forwarded delta triple (dA,dB,dC — the SAME triple that
 *  originated at the dragged node and is relayed unmodified). Every node forwards
 *  the delta it first picks up to its OTHER neighbors exactly once per drag
 *  (forwardedThisDrag), so this triple spreads across the whole reachable graph via
 *  independent concurrent one-hop relays, not just past the direct drag-recipient. */
export interface DeltaForwardRow {
  name: string;
  forwarderName: string;
  dA: number;
  dB: number;
  dC: number;
}

let cachedForwardRowsKey = "\0";
let cachedForwardRows: DeltaForwardRow[] = [];

/** Decode the current drag's delta-forward recipient ROWS (name + forwarder name +
 *  forwarded delta triple) from the Node block's per-row GotForwardMsg flag +
 *  ForwardDeltaA/B/C + ForwardFromRow columns. Go-owned and drag-scoped (cleared
 *  alongside GotDragMsg on KindAbcDragReset at drag start). Mirrors readAbcDragRows'
 *  shape/caching exactly. */
export function readDeltaForwardRows(): DeltaForwardRow[] {
  const decoded = getNodeFrame();
  if (!decoded) return cachedForwardRows;
  const rows: DeltaForwardRow[] = [];
  for (let row = 0; row < decoded.nodeCount; row++) {
    if (!readNodeGotForwardMsg(decoded.nodeView, row)) continue;
    const fromRow = readNodeForwardFromRow(decoded.nodeView, row);
    const forwarderName = fromRow >= 0 && fromRow < decoded.nodeCount ? nodeLabel(decoded, fromRow) : "";
    rows.push({
      name: nodeLabel(decoded, row),
      forwarderName,
      dA: readNodeForwardDeltaA(decoded.nodeView, row),
      dB: readNodeForwardDeltaB(decoded.nodeView, row),
      dC: readNodeForwardDeltaC(decoded.nodeView, row),
    });
  }
  const key = rows.map((r) => `${r.name}\0${r.forwarderName}\0${r.dA},${r.dB},${r.dC}`).join("\0");
  if (key === cachedForwardRowsKey) return cachedForwardRows;
  cachedForwardRowsKey = key;
  cachedForwardRows = rows;
  return cachedForwardRows;
}

/** React hook: re-renders the caller when the current drag's delta-forward recipient
 *  rows change — including when cleared to empty at drag start. Same Node-block
 *  subscription as useAbcDragRows. */
export function useDeltaForwardRows(): DeltaForwardRow[] {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readDeltaForwardRows, readDeltaForwardRows);
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

