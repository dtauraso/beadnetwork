import * as THREE from "three";
import { getViewBlocks } from "../view-blocks";
import { readRingPointX, readRingPointY, readRingPointZ } from "../../../../schema/buffer-layout/buffer-layout";
import {
  SHADING_PARAM_NODE_RING_SURFACE_NU,
  SHADING_PARAM_NODE_RING_SURFACE_NV,
} from "../../../../schema/buffer-layout/shading-params";

let cachedGeometry: THREE.BufferGeometry | null = null;

export function getCanonicalRingSurfaceGeometry(): THREE.BufferGeometry | null {
  if (cachedGeometry) return cachedGeometry;

  const blocks = getViewBlocks();
  if (!blocks) return null;
  const view = blocks.ringSurfacePointsView;

  const nu = SHADING_PARAM_NODE_RING_SURFACE_NU;
  const nv = SHADING_PARAM_NODE_RING_SURFACE_NV;
  const count = nu * nv;

  const positions = new Float32Array(count * 3);
  for (let k = 0; k < count; k++) {
    positions[k * 3]     = readRingPointX(view, k);
    positions[k * 3 + 1] = readRingPointY(view, k);
    positions[k * 3 + 2] = readRingPointZ(view, k);
  }

  const indices = new Uint32Array(nu * nv * 6);
  let ii = 0;
  for (let j = 0; j < nv; j++) {
    const jNext = (j + 1) % nv;
    for (let i = 0; i < nu; i++) {
      const iNext = (i + 1) % nu;
      const a0 = j * nu + i;
      const a1 = j * nu + iNext;
      const b0 = jNext * nu + i;
      const b1 = jNext * nu + iNext;
      indices[ii++] = a0; indices[ii++] = b0; indices[ii++] = a1;
      indices[ii++] = a1; indices[ii++] = b0; indices[ii++] = b1;
    }
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
  geometry.setIndex(new THREE.BufferAttribute(indices, 1));
  geometry.computeVertexNormals();
  geometry.computeBoundingSphere();

  cachedGeometry = geometry;
  return geometry;
}
