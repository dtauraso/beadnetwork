import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { INTERIOR_SLOTS_PER_NODE } from "./buffer-decode-interior";
import { interiorBeadStyleForValue } from "../Bead/bead-style";
import { columnBytes } from "../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../Buffer/column-owners";
import {
  COL_STREAM_INTERIOR_PRESENT, COL_STREAM_INTERIOR_VALUE,
  COL_STREAM_INTERIOR_X, COL_STREAM_INTERIOR_Y, COL_STREAM_INTERIOR_Z,
} from "../Buffer/column-streams-gen";

function presentAt(node: number, slot: number): number {
  const b = columnBytes(nodeColumn(node, COL_STREAM_INTERIOR_PRESENT));
  return b && slot < b.byteLength ? b.getUint8(slot) : 0;
}

function valueAt(node: number, slot: number): number {
  const b = columnBytes(nodeColumn(node, COL_STREAM_INTERIOR_VALUE));
  return b && slot * 4 + 4 <= b.byteLength ? b.getInt32(slot * 4, true) : 0;
}

function coordAt(node: number, slot: number, col: number): number {
  const b = columnBytes(nodeColumn(node, col));
  return b && slot * 4 + 4 <= b.byteLength ? b.getFloat32(slot * 4, true) : 0;
}

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

    const { nodes: nodeCount } = ownerCounts();
    if (nodeCount <= 0) { body.count = 0; ring.count = 0; return; }

    const q = quatRef.current; 
    sclRef.current.setScalar(INTERIOR_BEAD_R);
    let slot = 0;
    for (let i = 0; i < nodeCount && slot < capacity; i++) {
      for (let s = 0; s < INTERIOR_SLOTS_PER_NODE && slot < capacity; s++) {
                if (!presentAt(i, s)) continue;
        const style = interiorBeadStyleForValue(valueAt(i, s));
        if (!style) continue; 

        posRef.current.set(
          coordAt(i, s, COL_STREAM_INTERIOR_X),
          coordAt(i, s, COL_STREAM_INTERIOR_Y),
          coordAt(i, s, COL_STREAM_INTERIOR_Z),
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
