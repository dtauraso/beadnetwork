import React, { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "../scene/nodes/node-frame-aggregate";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readNodeTopTiltVectorLen, readNodeTopTiltVectorTheta,
  readNodeBottomTiltVectorTheta,
  readNodeCoplanarNormalTheta,
  readNodeReceivedVectorLen, readNodeReceivedVectorTheta,
} from "../../../schema/buffer-layout/buffer-layout";

const SHAFT_RADIUS_FRAC = 0.035;
const HEAD_LEN_FRAC = 0.22;
const HEAD_RADIUS_FRAC = 0.09;

const VECTOR_COLOR = "#FF2E88";

const RECEIVED_VECTOR_COLOR = "#00E5FF";

const GEOMETRY_AXIS = new THREE.Vector3(0, 1, 0);

export function TiltVectors({ capacity, receivedCapacity }: { capacity: number; receivedCapacity: number }) {
  const shaftRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const receivedShaftRef = useRef<THREE.InstancedMesh>(null);
  const receivedHeadRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const posRef = useRef(new THREE.Vector3());
  const axisRef = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const shaft = shaftRef.current;
    const head = headRef.current;
    const receivedShaft = receivedShaftRef.current;
    const receivedHead = receivedHeadRef.current;
    if (!shaft || !head || !receivedShaft || !receivedHead) return;

    const decoded = getNodeFrame();
    if (!decoded) {
      shaft.count = 0;
      head.count = 0;
      receivedShaft.count = 0;
      receivedHead.count = 0;
      return;
    }
    const { nodeCount, nodeView } = decoded;

    const writeArrowInto = (
      targetShaft: THREE.InstancedMesh, targetHead: THREE.InstancedMesh, idx: number,
      cx: number, cy: number, cz: number, len: number, theta: number,
    ) => {

      if (theta === 0) {
        axisRef.current.set(0, 1, 0);
      } else {
        axisRef.current.set(Math.sin(theta), Math.cos(theta), 0);
      }
      quatRef.current.setFromUnitVectors(GEOMETRY_AXIS, axisRef.current);

      const shaftLen = len * (1 - HEAD_LEN_FRAC);
      posRef.current.set(
        cx + axisRef.current.x * (shaftLen / 2),
        cy + axisRef.current.y * (shaftLen / 2),
        cz + axisRef.current.z * (shaftLen / 2),
      );
      sclRef.current.set(len * SHAFT_RADIUS_FRAC, shaftLen, len * SHAFT_RADIUS_FRAC);
      matRef.current.compose(posRef.current, quatRef.current, sclRef.current);
      targetShaft.setMatrixAt(idx, matRef.current);

      const headLen = len * HEAD_LEN_FRAC;
      const headCentre = len - headLen / 2;
      posRef.current.set(
        cx + axisRef.current.x * headCentre,
        cy + axisRef.current.y * headCentre,
        cz + axisRef.current.z * headCentre,
      );
      sclRef.current.set(len * HEAD_RADIUS_FRAC, headLen, len * HEAD_RADIUS_FRAC);
      matRef.current.compose(posRef.current, quatRef.current, sclRef.current);
      targetHead.setMatrixAt(idx, matRef.current);
    };

    let drawn = 0;

    let receivedDrawn = 0;

    for (let row = 0; row < nodeCount; row++) {
      const cx = readNodeCX(nodeView, row);
      const cy = readNodeCY(nodeView, row);
      const cz = readNodeCZ(nodeView, row);

      const len = readNodeTopTiltVectorLen(nodeView, row);
      if (len > 0 && drawn + 2 < capacity) {
        writeArrowInto(shaft, head, drawn, cx, cy, cz, len, readNodeTopTiltVectorTheta(nodeView, row));
        drawn++;
        writeArrowInto(shaft, head, drawn, cx, cy, cz, len, readNodeBottomTiltVectorTheta(nodeView, row));
        drawn++;
        writeArrowInto(shaft, head, drawn, cx, cy, cz, len, readNodeCoplanarNormalTheta(nodeView, row));
        drawn++;
      }

      const receivedLen = readNodeReceivedVectorLen(nodeView, row);
      if (receivedLen > 0 && receivedDrawn < receivedCapacity) {
        writeArrowInto(
          receivedShaft, receivedHead, receivedDrawn,
          cx, cy, cz, receivedLen,
          readNodeReceivedVectorTheta(nodeView, row),
        );
        receivedDrawn++;
      }
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
      {}
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
