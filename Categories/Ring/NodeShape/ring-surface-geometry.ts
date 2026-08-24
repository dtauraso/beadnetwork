import * as THREE from "three";
import { nodeRingPointBytes } from "./ring-point-leaves";
import {
  SHADING_PARAM_NODE_RING_SURFACE_NU,
  SHADING_PARAM_NODE_RING_SURFACE_NV,
} from "./shading-params";

let cachedGeometry: THREE.BufferGeometry | null = null;

export function getCanonicalRingSurfaceGeometry(): THREE.BufferGeometry | null {
  if (cachedGeometry) return cachedGeometry;

  const xs = nodeRingPointBytes("x");
  const ys = nodeRingPointBytes("y");
  const zs = nodeRingPointBytes("z");
  if (!xs || !ys || !zs) return null;

  const nu = SHADING_PARAM_NODE_RING_SURFACE_NU;
  const nv = SHADING_PARAM_NODE_RING_SURFACE_NV;
  const count = nu * nv;
  if (xs.byteLength < count * 4) return null;

  const positions = new Float32Array(count * 3);
  for (let k = 0; k < count; k++) {
    positions[k * 3]     = xs.getFloat32(k * 4, true);
    positions[k * 3 + 1] = ys.getFloat32(k * 4, true);
    positions[k * 3 + 2] = zs.getFloat32(k * 4, true);
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
  return cachedGeometry;
}
