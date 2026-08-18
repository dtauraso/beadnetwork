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
  readCameraPosPhi, readCameraPosTheta, readCameraUpPhi, readCameraUpTheta,
  readEventKind, readEventNodeRow, readEventPortRow, readEventTargetRow, readEventTargetPortRow,
  readEventEdgeRow, readEventSlot, readEventValue, readEventBead,
  readEventBeadSteps, readEventX, readEventY, readEventZ, readEventF,
  readEventLabel, readEventDebug, readEventTextOff, readEventTextLen,
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
} from "../../../schema/buffer-layout/buffer-layout";

export type Line = Record<string, unknown>;

const EVENT_TEXT_DECODER = new TextDecoder();

export interface ViewBlocksOrNull {
  cameraView: DataView | null;
  overlayView: DataView | null;
  sceneView: DataView | null;
}

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

  const port = "";

  if (kind === "breadcrumb") {

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
      if (beadSteps !== 0) {
        const l: Line = { kind, node, port, value, beadSteps };
        const t = dn && targetRow >= 0 ? nodeLabel(dn, targetRow) : "";
        if (t) l.target = t;

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
        posTheta: readCameraPosPhi(c), posPhi: readCameraPosTheta(c),
        upTheta: readCameraUpPhi(c), upPhi: readCameraUpTheta(c),
      };
    }
    case "scene-sphere": {
      const sc = vb.sceneView;
      if (!sc) return { kind };
      return { kind, cx: readSceneCX(sc), cy: readSceneCY(sc), cz: readSceneCZ(sc), radius: readSceneRadius(sc) };
    }
    case "select":

      return { kind, node, port: "", value };
    case "hover":
      return { kind, node, port, value };
    default:
      if (OVERLAY_KINDS.has(kind)) return { kind, visible: overlayFlag(vb, kind) === 1 };
      return { kind, node, port, value };
  }
}
