import * as THREE from "three";
import { type DecodedNodeFrame, nodeLabel } from "../decode/buffer-decode-node";
import { type ViewBlocks } from "../scene/view-blocks";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readNodeRadius, readNodeSphereR, readNodeSelected, readNodeLatchedSel,
  readNodePoleTheta, readNodePolePhi,
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
} from "../../../schema/buffer-layout/buffer-layout";

export interface NavNode {

  row: number;

  label: string;
  center: THREE.Vector3;
  radius: number;

  sphereR: number | undefined;
  selected: boolean;

  latchedSel: boolean;

  pole: THREE.Vector3;
}

function poleVec(theta: number, phi: number): THREE.Vector3 {
  const st = Math.sin(theta);
  return new THREE.Vector3(st * Math.cos(phi), Math.cos(theta), st * Math.sin(phi));
}

export function decodeNavNodes(decoded: DecodedNodeFrame): NavNode[] {
  const { nodeCount, nodeView } = decoded;
  const out: NavNode[] = [];
  for (let i = 0; i < nodeCount; i++) {
    const sphereR = readNodeSphereR(nodeView, i);
    out.push({
      row: i,
      label: nodeLabel(decoded, i),
      center: new THREE.Vector3(
        readNodeCX(nodeView, i),
        readNodeCY(nodeView, i),
        readNodeCZ(nodeView, i),
      ),
      radius: readNodeRadius(nodeView, i),

      sphereR: sphereR || undefined,
      selected: readNodeSelected(nodeView, i) !== 0,
      latchedSel: readNodeLatchedSel(nodeView, i) !== 0,
      pole: poleVec(readNodePoleTheta(nodeView, i), readNodePolePhi(nodeView, i)),
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
