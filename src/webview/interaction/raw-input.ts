import * as THREE from "three";
import { postGoRecord } from "../vscode-api";
import { encodeRawInput } from "../../Input/Codec/input-encode-scene-tilt";
import type { RawInputEvent, RawHit, RawPointerKind } from "../../Input/Codec/messages";
import type { PickRef } from "./pick-types";
import { pixelToNDC } from "./geometry-helpers";

type CamRef = React.MutableRefObject<THREE.PerspectiveCamera | null>;

export function sendRawInput(event: RawInputEvent): void {
  postGoRecord(encodeRawInput(event));
}

const wheelTotal = { x: 0, y: 0 };

function classifyHit(pickRequest: PickRef, ndcX: number, ndcY: number): { kind: RawHit["kind"]; isInput: boolean; nodeRow: number; portRow: number; edgeRow: number } {

  const edgeStr = pickRequest.current?.(ndcX, ndcY, { edgeOnly: true }) ?? null;
  if (edgeStr !== null) return { kind: "edge", isInput: false, nodeRow: -1, portRow: -1, edgeRow: Number(edgeStr) };

  const handholdStr = pickRequest.current?.(ndcX, ndcY, { handholdOnly: true }) ?? null;
  if (handholdStr !== null) return { kind: "handhold", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 };

  const torusStr = pickRequest.current?.(ndcX, ndcY, { ringOnly: true }) ?? null;
  if (torusStr !== null) return { kind: "torus", isInput: false, nodeRow: Number(torusStr), portRow: -1, edgeRow: -1 };

  const nodeStr = pickRequest.current?.(ndcX, ndcY) ?? null;
  if (nodeStr !== null) return { kind: "node", isInput: false, nodeRow: Number(nodeStr), portRow: -1, edgeRow: -1 };
  return { kind: "empty", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 };
}

export function buildPointerRaw(
  e: React.PointerEvent<HTMLDivElement>,
  kind: RawPointerKind,
  cameraRef: CamRef,
  pickRequest: PickRef,
): RawInputEvent | null {
  const cam = cameraRef.current;
  if (!cam) return null;
  const rect = e.currentTarget.getBoundingClientRect();
  const { ndcX, ndcY } = pixelToNDC(e.clientX, e.clientY, rect);
  const c = classifyHit(pickRequest, ndcX, ndcY);
  const hit: RawHit = { kind: c.kind, isInput: c.isInput, nodeRow: c.nodeRow, portRow: c.portRow, edgeRow: c.edgeRow };
  return {
    kind,
    x: e.clientX, y: e.clientY,
    rectLeft: rect.left, rectTop: rect.top, rectWidth: rect.width, rectHeight: rect.height,
    button: e.button,
    ctrl: e.ctrlKey, shift: e.shiftKey, alt: e.altKey, meta: e.metaKey,
    deltaX: 0, deltaY: 0,
    hit,
  };
}

export function buildHomeRaw(aspect: number): RawInputEvent {
  const hit: RawHit = { kind: "empty", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 };
  return {
    kind: "home",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: aspect, rectHeight: 1,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
  };
}

export function buildDeleteRaw(): RawInputEvent {
  const hit: RawHit = { kind: "empty", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 };
  return {
    kind: "delete",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: 0, rectHeight: 0,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
  };
}

export function buildKeyRaw(key: string): RawInputEvent {
  const hit: RawHit = { kind: "empty", isInput: false, nodeRow: -1, portRow: -1, edgeRow: -1 };
  return {
    kind: "key",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: 0, rectHeight: 0,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
    key,
  };
}

export function buildWheelRaw(
  e: WheelEvent,
  cameraRef: CamRef,
  pickRequest: PickRef,
): RawInputEvent | null {
  const cam = cameraRef.current;
  if (!cam) return null;
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  const { ndcX, ndcY } = pixelToNDC(e.clientX, e.clientY, rect);
  const c = classifyHit(pickRequest, ndcX, ndcY);
  const hit: RawHit = { kind: c.kind, isInput: c.isInput, nodeRow: c.nodeRow, portRow: c.portRow, edgeRow: c.edgeRow };
  wheelTotal.x += e.deltaX;
  wheelTotal.y += e.deltaY;
  return {
    kind: "wheel",
    x: e.clientX, y: e.clientY,
    rectLeft: rect.left, rectTop: rect.top, rectWidth: rect.width, rectHeight: rect.height,
    button: -1,
    ctrl: e.ctrlKey, shift: e.shiftKey, alt: e.altKey, meta: e.metaKey,
    deltaX: wheelTotal.x, deltaY: wheelTotal.y,
    hit,
  };
}
