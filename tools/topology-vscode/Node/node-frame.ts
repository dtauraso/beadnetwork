import {
  readNodeIndexR, readNodeIndexPhi, readNodeIndexTheta, readNodeHasPos,
  readNodeRadius, readNodeSelected,
} from "../Buffer/buffer-layout";
import { sceneSteps } from "../Scene/scene-frame";
import { NODE_SPHERE_RADIUS } from "../src/webview/three/scene/buffer-scene-shared";
import { getViewBlocks } from "../src/webview/three/scene/view-blocks";

function sceneStep() {
  return sceneSteps();
}

function centerInto(nodeView: DataView, row: number, out: [number, number, number]): void {
  const s = sceneStep();
  if (!s || !readNodeHasPos(nodeView, row)) {
    out[0] = 0; out[1] = 0; out[2] = 0;
    return;
  }
  const r = readNodeIndexR(nodeView, row) * s.constantR;
  const phi = readNodeIndexPhi(nodeView, row) * s.constantPhi;
  const theta = readNodeIndexTheta(nodeView, row) * s.constantTheta;
  const st = Math.sin(phi);
  out[0] = s.centerX + r * st * Math.cos(theta);
  out[1] = s.centerY + r * Math.cos(phi);
  out[2] = s.centerZ + r * st * Math.sin(theta);
}

const scratch: [number, number, number] = [0, 0, 0];

export function nodeCenterX(nodeView: DataView, row: number): number {
  centerInto(nodeView, row, scratch);
  return scratch[0];
}

export function nodeCenterY(nodeView: DataView, row: number): number {
  centerInto(nodeView, row, scratch);
  return scratch[1];
}

export function nodeCenterZ(nodeView: DataView, row: number): number {
  centerInto(nodeView, row, scratch);
  return scratch[2];
}

export function nodeRadius(nodeView: DataView, row: number): number {
  return readNodeRadius(nodeView, row) || NODE_SPHERE_RADIUS;
}

export function nodeRadiusRaw(nodeView: DataView, row: number): number {
  return readNodeRadius(nodeView, row);
}

export function nodeSelected(nodeView: DataView, row: number): boolean {
  return readNodeSelected(nodeView, row) !== 0;
}
