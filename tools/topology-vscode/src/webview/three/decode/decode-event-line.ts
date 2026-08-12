// decode-event-line.ts — per-EVENT-row decode: one buffer EVENT row -> one JSON-shaped
// Line, plus the overlay-flag lookup an "overlays-vis"-family kind resolves through.
// Split out of buffer-log.ts (see that file's doc comment for the surrounding contract):
// this is the pure per-row decode step; buffer-log.ts keeps the per-frame loop that calls
// it once per event and serializes the result to a .probe/go*.jsonl line.

import { TRACE_EVENT_KINDS, BREADCRUMB_LABELS } from "../../../schema/trace-kinds";
import { nodeLabel, type DecodedNodeFrame } from "./buffer-decode-node";
import { edgeLabel, type DecodedEdgeFrame } from "./buffer-decode-edge";
import { INTERIOR_SLOTS_PER_NODE } from "./buffer-decode-interior";
import { overlayFlag, OVERLAY_KINDS } from "./decode-event-overlay";
import { nodeGeometryLine } from "./decode-event-node-geometry";
import {
  readInteriorPresent, readInteriorValue, readInteriorOX, readInteriorOY, readInteriorOZ,
  readEdgeSX, readEdgeSY, readEdgeSZ, readEdgeEX, readEdgeEY, readEdgeEZ,
  readCameraPX, readCameraPY, readCameraPZ, readCameraR,
  readCameraPosTheta, readCameraPosPhi, readCameraUpTheta, readCameraUpPhi,
  readEventKind, readEventNodeRow, readEventPortRow, readEventTargetRow, readEventTargetPortRow,
  readEventEdgeRow, readEventSlot, readEventValue, readEventBead,
  readEventBeadSteps, readEventSimLatencyMs, readEventX, readEventY, readEventZ, readEventF,
  readEventLabel, readEventDebug, readEventTextOff, readEventTextLen,
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
} from "../../../schema/buffer-layout";

export type Line = Record<string, unknown>;

/** Shared UTF-8 decoder for the event-text-bytes section (the sanctioned single
 *  free-form string escape hatch on the EVENT row — see Buffer.BuildEventsSection). */
const EVENT_TEXT_DECODER = new TextDecoder();

/** camera/overlay/scene views resolved from EITHER source — the SCENE frame's embedded
 *  blocks (fallback, no dedicated view fd) OR the dedicated VIEW frame (see
 *  webview/three/scene/view-blocks.ts's ext-host-side mirror). Null fields mean neither
 *  source has landed for that block yet. */
export interface ViewBlocksOrNull {
  cameraView: DataView | null;
  overlayView: DataView | null;
  sceneView: DataView | null;
}

/** Decode ONE EVENT row (index i) into its JSON-shaped Line, or null when the row's kind
 *  id is unknown (an out-of-range/never-emitted kind byte). dn/de resolve node/edge
 *  identity strings when available (best-effort — see each caller's doc comment); vb
 *  resolves camera/overlay/scene-sphere fields when this row's kind needs them. */
