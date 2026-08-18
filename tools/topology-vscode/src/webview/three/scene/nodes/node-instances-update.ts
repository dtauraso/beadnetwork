import * as THREE from "three";
import { getNodeFrame } from "./node-frame-aggregate";
import { getViewBlocks } from "../view-blocks";
import {
  readNodeRingAxisPhi,
  readNodeRingAxisTheta,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOverlayNodeBody, readOverlayNodeRing, readOverlayRingPick,
  readNodeRingM0, readNodeRingM1, readNodeRingM2, readNodeRingM3,
  readNodeRingM4, readNodeRingM5, readNodeRingM6, readNodeRingM7,
  readNodeRingM8, readNodeRingM9, readNodeRingM10, readNodeRingM11,
  readNodeRingM12, readNodeRingM13, readNodeRingM14, readNodeRingM15,
} from "../../../../../Buffer/buffer-layout";
import { NODE_SPHERE_RADIUS, nodeRowColors, poleAxis } from "../buffer-scene-shared";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";

const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

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
  mat: THREE.Matrix4;
  pos: THREE.Vector3;
  quat: THREE.Quaternion;
  ringQuat: THREE.Quaternion;
  ringAxis: THREE.Vector3;
  scl: THREE.Vector3;
  col: THREE.Color;
}

export function updateNodeInstances(refs: NodeInstanceRefs, capacity: number, camera: THREE.Camera): void {
  const { body, ring, ringPick, ringBand, mat, pos, quat, ringQuat, ringAxis, scl, col } = refs;

  const blocks = getViewBlocks();
  const decodedNode = getNodeFrame();
  if (!decodedNode || !blocks) {
    body.count = 0; ring.count = 0; ringPick.count = 0; ringBand.count = 0;
    return;
  }
  const { overlayView } = blocks;
  const { nodeCount, nodeView } = decodedNode;

  const showBody = readOverlayNodeBody(overlayView) !== 0;
  const showRing = readOverlayNodeRing(overlayView) !== 0;

  const showPickBand = readOverlayRingPick(overlayView) !== 0;

  const n = Math.min(nodeCount, capacity);

  const order = computeNodeDepthOrder(
    n,
    (row) => readNodeCX(nodeView, row),
    (row) => readNodeCY(nodeView, row),
    (row) => readNodeCZ(nodeView, row),
    camera.position.x, camera.position.y, camera.position.z,
  );
  setNodeDrawOrder(order);
  for (let slot = 0; slot < n; slot++) {
    const row = order[slot]!;
    const r = readNodeRadius(nodeView, row) || NODE_SPHERE_RADIUS;
    pos.set(
      readNodeCX(nodeView, row),
      readNodeCY(nodeView, row),
      readNodeCZ(nodeView, row),
    );

    scl.setScalar(r);
    mat.compose(pos, quat, scl);
    body.setMatrixAt(slot, mat);

    const poleTheta = readNodeRingAxisPhi(nodeView, row);
    const polePhi = readNodeRingAxisTheta(nodeView, row);
    const [ax, ay, az] = poleAxis(poleTheta, polePhi);
    ringAxis.set(ax, ay, az);
    ringQuat.setFromUnitVectors(TORUS_DEFAULT_NORMAL, ringAxis);
    mat.compose(pos, ringQuat, scl);
    ringPick.setMatrixAt(slot, mat);
    ringBand.setMatrixAt(slot, mat);

    copyRingMatrix(nodeView, row, ring, slot);

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

  if (showBody) body.computeBoundingSphere();
  ringPick.computeBoundingSphere();
}
