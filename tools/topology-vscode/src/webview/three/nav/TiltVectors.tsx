import React, { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeSections } from "../scene/nodes/node-sections";
import {
  readTiltArrowReceived,
  readTiltArrowShaftM0, readTiltArrowShaftM1, readTiltArrowShaftM2, readTiltArrowShaftM3,
  readTiltArrowShaftM4, readTiltArrowShaftM5, readTiltArrowShaftM6, readTiltArrowShaftM7,
  readTiltArrowShaftM8, readTiltArrowShaftM9, readTiltArrowShaftM10, readTiltArrowShaftM11,
  readTiltArrowShaftM12, readTiltArrowShaftM13, readTiltArrowShaftM14, readTiltArrowShaftM15,
  readTiltArrowHeadM0, readTiltArrowHeadM1, readTiltArrowHeadM2, readTiltArrowHeadM3,
  readTiltArrowHeadM4, readTiltArrowHeadM5, readTiltArrowHeadM6, readTiltArrowHeadM7,
  readTiltArrowHeadM8, readTiltArrowHeadM9, readTiltArrowHeadM10, readTiltArrowHeadM11,
  readTiltArrowHeadM12, readTiltArrowHeadM13, readTiltArrowHeadM14, readTiltArrowHeadM15,
} from "../../../../Buffer/buffer-layout";

const VECTOR_COLOR = "#FF2E88";

const RECEIVED_VECTOR_COLOR = "#00E5FF";

function copyShaft(view: DataView, row: number, mesh: THREE.InstancedMesh, slot: number): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  out[b]      = readTiltArrowShaftM0(view, row);
  out[b + 1]  = readTiltArrowShaftM1(view, row);
  out[b + 2]  = readTiltArrowShaftM2(view, row);
  out[b + 3]  = readTiltArrowShaftM3(view, row);
  out[b + 4]  = readTiltArrowShaftM4(view, row);
  out[b + 5]  = readTiltArrowShaftM5(view, row);
  out[b + 6]  = readTiltArrowShaftM6(view, row);
  out[b + 7]  = readTiltArrowShaftM7(view, row);
  out[b + 8]  = readTiltArrowShaftM8(view, row);
  out[b + 9]  = readTiltArrowShaftM9(view, row);
  out[b + 10] = readTiltArrowShaftM10(view, row);
  out[b + 11] = readTiltArrowShaftM11(view, row);
  out[b + 12] = readTiltArrowShaftM12(view, row);
  out[b + 13] = readTiltArrowShaftM13(view, row);
  out[b + 14] = readTiltArrowShaftM14(view, row);
  out[b + 15] = readTiltArrowShaftM15(view, row);
}

function copyHead(view: DataView, row: number, mesh: THREE.InstancedMesh, slot: number): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  out[b]      = readTiltArrowHeadM0(view, row);
  out[b + 1]  = readTiltArrowHeadM1(view, row);
  out[b + 2]  = readTiltArrowHeadM2(view, row);
  out[b + 3]  = readTiltArrowHeadM3(view, row);
  out[b + 4]  = readTiltArrowHeadM4(view, row);
  out[b + 5]  = readTiltArrowHeadM5(view, row);
  out[b + 6]  = readTiltArrowHeadM6(view, row);
  out[b + 7]  = readTiltArrowHeadM7(view, row);
  out[b + 8]  = readTiltArrowHeadM8(view, row);
  out[b + 9]  = readTiltArrowHeadM9(view, row);
  out[b + 10] = readTiltArrowHeadM10(view, row);
  out[b + 11] = readTiltArrowHeadM11(view, row);
  out[b + 12] = readTiltArrowHeadM12(view, row);
  out[b + 13] = readTiltArrowHeadM13(view, row);
  out[b + 14] = readTiltArrowHeadM14(view, row);
  out[b + 15] = readTiltArrowHeadM15(view, row);
}

export function TiltVectors({ capacity, receivedCapacity }: { capacity: number; receivedCapacity: number }) {
  const shaftRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const receivedShaftRef = useRef<THREE.InstancedMesh>(null);
  const receivedHeadRef = useRef<THREE.InstancedMesh>(null);

  useFrame(() => {
    const shaft = shaftRef.current;
    const head = headRef.current;
    const receivedShaft = receivedShaftRef.current;
    const receivedHead = receivedHeadRef.current;
    if (!shaft || !head || !receivedShaft || !receivedHead) return;

    const decoded = getNodeSections();
    if (!decoded) {
      shaft.count = 0;
      head.count = 0;
      receivedShaft.count = 0;
      receivedHead.count = 0;
      return;
    }
    const { tiltArrowCount, tiltArrowView } = decoded;

    let drawn = 0;
    let receivedDrawn = 0;

    for (let row = 0; row < tiltArrowCount; row++) {
      if (readTiltArrowReceived(tiltArrowView, row) !== 0) {
        if (receivedDrawn >= receivedCapacity) continue;
        copyShaft(tiltArrowView, row, receivedShaft, receivedDrawn);
        copyHead(tiltArrowView, row, receivedHead, receivedDrawn);
        receivedDrawn++;
        continue;
      }
      if (drawn >= capacity) continue;
      copyShaft(tiltArrowView, row, shaft, drawn);
      copyHead(tiltArrowView, row, head, drawn);
      drawn++;
    }

    shaft.count = drawn;
    head.count = drawn;
    shaft.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    receivedShaft.count = receivedDrawn;
    receivedHead.count = receivedDrawn;
    receivedShaft.instanceMatrix.needsUpdate = true;
    receivedHead.instanceMatrix.needsUpdate = true;
  });

  return (
    <>
      {}
      <instancedMesh ref={shaftRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[1, 1, 1, 12]} />
        <meshBasicMaterial color={VECTOR_COLOR} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[1, 1, 14]} />
        <meshBasicMaterial color={VECTOR_COLOR} />
      </instancedMesh>
      <instancedMesh ref={receivedShaftRef} args={[undefined, undefined, receivedCapacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[1, 1, 1, 12]} />
        <meshBasicMaterial color={RECEIVED_VECTOR_COLOR} />
      </instancedMesh>
      <instancedMesh ref={receivedHeadRef} args={[undefined, undefined, receivedCapacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[1, 1, 14]} />
        <meshBasicMaterial color={RECEIVED_VECTOR_COLOR} />
      </instancedMesh>
    </>
  );
}
