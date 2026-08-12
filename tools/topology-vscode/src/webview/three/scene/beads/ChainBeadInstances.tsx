import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getChainBeads } from "../nodes/node-stream-blocks";
import { beadStyleForValue } from "./bead-style";
import {
  SHADING_PARAM_BEAD_RADIUS,
  SHADING_PARAM_BEAD_RING_TUBE_RATIO,
  SHADING_PARAM_CHAIN_BEAD_FILL,
} from "../../../../schema/shading-params";

const RING_COLOR = beadStyleForValue(1)!.ring;

const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);
const BEAD_UNIT_SCALE = new THREE.Vector3(1, 1, 1);

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const litBodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const beadRingMatRef = useRef(new THREE.Matrix4());
  const beadQuatRef = useRef(new THREE.Quaternion());
  const beadAxisRef = useRef(new THREE.Vector3());
  const beadPosRef = useRef(new THREE.Vector3());
  const colRef = useRef(new THREE.Color());

  useFrame(() => {
    const litBody = litBodyRef.current;
    const ring = ringRef.current;
    if (!litBody || !ring) return;

    const { positions, ringAxis, count, lit, litValue } = getChainBeads();

    const drawn = Math.min(count, capacity);

    let litCount = 0;
    let ringCount = 0;
    for (let i = 0; i < drawn; i++) {

      if (lit[i] !== 1) continue;

      if (!beadStyleForValue(litValue[i])) continue;

      matRef.current.makeTranslation(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);

      beadAxisRef.current.set(ringAxis[i * 3]!, ringAxis[i * 3 + 1]!, ringAxis[i * 3 + 2]!);
      beadQuatRef.current.setFromUnitVectors(TORUS_DEFAULT_NORMAL, beadAxisRef.current);
      beadRingMatRef.current.compose(
        beadPosRef.current.set(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!),
        beadQuatRef.current,
        BEAD_UNIT_SCALE,
      );
      ring.setMatrixAt(ringCount, beadRingMatRef.current);
      ring.setColorAt(ringCount, colRef.current.set(RING_COLOR));
      ringCount++;

      const style = beadStyleForValue(litValue[i]);
      if (!style) continue;
      litBody.setMatrixAt(litCount, matRef.current);
      litBody.setColorAt(litCount, colRef.current.set(style.fill));
      litCount++;
    }
    litBody.count = litCount;
    ring.count = ringCount;
    litBody.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (litBody.instanceColor) litBody.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      {}
      {}
      <instancedMesh ref={litBodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <torusGeometry
          args={[SHADING_PARAM_BEAD_RADIUS, SHADING_PARAM_BEAD_RADIUS * SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8, 24]}
        />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
