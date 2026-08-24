import * as THREE from "three";
import { nodeLabel } from "./Labels/node-label";
import { polarToCart } from "../Polar/polar-convert";
import { sceneSteps, sceneRadius } from "../Scene/scene-frame";
import { ownerCounts } from "../Scene/owner-counts";
import { nodeF32, nodeU8 } from "./node-leaves";
import { readSelectedNodeRow } from "../Scene/View/Flags/overlay-flags-selection";

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

export function decodeNavNodes(): NavNode[] {
  const { nodes } = ownerCounts();
  const selected = readSelectedNodeRow();
  const out: NavNode[] = [];
  for (let i = 0; i < nodes; i++) {
    out.push({
      row: i,
      label: nodeLabel(i),
      center: new THREE.Vector3(
        nodeF32(i, "poleAnchorX"),
        nodeF32(i, "poleAnchorY"),
        nodeF32(i, "poleAnchorZ"),
      ),
      radius: nodeF32(i, "navTubeR"),

      selected: selected === i,
      latchedSel: nodeU8(i, "latchedSel") !== 0,
      pole: poleVec(
        nodeF32(i, "polePhi"),
        nodeF32(i, "poleTheta"),
      ),
      poleRingR: nodeF32(i, "poleRingR"),
    });
  }
  return out;
}

export function sceneSphereFromColumns(): { center: THREE.Vector3; radius: number } {
  const radius = sceneRadius();
  if (radius <= 0) return { center: new THREE.Vector3(), radius: 100 };
  const s = sceneSteps();
  return { center: new THREE.Vector3(s.centerX, s.centerY, s.centerZ), radius };
}

export function navSignature(nav: NavNode[]): string {
  let s = "";
  for (const n of nav) {
    s += `${n.row}:${Math.round(n.center.x)},${Math.round(n.center.y)},${Math.round(n.center.z)},${Math.round(n.radius)},${n.selected ? 1 : 0},${n.latchedSel ? 1 : 0};`;
  }
  return s;
}
