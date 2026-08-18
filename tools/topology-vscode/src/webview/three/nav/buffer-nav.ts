import * as THREE from "three";
import { type DecodedNodeFrame, nodeLabel } from "../decode/buffer-decode-node";
import { type ViewBlocks } from "../scene/view-blocks";
import { polarToCart } from "../polar-convert";
import { nodeCenterX, nodeCenterY, nodeCenterZ, nodeRadiusRaw, nodeSelected } from "../../../../Node/node-frame";
import { sceneSteps, sceneRadius } from "../../../../Scene/scene-frame";
import {
  readNodeLatchedSel,
  readNodePolePhi, readNodePoleTheta, readNodePoleRingR,
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
        nodeCenterX(nodeView, i),
        nodeCenterY(nodeView, i),
        nodeCenterZ(nodeView, i),
      ),
      radius: nodeRadiusRaw(nodeView, i),

      selected: nodeSelected(nodeView, i),
      latchedSel: readNodeLatchedSel(nodeView, i) !== 0,
      pole: poleVec(readNodePolePhi(nodeView, i), readNodePoleTheta(nodeView, i)),
      poleRingR: readNodePoleRingR(nodeView, i),
    });
  }
  return out;
}

export function sceneSphereFromSnapshot(decoded: ViewBlocks): { center: THREE.Vector3; radius: number } {
  const radius = sceneRadius();
  if (radius <= 0) return { center: new THREE.Vector3(), radius: 100 };
  const s = sceneSteps();
  return { center: new THREE.Vector3(s.centerX, s.centerY, s.centerZ), radius };
}
