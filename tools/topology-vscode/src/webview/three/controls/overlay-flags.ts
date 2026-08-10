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
//
// The REST of this Overlay-column read surface — drag row / dragged node name,
// edit-refused count, scene editable + scene kinds, selected node row, distance-group
// lens, playback speed, and per-node tilt-vector rows — lives in sibling files split by
// which buffer state each pair reflects: overlay-flags-drag.ts, overlay-flags-edit-
// refused.ts, overlay-flags-scene.ts, overlay-flags-selection.ts, overlay-flags-
// distance-groups.ts, overlay-flags-speed.ts, overlay-flags-tilt-vectors.ts. Every one
// of them is the same shape as this file (read + use + equality-helper trio, read-only
// reflect of the buffer) and is covered by the same check-no-webview-state.sh allowlist.

import { useSyncExternalStore } from "react";
import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "../../../messages";
import { getViewBlocks, subscribeViewBlocks } from "../scene/view-blocks";
import {
  readOverlaySceneTori,
  readOverlayScenePoles,
  readOverlayNodePoles,
  readOverlaySelSpherePoles,
  readOverlayHandholds,
  readOverlayLabelsGlobal,
  readOverlayOverlaysVis,
  readOverlayNodeBody,
  readOverlayNodeRing,
  readOverlayRingPick,
  readOverlaySelectionRing,
  readOverlayHoverRing,
  readOverlayReachSphere,
} from "../../../schema/buffer-layout";

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

// Compared over OVERLAY_FLAG_ORDER — the flag vocabulary itself — NOT a hand-written
// conjunction of field names. The conjunction was a second list of the flags that had to be
// extended in step with the first, and when six flags were added and it was not, this
// function reported "unchanged" for every one of them: the cached object kept its identity,
// useSyncExternalStore saw nothing, and the row's checkmark froze while the drawing itself
// (which reads the buffer directly, not this) turned off correctly. Iterating the vocabulary
// means a new flag is covered by existing.
function overlayFlagsEqual(a: OverlayFlagVals, b: OverlayFlagVals): boolean {
  return OVERLAY_FLAG_ORDER.every((flag) => a[flag] === b[flag]);
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
    nodeBody: !!readOverlayNodeBody(v),
    nodeRing: !!readOverlayNodeRing(v),
    ringPick: !!readOverlayRingPick(v),
    selectionRing: !!readOverlaySelectionRing(v),
    hoverRing: !!readOverlayHoverRing(v),
    reachSphere: !!readOverlayReachSphere(v),
  };
  if (cachedVals && overlayFlagsEqual(cachedVals, next)) return cachedVals;
  cachedVals = next;
  return cachedVals;
}

/** Read ONE overlay column from the latest snapshot, for a renderer inside a useFrame that
 *  cannot use the hook (it is not a React render) and wants the flag from the same frame as
 *  the geometry it is deciding about. `read` is the generated column reader.
 *
 *  Missing snapshot ⇒ FALSE, not true: before the first frame there is nothing to draw
 *  anyway, and "draw it until told otherwise" is how a decoration outlives the flag that
 *  turned it off. */
export function overlayOn(read: (v: DataView) => number): boolean {
  const blocks = getViewBlocks();
  if (!blocks) return false;
  return read(blocks.overlayView) !== 0;
}

/** React hook: re-renders the caller when any overlay flag flips (Go-owned). Returns
 *  null until the first snapshot lands. */
export function useOverlayFlags(): OverlayFlagVals | null {
  return useSyncExternalStore(subscribeViewBlocks, readOverlayFlags, readOverlayFlags);
}

