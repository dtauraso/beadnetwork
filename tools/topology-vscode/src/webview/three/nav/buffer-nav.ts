import * as THREE from "three";
import { type DecodedNodeFrame, nodeLabel } from "../decode/buffer-decode-node";
import { type ViewBlocks } from "../scene/view-blocks";
import { polarToCart } from "../polar-convert";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readNodeRadius, readNodeSelected, readNodeLatchedSel,
  readNodePolePhi, readNodePoleTheta, readNodePoleRingR,
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
} from "../../../../Buffer/buffer-layout";

export interface NavNode {

  row: number;

  label: string;
  center: THREE.Vector3;
  radius: number;

  selected: boolean;

  latchedSel: boolean;

  pole: THREE.Vector3;

  poleRingR: number;
}

function poleVec(phi: number, theta: number): THREE.Vector3 {
  return new THREE.Vector3(...polarToCart(1, phi, theta));
}

export function decodeNavNodes(decoded: DecodedNodeFrame): NavNode[] {
  const { nodeCount, nodeView } = decoded;
  const out: NavNode[] = [];
  for (let i = 0; i < nodeCount; i++) {
    out.push({
      row: i,
      label: nodeLabel(decoded, i),
      center: new THREE.Vector3(
        readNodeCX(nodeView, i),
        readNodeCY(nodeView, i),
        readNodeCZ(nodeView, i),
      ),
      radius: readNodeRadius(nodeView, i),

      selected: readNodeSelected(nodeView, i) !== 0,
      latchedSel: readNodeLatchedSel(nodeView, i) !== 0,
      pole: poleVec(readNodePolePhi(nodeView, i), readNodePoleTheta(nodeView, i)),
      poleRingR: readNodePoleRingR(nodeView, i),
    });
  }
  return out;
}

export function sceneSphereFromSnapshot(decoded: ViewBlocks): { center: THREE.Vector3; radius: number } {
  const radius = readSceneRadius(decoded.sceneView);
  if (radius <= 0) return { center: new THREE.Vector3(), radius: 100 };
  return {
    center: new THREE.Vector3(
      readSceneCX(decoded.sceneView),
      readSceneCY(decoded.sceneView),
      readSceneCZ(decoded.sceneView),
    ),
    radius,
  };
}
