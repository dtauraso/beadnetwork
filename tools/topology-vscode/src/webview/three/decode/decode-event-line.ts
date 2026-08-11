// decode-event-line.ts — per-EVENT-row decode: one buffer EVENT row -> one JSON-shaped
// Line, plus the overlay-flag lookup an "overlays-vis"-family kind resolves through.
// Split out of buffer-log.ts (see that file's doc comment for the surrounding contract):
// this is the pure per-row decode step; buffer-log.ts keeps the per-frame loop that calls
// it once per event and serializes the result to a .probe/go*.jsonl line.

import { TRACE_EVENT_KINDS, BREADCRUMB_LABELS } from "../../../schema/trace-kinds";
import { NODE_KIND_NAMES } from "../../../schema/node-defs";
import { nodeLabel, type DecodedNodeFrame } from "./buffer-decode-node";
import { edgeLabel, type DecodedEdgeFrame } from "./buffer-decode-edge";
import { INTERIOR_SLOTS_PER_NODE } from "./buffer-decode-interior";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSphereR,
  readNodeVRX, readNodeVRY, readNodeVRZ, readNodeFRX, readNodeFRY, readNodeFRZ,
  readNodeKindId,
  readInteriorPresent, readInteriorValue, readInteriorOX, readInteriorOY, readInteriorOZ,
  readEdgeSX, readEdgeSY, readEdgeSZ, readEdgeEX, readEdgeEY, readEdgeEZ,
  readCameraPX, readCameraPY, readCameraPZ, readCameraR,
  readCameraPosTheta, readCameraPosPhi, readCameraUpTheta, readCameraUpPhi,
  readOverlaySceneTori, readOverlayScenePoles, readOverlayNodePoles,
  readOverlaySelSpherePoles, readOverlayHandholds, readOverlayLabelsGlobal,
  readOverlayOverlaysVis,
  readOverlayNodeBody,
  readOverlayNodeRing,
  readOverlayRingPick,
  readOverlaySelectionRing,
  readOverlayHoverRing,
  readOverlayReachSphere,
  readEventKind, readEventNodeRow, readEventPortRow, readEventTargetRow, readEventTargetPortRow,
  readEventEdgeRow, readEventSlot, readEventValue, readEventBead,
  readEventBeadSteps, readEventSimLatencyMs, readEventX, readEventY, readEventZ, readEventF,
  readEventLabel, readEventDebug, readEventTextOff, readEventTextLen,
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
  UNKNOWN_KIND_ID,
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

function overlayFlag(vb: ViewBlocksOrNull, kind: string): number {
  const v = vb.overlayView;
  if (!v) return 0;
  switch (kind) {
    case "scene-tori": return readOverlaySceneTori(v);
    case "scene-poles": return readOverlayScenePoles(v);
    case "node-poles": return readOverlayNodePoles(v);
    case "sel-sphere-poles": return readOverlaySelSpherePoles(v);
    case "handholds": return readOverlayHandholds(v);
    case "labels-global": return readOverlayLabelsGlobal(v);
    case "overlays-vis": return readOverlayOverlaysVis(v);
    case "node-body": return readOverlayNodeBody(v);
    case "node-ring": return readOverlayNodeRing(v);
    case "ring-pick": return readOverlayRingPick(v);
    case "selection-ring": return readOverlaySelectionRing(v);
    case "hover-ring": return readOverlayHoverRing(v);
    case "reach-sphere": return readOverlayReachSphere(v);
    default: return 0;
  }
}

const OVERLAY_KINDS = new Set([
  "scene-tori", "scene-poles", "node-poles", "sel-sphere-poles",
  "handholds", "labels-global", "overlays-vis",
  "node-body", "node-ring", "ring-pick", "selection-ring", "hover-ring", "reach-sphere",
]);

function nodeGeometryLine(dn: DecodedNodeFrame, nodeRow: number, node: string): Line {
  // A node-geometry event riding the VIEW bucket resolves its node columns against the
  // last cached per-node stream frame, which can be a STALE generation with fewer rows than the
  // topology — reading nodeRow past nodeView would throw. Degrade to the label-only line
  // (same graceful-empty contract as nodeLabel), never crash the .probe logger.
  if (nodeRow < 0 || nodeRow >= dn.nodeCount) return { kind: "node-geometry", node };
  const n = dn.nodeView;
  const cx = readNodeCX(n, nodeRow), cy = readNodeCY(n, nodeRow), cz = readNodeCZ(n, nodeRow);
  const radius = readNodeRadius(n, nodeRow);
  const sphereR = readNodeSphereR(n, nodeRow);
  const kindId = readNodeKindId(n, nodeRow);
  // No `ports` array any more (docs/bead-model/channels-not-ports.md): a port carries no geometry,
  // so there is nothing to report per node beyond its own fields below.
  const l: Line = { kind: "node-geometry", node };
  if (node) l.label = node;
  if (kindId !== UNKNOWN_KIND_ID && NODE_KIND_NAMES[kindId] !== undefined) l.nodeKind = NODE_KIND_NAMES[kindId];
  l.nx = cx; l.ny = cy; l.nz = cz; l.radius = radius;
  if (sphereR !== 0) l.sphereR = sphereR;
  l.vrx = readNodeVRX(n, nodeRow); l.vry = readNodeVRY(n, nodeRow); l.vrz = readNodeVRZ(n, nodeRow);
  l.frx = readNodeFRX(n, nodeRow); l.fry = readNodeFRY(n, nodeRow); l.frz = readNodeFRZ(n, nodeRow);
  return l;
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