export function decodeEventLine(ev: DataView, eventTextView: DataView, dn: DecodedNodeFrame | null, de: DecodedEdgeFrame | null, vb: ViewBlocksOrNull, i: number): Line | null {
  const kindId = readEventKind(ev, i);
  const kind = TRACE_EVENT_KINDS[kindId];
  if (kind === undefined) return null;
  const nodeRow = readEventNodeRow(ev, i);
  const portRow = readEventPortRow(ev, i);
  const targetRow = readEventTargetRow(ev, i);
  const targetPortRow = readEventTargetPortRow(ev, i);
  const edgeRow = readEventEdgeRow(ev, i);
  const value = readEventValue(ev, i);
  const bead = readEventBead(ev, i);
  const node = dn && nodeRow >= 0 ? nodeLabel(dn, nodeRow) : "";
  // port is always "" now: a port has no name/row on the buffer any more
  // (docs/bead-model/channels-not-ports.md). portRow still rides the event as a sentinel (-1).
  const port = "";

  if (kind === "breadcrumb") {
    // DEBUG BREADCRUMB channel (task/breadcrumbs-binary-buffer): a structured EVENT
    // row instead of the retired free-form JSON stdout line. Column meanings are
    // REUSED per breadcrumb Label (see each Go call site's own comment) — this decode
    // exposes every raw column plus the resolved label name and free-form text (if
    // any), so probe-merge.sh --debug (filtering on debug===true) can display whatever
    // a given label populated without this decoder needing per-label knowledge.
    const labelId = readEventLabel(ev, i);
    const label = BREADCRUMB_LABELS[labelId] ?? String(labelId);
    const textOff = readEventTextOff(ev, i);
    const textLen = readEventTextLen(ev, i);
    const text = textLen > 0 && eventTextView.byteLength >= textOff + textLen
      ? EVENT_TEXT_DECODER.decode(new Uint8Array(eventTextView.buffer, eventTextView.byteOffset + textOff, textLen))
      : "";
    const t = dn && targetRow >= 0 ? nodeLabel(dn, targetRow) : "";
    const line: Line = {
      kind, label, debug: readEventDebug(ev, i) === 1,
      node, port, value,
      x: readEventX(ev, i), y: readEventY(ev, i), z: readEventZ(ev, i),
      nodeRow, portRow, targetRow, targetPortRow, edgeRow, slot: readEventSlot(ev, i),
    };
    if (t) line.target = t;
    if (text) line.text = text;
    return line;
  }

  switch (kind) {
    case "recv":
      return { kind, node, port, value };
    case "fire":
      return { kind, node };
    case "send": {
      const beadSteps = readEventBeadSteps(ev, i);
      const lat = readEventSimLatencyMs(ev, i);
      if (beadSteps !== 0 || lat !== 0) {
        const l: Line = { kind, node, port, value, beadSteps, simLatencyMs: lat };
        const t = dn && targetRow >= 0 ? nodeLabel(dn, targetRow) : "";
        if (t) l.target = t;
        // No targetHandle any more: a port has no name on the buffer (docs/bead-model/channels-not-ports.md).
        return l;
      }
      return { kind, node, port, value };
    }
    case "edge-bead": {
      const l: Line = { kind, node, port, value, x: readEventX(ev, i), y: readEventY(ev, i), z: readEventZ(ev, i), f: readEventF(ev, i) };
      if (bead !== 0) l.bead = bead;
      return l;
    }
    case "arrive": {
      const l: Line = { kind, node, port, value };
      if (bead !== 0) l.bead = bead;
      return l;
    }
    case "geometry": {
      const edge = de ? edgeLabel(de, edgeRow) : "";
      // The Edge block carries its own SEGMENT (SX..EZ) directly — node surface to node
      // surface (docs/bead-model/channels-not-ports.md), not a reference through a port row.
      let sx = 0, sy = 0, sz = 0, ex = 0, ey = 0, ez = 0;
      if (de && edgeRow >= 0 && edgeRow < de.edgeCount) {
        sx = readEdgeSX(de.edgeView, edgeRow); sy = readEdgeSY(de.edgeView, edgeRow); sz = readEdgeSZ(de.edgeView, edgeRow);
        ex = readEdgeEX(de.edgeView, edgeRow); ey = readEdgeEY(de.edgeView, edgeRow); ez = readEdgeEZ(de.edgeView, edgeRow);
      }
      return { kind, edge, sx, sy, sz, ex, ey, ez };
    }
    case "node-geometry":
      return dn ? nodeGeometryLine(dn, nodeRow, node) : { kind, node };
    case "node-bead": {
      if (!dn) return { kind, node };
      const slot = readEventSlot(ev, i);
      const irow = nodeRow * INTERIOR_SLOTS_PER_NODE + slot;
      return {
        kind, node, row: Math.floor(slot / 2), col: slot % 2,
        present: readInteriorPresent(dn.interiorView, irow) === 1,
        value: readInteriorValue(dn.interiorView, irow),
        x: readInteriorOX(dn.interiorView, irow), y: readInteriorOY(dn.interiorView, irow), z: readInteriorOZ(dn.interiorView, irow),
      };
    }
    case "camera": {
      const c = vb.cameraView;
      if (!c) return { kind };
      return {
        kind,
        px: readCameraPX(c), py: readCameraPY(c), pz: readCameraPZ(c), r: readCameraR(c),
        posTheta: readCameraPosTheta(c), posPhi: readCameraPosPhi(c),
        upTheta: readCameraUpTheta(c), upPhi: readCameraUpPhi(c),
      };
    }
    case "scene-sphere": {
      const sc = vb.sceneView;
      if (!sc) return { kind };
      return { kind, cx: readSceneCX(sc), cy: readSceneCY(sc), cz: readSceneCZ(sc), radius: readSceneRadius(sc) };
    }
    case "select":
      // stdout marshals select via the default {node,port,value} shape (edge label not emitted).
      return { kind, node, port: "", value };
    case "hover":
      return { kind, node, port, value };
    default:
      if (OVERLAY_KINDS.has(kind)) return { kind, visible: overlayFlag(vb, kind) === 1 };
      return { kind, node, port, value };
  }
}
