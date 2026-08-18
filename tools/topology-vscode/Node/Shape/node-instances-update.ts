import * as THREE from "three";
import { getNodeFrame } from "../../src/webview/three/scene/nodes/node-frame-aggregate";
import { getViewBlocks } from "../../src/webview/three/scene/view-blocks";
import {
  readNodeHovered,
  readNodeRingM0, readNodeRingM1, readNodeRingM2, readNodeRingM3,
  readNodeRingM4, readNodeRingM5, readNodeRingM6, readNodeRingM7,
  readNodeRingM8, readNodeRingM9, readNodeRingM10, readNodeRingM11,
  readNodeRingM12, readNodeRingM13, readNodeRingM14, readNodeRingM15,
} from "../../Buffer/buffer-layout";
import { nodeCenterX, nodeCenterY, nodeCenterZ, nodeRadius, nodeSelected } from "../node-frame";
import { nodeRowColors } from "../node-kind";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";
import { SELECTION_HALO_R_RATIO } from "./node-highlight-shape";
import { overlayFlag } from "../../src/webview/three/controls/flags/overlay-flags";

function copyRingMatrix(nodeView: DataView, row: number, ring: THREE.InstancedMesh, slot: number): void {
  const out = ring.instanceMatrix.array;
  const base = slot * 16;
  out[base]      = readNodeRingM0(nodeView, row);
  out[base + 1]  = readNodeRingM1(nodeView, row);
  out[base + 2]  = readNodeRingM2(nodeView, row);
  out[base + 3]  = readNodeRingM3(nodeView, row);
  out[base + 4]  = readNodeRingM4(nodeView, row);
  out[base + 5]  = readNodeRingM5(nodeView, row);
  out[base + 6]  = readNodeRingM6(nodeView, row);
  out[base + 7]  = readNodeRingM7(nodeView, row);
  out[base + 8]  = readNodeRingM8(nodeView, row);
  out[base + 9]  = readNodeRingM9(nodeView, row);
  out[base + 10] = readNodeRingM10(nodeView, row);
  out[base + 11] = readNodeRingM11(nodeView, row);
  out[base + 12] = readNodeRingM12(nodeView, row);
  out[base + 13] = readNodeRingM13(nodeView, row);
  out[base + 14] = readNodeRingM14(nodeView, row);
  out[base + 15] = readNodeRingM15(nodeView, row);
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

  const blocks = getViewBlocks();
  const decodedNode = getNodeFrame();
  if (!decodedNode || !blocks) {
    body.count = 0; ring.count = 0; ringPick.count = 0; ringBand.count = 0;
    selRing.count = 0; selHalo.count = 0; hoverRing.count = 0;
    return;
  }
  const { nodeCount, nodeView } = decodedNode;

  const showBody = overlayFlag("nodeBody");
  const showRing = overlayFlag("nodeRing");

  const showPickBand = overlayFlag("ringPick");

  const n = Math.min(nodeCount, capacity);

  const order = computeNodeDepthOrder(
    n,
    (row) => nodeCenterX(nodeView, row),
    (row) => nodeCenterY(nodeView, row),
    (row) => nodeCenterZ(nodeView, row),
    camera.position.x, camera.position.y, camera.position.z,
  );
  setNodeDrawOrder(order);
  for (let slot = 0; slot < n; slot++) {
    const row = order[slot]!;
    const r = nodeRadius(nodeView, row);
    pos.set(
      nodeCenterX(nodeView, row),
      nodeCenterY(nodeView, row),
      nodeCenterZ(nodeView, row),
    );

    scl.setScalar(r);
    mat.compose(pos, quat, scl);
    body.setMatrixAt(slot, mat);

    copyRingMatrix(nodeView, row, ring, slot);
    copyRingMatrix(nodeView, row, ringPick, slot);
    copyRingMatrix(nodeView, row, ringBand, slot);

    const { fill, stroke } = nodeRowColors(nodeView, row);
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

  placeHighlight(nodeView, n, selectedFlag, selRing, overlayFlag("selectionRing"), 1, mat, pos, quat, scl);
  placeHighlight(nodeView, n, selectedFlag, selHalo, overlayFlag("selectionRing"), SELECTION_HALO_R_RATIO, mat, pos, quat, scl);

  const hoveredRow = firstRowWhere(nodeView, n, readNodeHovered);
  const hoverSuppressed =
    hoveredRow >= 0 && nodeSelected(nodeView, hoveredRow) && overlayFlag("selectionRing");
  placeHighlight(nodeView, n, readNodeHovered, hoverRing, overlayFlag("hoverRing") && !hoverSuppressed, 1, mat, pos, quat, scl);

  if (showBody) body.computeBoundingSphere();
  ringPick.computeBoundingSphere();
}

const selectedFlag = (v: DataView, row: number): number => (nodeSelected(v, row) ? 1 : 0);

function firstRowWhere(
  nodeView: DataView, n: number, read: (v: DataView, row: number) => number,
): number {
  for (let i = 0; i < n; i++) {
    if (read(nodeView, i)) return i;
  }
  return -1;
}

function placeHighlight(
  nodeView: DataView, n: number,
  read: (v: DataView, row: number) => number,
  mesh: THREE.InstancedMesh, visible: boolean, rRatio: number,
  mat: THREE.Matrix4, pos: THREE.Vector3, quat: THREE.Quaternion, scl: THREE.Vector3,
): void {
  const row = visible ? firstRowWhere(nodeView, n, read) : -1;
  if (row < 0) {
    mesh.count = 0;
    return;
  }
  const r = nodeRadius(nodeView, row) * rRatio;
  pos.set(nodeCenterX(nodeView, row), nodeCenterY(nodeView, row), nodeCenterZ(nodeView, row));
  scl.setScalar(r);
  mat.compose(pos, quat, scl);
  mesh.setMatrixAt(0, mat);
  mesh.count = 1;
  mesh.instanceMatrix.needsUpdate = true;
}
