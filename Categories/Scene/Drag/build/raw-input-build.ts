import * as THREE from "three";
import type { RawInputEvent, RawHit, RawPointerKind } from "./raw-input";
import type { PickRef } from "./pick-types";
import { pixelToNDC } from "./ndc";
import { planePoint } from "./plane-point";
import { ballPoint } from "./ball-point";

type CamRef = React.MutableRefObject<THREE.PerspectiveCamera | null>;

const wheelTotal = { x: 0, y: 0 };

let pressOnRim = false;

let pressNdc: { x: number; y: number } | null = null;

let pressCam: THREE.PerspectiveCamera | null = null;

function freezeCam(cam: THREE.PerspectiveCamera): THREE.PerspectiveCamera {
  const c = cam.clone();
  c.matrixWorld.copy(cam.matrixWorld);
  c.projectionMatrixInverse.copy(cam.projectionMatrixInverse);
  return c;
}

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
  const p = planePoint(cam, ndcX, ndcY, c.nodeRow);
  if (kind === "pointerdown") {
    pressOnRim = c.kind === "handhold" || ballPoint(cam, ndcX, ndcY, false).onRim;
    pressNdc = { x: ndcX, y: ndcY };
    pressCam = freezeCam(cam);
  }
  const ballCam = pressCam ?? cam;
  const ball = ballPoint(ballCam, ndcX, ndcY, pressOnRim);
  const prev = ballPoint(ballCam, pressNdc?.x ?? ndcX, pressNdc?.y ?? ndcY, pressOnRim);
  const hit: RawHit = {
    kind: c.kind, isInput: c.isInput, onRim: pressOnRim,
    nodeRow: c.nodeRow, portRow: c.portRow, edgeRow: c.edgeRow,
    pointX: p.x, pointY: p.y, pointZ: p.z,
  };
  return {
    kind,
    x: e.clientX, y: e.clientY,
    rectLeft: rect.left, rectTop: rect.top, rectWidth: rect.width, rectHeight: rect.height,
    button: e.button,
    ctrl: e.ctrlKey, shift: e.shiftKey, alt: e.altKey, meta: e.metaKey,
    deltaX: 0, deltaY: 0,
    hit,
    ballX: ball.x, ballY: ball.y, ballZ: ball.z,
    ballPrevX: prev.x, ballPrevY: prev.y, ballPrevZ: prev.z,
  };
}

export function buildHomeRaw(aspect: number): RawInputEvent {
  const ball = { x: 0, y: 0, z: 0 };
  const hit: RawHit = {
    kind: "empty", isInput: false, onRim: false, nodeRow: -1, portRow: -1, edgeRow: -1,
    pointX: 0, pointY: 0, pointZ: 0,
  };
  return {
    kind: "home",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: aspect, rectHeight: 1,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
    ballX: ball.x, ballY: ball.y, ballZ: ball.z,
    ballPrevX: ball.x, ballPrevY: ball.y, ballPrevZ: ball.z,
  };
}

export function buildDeleteRaw(): RawInputEvent {
  const ball = { x: 0, y: 0, z: 0 };
  const hit: RawHit = {
    kind: "empty", isInput: false, onRim: false, nodeRow: -1, portRow: -1, edgeRow: -1,
    pointX: 0, pointY: 0, pointZ: 0,
  };
  return {
    kind: "delete",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: 0, rectHeight: 0,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
    ballX: ball.x, ballY: ball.y, ballZ: ball.z,
    ballPrevX: ball.x, ballPrevY: ball.y, ballPrevZ: ball.z,
  };
}

export function buildKeyRaw(key: string): RawInputEvent {
  const ball = { x: 0, y: 0, z: 0 };
  const hit: RawHit = {
    kind: "empty", isInput: false, onRim: false, nodeRow: -1, portRow: -1, edgeRow: -1,
    pointX: 0, pointY: 0, pointZ: 0,
  };
  return {
    kind: "key",
    x: 0, y: 0,
    rectLeft: 0, rectTop: 0, rectWidth: 0, rectHeight: 0,
    button: -1,
    ctrl: false, shift: false, alt: false, meta: false,
    deltaX: 0, deltaY: 0,
    hit,
    key,
    ballX: ball.x, ballY: ball.y, ballZ: ball.z,
    ballPrevX: ball.x, ballPrevY: ball.y, ballPrevZ: ball.z,
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
  const p = planePoint(cam, ndcX, ndcY, c.nodeRow);
  const ball = ballPoint(cam, ndcX, ndcY);
  const hit: RawHit = {
    kind: c.kind, isInput: c.isInput, onRim: false,
    nodeRow: c.nodeRow, portRow: c.portRow, edgeRow: c.edgeRow,
    pointX: p.x, pointY: p.y, pointZ: p.z,
  };
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
    ballX: ball.x, ballY: ball.y, ballZ: ball.z,
    ballPrevX: ball.x, ballPrevY: ball.y, ballPrevZ: ball.z,
  };
}
