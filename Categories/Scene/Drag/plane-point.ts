import * as THREE from "three";
import { nodeF32, nodeU8 } from "../../Node/node-leaves";
import { ownerCounts } from "../owner-counts";
import { sceneValue } from "../scene-leaves";

const ray = new THREE.Ray();
const plane = new THREE.Plane();
const anchor = new THREE.Vector3();
const forward = new THREE.Vector3();
const hit = new THREE.Vector3();

function nodeCentre(row: number, out: THREE.Vector3): boolean {
  if (!(row >= 0)) return false;
  out.set(nodeF32(row, "bodyM12"), nodeF32(row, "bodyM13"), nodeF32(row, "bodyM14"));
  return true;
}

function selectedRow(): number {
  const { nodes } = ownerCounts();
  for (let row = 0; row < nodes; row++) {
    if (nodeU8(row, "selected") !== 0) return row;
  }
  return -1;
}

export function planePoint(
  cam: THREE.PerspectiveCamera,
  ndcX: number,
  ndcY: number,
  nodeRow: number,
): { x: number; y: number; z: number } {
  if (!nodeCentre(nodeRow, anchor) && !nodeCentre(selectedRow(), anchor)) {
    anchor.set(sceneValue("cx"), sceneValue("cy"), sceneValue("cz"));
  }

  cam.getWorldDirection(forward);
  plane.setFromNormalAndCoplanarPoint(forward, anchor);

  ray.origin.setFromMatrixPosition(cam.matrixWorld);
  ray.direction.set(ndcX, ndcY, 0.5).unproject(cam).sub(ray.origin).normalize();

  if (!ray.intersectPlane(plane, hit)) return { x: 0, y: 0, z: 0 };
  return { x: hit.x, y: hit.y, z: hit.z };
}
