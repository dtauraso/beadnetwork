import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { ownerCounts } from "../../Scene/owner-counts";
import {
  topVectorBytes, TOP_VECTOR_SHAFT_NAMES, TOP_VECTOR_HEAD_NAMES,
} from "./top-vector-leaves";

const TOP_VECTOR_COLOR = "#4fd6a0";

function copyMatrix(
  cols: Array<DataView | undefined>,
  mesh: THREE.InstancedMesh, slot: number,
): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  for (let m = 0; m < 16; m++) {
    const col = cols[m];
    out[b + m] = col && col.byteLength >= 4 ? col.getFloat32(0, true) : 0;
  }
}

export function TopVectors({ capacity }: { capacity: number }) {
  const shaftRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);

  useFrame(() => {
    const shaft = shaftRef.current;
    const head = headRef.current;
    if (!shaft || !head) return;

    let drawn = 0;
    const { nodes } = ownerCounts();
    for (let row = 0; row < nodes && drawn < capacity; row++) {
      const flag = topVectorBytes(row, "drawn");
      if (!flag || flag.byteLength === 0 || flag.getUint8(0) === 0) continue;

      copyMatrix(TOP_VECTOR_SHAFT_NAMES.map((n) => topVectorBytes(row, n)), shaft, drawn);
      copyMatrix(TOP_VECTOR_HEAD_NAMES.map((n) => topVectorBytes(row, n)), head, drawn);
      drawn++;
    }

    shaft.count = drawn;
    head.count = drawn;
    shaft.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    if (drawn > 0) {
      shaft.computeBoundingSphere();
      head.computeBoundingSphere();
    }
  });

  return (
    <>
      <instancedMesh ref={shaftRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[1, 1, 1, 10]} />
        <meshBasicMaterial color={TOP_VECTOR_COLOR} toneMapped={false} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[1, 1, 12]} />
        <meshBasicMaterial color={TOP_VECTOR_COLOR} toneMapped={false} />
      </instancedMesh>
    </>
  );
}
