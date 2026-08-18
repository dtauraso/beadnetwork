import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../nodes/node-frame-aggregate";
import { INTERIOR_SLOTS_PER_NODE } from "../../decode/buffer-decode-interior";
import { interiorBeadStyleForValue } from "./bead-style";
import {
  readInteriorPresent, readInteriorValue, readInteriorX, readInteriorY, readInteriorZ,
} from "../../../../../Buffer/buffer-layout";

const INTERIOR_BEAD_R = 5;
const INTERIOR_RING_TUBE_RATIO = 0.12;

export function InteriorBeadInstances({ capacity }: { capacity: number }) {
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef  = useRef(new THREE.Matrix4());
  const posRef  = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef  = useRef(new THREE.Vector3());
  const colRef  = useRef(new THREE.Color());

  useFrame(() => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    if (!body || !ring) return;

    const decoded = getNodeFrame();
    if (!decoded) { body.count = 0; ring.count = 0; return; }
    const { nodeCount, nodeView, interiorView } = decoded;

    const q = quatRef.current; 
    sclRef.current.setScalar(INTERIOR_BEAD_R);
    let slot = 0;
    for (let i = 0; i < nodeCount && slot < capacity; i++) {
      for (let s = 0; s < INTERIOR_SLOTS_PER_NODE && slot < capacity; s++) {
        const row = i * INTERIOR_SLOTS_PER_NODE + s;
        if (!readInteriorPresent(interiorView, row)) continue;
        const style = interiorBeadStyleForValue(readInteriorValue(interiorView, row));
        if (!style) continue; 

        posRef.current.set(
          readInteriorX(interiorView, row),
          readInteriorY(interiorView, row),
          readInteriorZ(interiorView, row),
        );
        matRef.current.compose(posRef.current, q, sclRef.current);
        body.setMatrixAt(slot, matRef.current);
        ring.setMatrixAt(slot, matRef.current);
        body.setColorAt(slot, colRef.current.set(style.fill));
        ring.setColorAt(slot, colRef.current.set(style.ring));
        slot++;
      }
    }
    body.count = slot;
    ring.count = slot;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      {}
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} renderOrder={1} frustumCulled={false}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial toneMapped={false} transparent opacity={1} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} renderOrder={1} frustumCulled={false}>
        <torusGeometry args={[1, INTERIOR_RING_TUBE_RATIO, 8, 24]} />
        <meshBasicMaterial toneMapped={false} transparent opacity={1} />
      </instancedMesh>
    </>
  );
}
