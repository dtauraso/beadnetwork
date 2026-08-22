import * as THREE from "three";
import { ownerCounts } from "../../Scene/owner-counts";
import { nodeF32, nodeI32, nodeU8, NODE_RING_NAMES } from "../../Node/node-leaves";
import { sceneSteps } from "../../Scene/scene-frame";
import { NODE_SPHERE_RADIUS } from "../../webview/scene/scene-tags";
import { readSelectedNodeRow } from "../../webview/flags/overlay-flags-selection";

const centerScratch: [number, number, number] = [0, 0, 0];

function centerInto(row: number, out: [number, number, number]): void {
  if (!nodeU8(row, "hasPos")) {
    out[0] = 0; out[1] = 0; out[2] = 0;
    return;
  }
  const s = sceneSteps();
  const r = nodeI32(row, "indexR") * s.constantR;
  const phi = nodeI32(row, "indexPhi") * s.constantPhi;
  const theta = nodeI32(row, "indexTheta") * s.constantTheta;
  const st = Math.sin(phi);
  out[0] = s.centerX + r * st * Math.cos(theta);
  out[1] = s.centerY + r * Math.cos(phi);
  out[2] = s.centerZ + r * st * Math.sin(theta);
}

function nodeCenterX(row: number): number { centerInto(row, centerScratch); return centerScratch[0]; }
function nodeCenterY(row: number): number { centerInto(row, centerScratch); return centerScratch[1]; }
function nodeCenterZ(row: number): number { centerInto(row, centerScratch); return centerScratch[2]; }

function nodeRadius(row: number): number {
  return nodeF32(row, "radius") || NODE_SPHERE_RADIUS;
}

function nodeCount(): number { return ownerCounts().nodes; }

const hoveredFlag = (row: number): boolean => nodeU8(row, "hovered") !== 0;
import { nodeRowColors } from "../../Node/node-kind";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";
import { SELECTION_HALO_R_RATIO } from "./node-highlight-shape";
import { overlayFlag } from "../../webview/flags/overlay-flags";

function copyRingMatrix(row: number, ring: THREE.InstancedMesh, slot: number): void {
  const out = ring.instanceMatrix.array;
  const base = slot * 16;
  for (let i = 0; i < 16; i++) out[base + i] = nodeF32(row, NODE_RING_NAMES[i]!);
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
