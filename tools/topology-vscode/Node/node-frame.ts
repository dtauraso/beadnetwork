import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSelected,
} from "../Buffer/buffer-layout";
import { NODE_SPHERE_RADIUS } from "../src/webview/three/scene/buffer-scene-shared";

export function nodeCenterX(nodeView: DataView, row: number): number {
  return readNodeCX(nodeView, row);
}

export function nodeCenterY(nodeView: DataView, row: number): number {
  return readNodeCY(nodeView, row);
}

export function nodeCenterZ(nodeView: DataView, row: number): number {
  return readNodeCZ(nodeView, row);
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
