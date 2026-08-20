import * as THREE from "three";
import { nodeLabel } from "../Node/buffer-decode-node";
import { polarToCart } from "../webview/three/polar-convert";
import { sceneSteps, sceneRadius } from "../Scene/scene-frame";
import { columnF32, columnU8 } from "../schema/buffer-layout/column-values";
import { nodeColumn, ownerCounts } from "../schema/buffer-layout/column-owners";
import {
  COL_STREAM_NODE_POLE_ANCHOR_X, COL_STREAM_NODE_POLE_ANCHOR_Y,
  COL_STREAM_NODE_POLE_ANCHOR_Z, COL_STREAM_NODE_NAV_TUBE_R, COL_STREAM_NODE_LATCHED_SEL,
  COL_STREAM_NODE_POLE_PHI, COL_STREAM_NODE_POLE_THETA, COL_STREAM_NODE_POLE_RING_R,
} from "../Node/columns-gen";
import { readSelectedNodeRow } from "../webview/three/controls/flags/overlay-flags-selection";

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
        columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_ANCHOR_X)),
        columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_ANCHOR_Y)),
        columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_ANCHOR_Z)),
      ),
      radius: columnF32(nodeColumn(i, COL_STREAM_NODE_NAV_TUBE_R)),

      selected: selected === i,
      latchedSel: columnU8(nodeColumn(i, COL_STREAM_NODE_LATCHED_SEL)) !== 0,
      pole: poleVec(
        columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_PHI)),
        columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_THETA)),
      ),
      poleRingR: columnF32(nodeColumn(i, COL_STREAM_NODE_POLE_RING_R)),
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
