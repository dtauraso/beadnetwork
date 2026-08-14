import * as THREE from "three";
import { getNodeFrame } from "./node-frame-aggregate";
import { getViewBlocks } from "../view-blocks";
import {
  readNodeRingAxisPhi,
  readNodeRingAxisTheta,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOverlaySelSpherePoles,
  readOverlayNodeBody, readOverlayNodeRing, readOverlayRingPick,
} from "../../../../schema/buffer-layout/buffer-layout";
import { NODE_SPHERE_RADIUS, nodeRowColors, poleAxis } from "../buffer-scene-shared";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";

const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

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

  const selectModeOn = readOverlaySelSpherePoles(overlayView) !== 0;

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
    ring.setMatrixAt(slot, mat);

    ringPick.setMatrixAt(slot, mat);
    ringBand.setMatrixAt(slot, mat);

    const { fill, stroke } = nodeRowColors(nodeView, row);
    body.setColorAt(slot, col.set(fill));
    ring.setColorAt(slot, col.set(stroke));
  }
  body.count = showBody ? n : 0;
  ring.count = showRing ? n : 0;

  ringPick.count = selectModeOn ? n : 0;

  ringBand.count = showRing && showPickBand ? n : 0;
  body.instanceMatrix.needsUpdate = true;
  ring.instanceMatrix.needsUpdate = true;
  ringPick.instanceMatrix.needsUpdate = true;
  ringBand.instanceMatrix.needsUpdate = true;
  if (body.instanceColor) body.instanceColor.needsUpdate = true;
  if (ring.instanceColor) ring.instanceColor.needsUpdate = true;

  if (showBody) body.computeBoundingSphere();
  if (selectModeOn) ringPick.computeBoundingSphere();
}
