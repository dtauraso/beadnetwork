import * as THREE from "three";
import { columnF32, columnI32, columnU8, columnBytes } from "../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../Buffer/column-owners";
import {
  COL_STREAM_NODE_INDEX_R, COL_STREAM_NODE_INDEX_PHI, COL_STREAM_NODE_INDEX_THETA,
  COL_STREAM_NODE_HAS_POS, COL_STREAM_NODE_RADIUS, COL_STREAM_NODE_HOVERED,
  COL_STREAM_NODE_RING_MATRIX,
} from "../../Buffer/column-streams-gen";
import { sceneSteps } from "../../Scene/scene-frame";
import { NODE_SPHERE_RADIUS } from "../../src/webview/three/scene/buffer-scene-shared";
import { readSelectedNodeRow } from "../../src/webview/three/controls/flags/overlay-flags-selection";

const centerScratch: [number, number, number] = [0, 0, 0];

function centerInto(row: number, out: [number, number, number]): void {
  if (!columnU8(nodeColumn(row, COL_STREAM_NODE_HAS_POS))) {
    out[0] = 0; out[1] = 0; out[2] = 0;
    return;
  }
  const s = sceneSteps();
  const r = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_R)) * s.constantR;
  const phi = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_PHI)) * s.constantPhi;
  const theta = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_THETA)) * s.constantTheta;
  const st = Math.sin(phi);
  out[0] = s.centerX + r * st * Math.cos(theta);
  out[1] = s.centerY + r * Math.cos(phi);
  out[2] = s.centerZ + r * st * Math.sin(theta);
}

function nodeCenterX(row: number): number { centerInto(row, centerScratch); return centerScratch[0]; }
function nodeCenterY(row: number): number { centerInto(row, centerScratch); return centerScratch[1]; }
function nodeCenterZ(row: number): number { centerInto(row, centerScratch); return centerScratch[2]; }

function nodeRadius(row: number): number {
  return columnF32(nodeColumn(row, COL_STREAM_NODE_RADIUS)) || NODE_SPHERE_RADIUS;
}

function nodeCount(): number { return ownerCounts().nodes; }

const hoveredFlag = (row: number): boolean => columnU8(nodeColumn(row, COL_STREAM_NODE_HOVERED)) !== 0;
import { nodeRowColors } from "../node-kind";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";
import { SELECTION_HALO_R_RATIO } from "./node-highlight-shape";
import { overlayFlag } from "../../src/webview/three/controls/flags/overlay-flags";

function copyRingMatrix(row: number, ring: THREE.InstancedMesh, slot: number): void {
  const out = ring.instanceMatrix.array;
  const base = slot * 16;
  const m = columnBytes(nodeColumn(row, COL_STREAM_NODE_RING_MATRIX));
  for (let i = 0; i < 16; i++) {
    out[base + i] = m && m.byteLength >= 64 ? m.getFloat32(i * 4, true) : 0;
  }
}

export interface NodeInstanceRefs {
  body: THREE.InstancedMesh;
  ring: THREE.InstancedMesh;
  ringPick: THREE.InstancedMesh;
  ringBand: THREE.InstancedMesh;
  selRing: THREE.InstancedMesh;
  selHalo: THREE.InstancedMesh;
  hoverRing: THREE.InstancedMesh;
  mat: THREE.Matrix4;
  pos: THREE.Vector3;
  quat: THREE.Quaternion;
  scl: THREE.Vector3;
  col: THREE.Color;
}

export function updateNodeInstances(refs: NodeInstanceRefs, capacity: number, camera: THREE.Camera): void {
  const { body, ring, ringPick, ringBand, selRing, selHalo, hoverRing, mat, pos, quat, scl, col } = refs;

  const n0 = nodeCount();
  if (n0 <= 0) {
    body.count = 0; ring.count = 0; ringPick.count = 0; ringBand.count = 0;
    selRing.count = 0; selHalo.count = 0; hoverRing.count = 0;
    return;
  }

  const showBody = overlayFlag("nodeBody");
  const showRing = overlayFlag("nodeRing");

  const showPickBand = overlayFlag("ringPick");

  const n = Math.min(n0, capacity);

  const order = computeNodeDepthOrder(
    n,
    (row) => nodeCenterX(row),
    (row) => nodeCenterY(row),
    (row) => nodeCenterZ(row),
    camera.position.x, camera.position.y, camera.position.z,
  );
  setNodeDrawOrder(order);
  for (let slot = 0; slot < n; slot++) {
    const row = order[slot]!;
    const r = nodeRadius(row);
    pos.set(
      nodeCenterX(row),
      nodeCenterY(row),
      nodeCenterZ(row),
    );

    scl.setScalar(r);
    mat.compose(pos, quat, scl);
    body.setMatrixAt(slot, mat);

    copyRingMatrix(row, ring, slot);
    copyRingMatrix(row, ringPick, slot);
    copyRingMatrix(row, ringBand, slot);

    const { fill, stroke } = nodeRowColors(row);
    body.setColorAt(slot, col.set(fill));
    ring.setColorAt(slot, col.set(stroke));
  }
  body.count = showBody ? n : 0;
  ring.count = showRing ? n : 0;

  ringPick.count = n;

  ringBand.count = showPickBand ? n : 0;
  body.instanceMatrix.needsUpdate = true;
  ring.instanceMatrix.needsUpdate = true;
  ringPick.instanceMatrix.needsUpdate = true;
  ringBand.instanceMatrix.needsUpdate = true;
  if (body.instanceColor) body.instanceColor.needsUpdate = true;
  if (ring.instanceColor) ring.instanceColor.needsUpdate = true;

  placeHighlight(n, selectedFlag, selRing, overlayFlag("selectionRing"), 1, mat, pos, quat, scl);
  placeHighlight(n, selectedFlag, selHalo, overlayFlag("selectionRing"), SELECTION_HALO_R_RATIO, mat, pos, quat, scl);

  const hoveredRow = firstRowWhere(n, hoveredFlag);
  const hoverSuppressed =
    hoveredRow >= 0 && readSelectedNodeRow() === hoveredRow && overlayFlag("selectionRing");
  placeHighlight(n, hoveredFlag, hoverRing, overlayFlag("hoverRing") && !hoverSuppressed, 1, mat, pos, quat, scl);

  if (showBody) body.computeBoundingSphere();
  ringPick.computeBoundingSphere();
}

const selectedFlag = (row: number): boolean => readSelectedNodeRow() === row;

function firstRowWhere(n: number, read: (row: number) => boolean): number {
  for (let i = 0; i < n; i++) {
    if (read(i)) return i;
  }
  return -1;
}

function placeHighlight(
  n: number,
  read: (row: number) => boolean,
  mesh: THREE.InstancedMesh, visible: boolean, rRatio: number,
  mat: THREE.Matrix4, pos: THREE.Vector3, quat: THREE.Quaternion, scl: THREE.Vector3,
): void {
  const row = visible ? firstRowWhere(n, read) : -1;
  if (row < 0) {
    mesh.count = 0;
    return;
  }
  const r = nodeRadius(row) * rRatio;
  pos.set(nodeCenterX(row), nodeCenterY(row), nodeCenterZ(row));
  scl.setScalar(r);
  mat.compose(pos, quat, scl);
  mesh.setMatrixAt(0, mat);
  mesh.count = 1;
  mesh.instanceMatrix.needsUpdate = true;
}
