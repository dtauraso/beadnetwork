import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeBeads } from "../edges/edge-bead-blocks";
import { beadStyleForValue } from "./bead-style";
import { getCanonicalBeadRingSurfaceGeometry } from "./bead-ring-surface-geometry";
import { SHADING_PARAM_BEAD_RADIUS } from "../../../../../Buffer/shading-params";

const RING_COLOR = beadStyleForValue(1)!.ring;

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const litBodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const posRef = useRef(new THREE.Vector3());
  const colRef = useRef(new THREE.Color());
  const ringGeomAppliedRef = useRef(false);

  useFrame(() => {
    const litBody = litBodyRef.current;
    const ring = ringRef.current;
    if (!litBody || !ring) return;

    if (!ringGeomAppliedRef.current) {
      const geom = getCanonicalBeadRingSurfaceGeometry();
      if (geom) {
        ring.geometry = geom;
        ringGeomAppliedRef.current = true;
      }
    }

    const { positions, ringMatrix, value, count } = getEdgeBeads();

    const drawn = Math.min(count, capacity);
    const ringOut = ring.instanceMatrix.array;

    let litCount = 0;
    for (let i = 0; i < drawn; i++) {
      const style = beadStyleForValue(value[i]);
      if (!style) continue;

      posRef.current.set(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      matRef.current.makeTranslation(posRef.current.x, posRef.current.y, posRef.current.z);
      litBody.setMatrixAt(litCount, matRef.current);
      litBody.setColorAt(litCount, colRef.current.set(style.fill));

      ringOut.set(ringMatrix.subarray(i * 16, i * 16 + 16), litCount * 16);
      ring.setColorAt(litCount, colRef.current.set(RING_COLOR));

      litCount++;
    }
    litBody.count = litCount;
    ring.count = litCount;
    litBody.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (litBody.instanceColor) litBody.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      <instancedMesh ref={litBodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <bufferGeometry />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
