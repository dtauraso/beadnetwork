import { useContext, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { EnvTexContext } from "../../../Start/extension/webview/scene/scene-env";
import {
  SHADING_PARAM_NODE_TRANSMISSION,
  SHADING_PARAM_NODE_THICKNESS,
  SHADING_PARAM_NODE_ROUGHNESS,
  SHADING_PARAM_NODE_IOR,
  SHADING_PARAM_NODE_METALNESS,
  SHADING_PARAM_NODE_CLEARCOAT,
  SHADING_PARAM_NODE_CLEARCOAT_ROUGHNESS,
  SHADING_PARAM_NODE_ENV_MAP_INTENSITY,
  SHADING_PARAM_NODE_OPACITY,
  SHADING_PARAM_RING_ROUGHNESS,
} from "../../Ring/NodeShape/shading-params";
import { getCanonicalRingSurfaceGeometry } from "../../Ring/NodeShape/ring-surface-geometry";
import { ownerCounts } from "../../Scene/owner-counts";
import { nodeBytes, nodeF32, NODE_RING_NAMES, NODE_BODY_NAMES } from "../node-leaves";
import { type NodeValueName } from "../node-values-gen";
import { nodeRowColors } from "../node-kind";
import { overlayFlag } from "../../Scene/View/Flags/overlay-flags";

const CAPACITY = 64;

function placeCopy(
  mesh: THREE.InstancedMesh, slot: number, row: number,
  names: readonly NodeValueName[], x: number, y: number, z: number,
): void {
  const out = mesh.instanceMatrix.array;
  const base = slot * 16;
  for (let i = 0; i < 16; i++) out[base + i] = nodeF32(row, names[i]!);
  out[base + 12] = x;
  out[base + 13] = y;
  out[base + 14] = z;
}

export function TipCopies() {
  const envTex = useContext(EnvTexContext);
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const colRef = useRef(new THREE.Color());
  const ringGeomAppliedRef = useRef(false);

  useFrame(() => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    if (!body || !ring) return;
    if (!ringGeomAppliedRef.current) {
      const geom = getCanonicalRingSurfaceGeometry();
      if (geom) {
        ring.geometry = geom;
        ringGeomAppliedRef.current = true;
      }
    }

    const { nodes } = ownerCounts();
    let slot = 0;
    for (let row = 0; row < nodes && slot < CAPACITY; row++) {
      const xs = nodeBytes(row, "copyX");
      const ys = nodeBytes(row, "copyY");
      const zs = nodeBytes(row, "copyZ");
      const rows = nodeBytes(row, "copyRow");
      const count = xs && ys && zs && rows ? Math.min(xs.byteLength / 4, ys.byteLength / 4, zs.byteLength / 4, rows.byteLength) : 0;
      for (let i = 0; i < count && slot < CAPACITY; i++) {
        const to = rows!.getUint8(i);
        if (to >= nodes) continue;
        const x = xs!.getFloat32(i * 4, true);
        const y = ys!.getFloat32(i * 4, true);
        const z = zs!.getFloat32(i * 4, true);
        placeCopy(body, slot, to, NODE_BODY_NAMES, x, y, z);
        placeCopy(ring, slot, to, NODE_RING_NAMES, x, y, z);
        const { fill, stroke } = nodeRowColors(to);
        body.setColorAt(slot, colRef.current.set(fill));
        ring.setColorAt(slot, colRef.current.set(stroke));
        slot++;
      }
    }

    body.count = overlayFlag("nodeBody") ? slot : 0;
    ring.count = overlayFlag("nodeRing") ? slot : 0;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      <instancedMesh ref={bodyRef} args={[undefined, undefined, CAPACITY]} frustumCulled={false} raycast={() => null}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshPhysicalMaterial
          transmission={SHADING_PARAM_NODE_TRANSMISSION}
          thickness={SHADING_PARAM_NODE_THICKNESS}
          roughness={SHADING_PARAM_NODE_ROUGHNESS}
          ior={SHADING_PARAM_NODE_IOR}
          metalness={SHADING_PARAM_NODE_METALNESS}
          clearcoat={SHADING_PARAM_NODE_CLEARCOAT}
          clearcoatRoughness={SHADING_PARAM_NODE_CLEARCOAT_ROUGHNESS}
          envMap={envTex ?? undefined}
          envMapIntensity={SHADING_PARAM_NODE_ENV_MAP_INTENSITY}
          transparent
          opacity={SHADING_PARAM_NODE_OPACITY}
          depthWrite={false}
        />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, CAPACITY]} frustumCulled={false} raycast={() => null}>
        <bufferGeometry />
        <meshStandardMaterial roughness={SHADING_PARAM_RING_ROUGHNESS} metalness={0} depthWrite={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
