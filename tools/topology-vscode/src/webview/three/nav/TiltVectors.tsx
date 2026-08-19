import React, { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { columnBytes } from "../../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../../Buffer/column-owners";
import {
  COL_STREAM_TILT_ARROW_RECEIVED,
  COL_STREAM_TILT_ARROW_SHAFT_M0, COL_STREAM_TILT_ARROW_SHAFT_M1,
  COL_STREAM_TILT_ARROW_SHAFT_M2, COL_STREAM_TILT_ARROW_SHAFT_M3,
  COL_STREAM_TILT_ARROW_SHAFT_M4, COL_STREAM_TILT_ARROW_SHAFT_M5,
  COL_STREAM_TILT_ARROW_SHAFT_M6, COL_STREAM_TILT_ARROW_SHAFT_M7,
  COL_STREAM_TILT_ARROW_SHAFT_M8, COL_STREAM_TILT_ARROW_SHAFT_M9,
  COL_STREAM_TILT_ARROW_SHAFT_M10, COL_STREAM_TILT_ARROW_SHAFT_M11,
  COL_STREAM_TILT_ARROW_SHAFT_M12, COL_STREAM_TILT_ARROW_SHAFT_M13,
  COL_STREAM_TILT_ARROW_SHAFT_M14, COL_STREAM_TILT_ARROW_SHAFT_M15,
  COL_STREAM_TILT_ARROW_HEAD_M0, COL_STREAM_TILT_ARROW_HEAD_M1,
  COL_STREAM_TILT_ARROW_HEAD_M2, COL_STREAM_TILT_ARROW_HEAD_M3,
  COL_STREAM_TILT_ARROW_HEAD_M4, COL_STREAM_TILT_ARROW_HEAD_M5,
  COL_STREAM_TILT_ARROW_HEAD_M6, COL_STREAM_TILT_ARROW_HEAD_M7,
  COL_STREAM_TILT_ARROW_HEAD_M8, COL_STREAM_TILT_ARROW_HEAD_M9,
  COL_STREAM_TILT_ARROW_HEAD_M10, COL_STREAM_TILT_ARROW_HEAD_M11,
  COL_STREAM_TILT_ARROW_HEAD_M12, COL_STREAM_TILT_ARROW_HEAD_M13,
  COL_STREAM_TILT_ARROW_HEAD_M14, COL_STREAM_TILT_ARROW_HEAD_M15,
} from "../../../Buffer/column-streams-gen";

const VECTOR_COLOR = "#FF2E88";

const RECEIVED_VECTOR_COLOR = "#00E5FF";

const SHAFT_COLS = [
  COL_STREAM_TILT_ARROW_SHAFT_M0, COL_STREAM_TILT_ARROW_SHAFT_M1,
  COL_STREAM_TILT_ARROW_SHAFT_M2, COL_STREAM_TILT_ARROW_SHAFT_M3,
  COL_STREAM_TILT_ARROW_SHAFT_M4, COL_STREAM_TILT_ARROW_SHAFT_M5,
  COL_STREAM_TILT_ARROW_SHAFT_M6, COL_STREAM_TILT_ARROW_SHAFT_M7,
  COL_STREAM_TILT_ARROW_SHAFT_M8, COL_STREAM_TILT_ARROW_SHAFT_M9,
  COL_STREAM_TILT_ARROW_SHAFT_M10, COL_STREAM_TILT_ARROW_SHAFT_M11,
  COL_STREAM_TILT_ARROW_SHAFT_M12, COL_STREAM_TILT_ARROW_SHAFT_M13,
  COL_STREAM_TILT_ARROW_SHAFT_M14, COL_STREAM_TILT_ARROW_SHAFT_M15,
];

const HEAD_COLS = [
  COL_STREAM_TILT_ARROW_HEAD_M0, COL_STREAM_TILT_ARROW_HEAD_M1,
  COL_STREAM_TILT_ARROW_HEAD_M2, COL_STREAM_TILT_ARROW_HEAD_M3,
  COL_STREAM_TILT_ARROW_HEAD_M4, COL_STREAM_TILT_ARROW_HEAD_M5,
  COL_STREAM_TILT_ARROW_HEAD_M6, COL_STREAM_TILT_ARROW_HEAD_M7,
  COL_STREAM_TILT_ARROW_HEAD_M8, COL_STREAM_TILT_ARROW_HEAD_M9,
  COL_STREAM_TILT_ARROW_HEAD_M10, COL_STREAM_TILT_ARROW_HEAD_M11,
  COL_STREAM_TILT_ARROW_HEAD_M12, COL_STREAM_TILT_ARROW_HEAD_M13,
  COL_STREAM_TILT_ARROW_HEAD_M14, COL_STREAM_TILT_ARROW_HEAD_M15,
];

function copyMatrix(
  cols: Array<DataView | undefined>, arrow: number,
  mesh: THREE.InstancedMesh, slot: number,
): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  const o = arrow * 4;
  for (let m = 0; m < 16; m++) {
    const col = cols[m];
    out[b + m] = col && col.byteLength >= o + 4 ? col.getFloat32(o, true) : 0;
  }
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

    let drawn = 0;
    let receivedDrawn = 0;

    const { nodes } = ownerCounts();
    for (let row = 0; row < nodes; row++) {
      const received = columnBytes(nodeColumn(row, COL_STREAM_TILT_ARROW_RECEIVED));
      if (!received || received.byteLength === 0) continue;
      const shaftCols = SHAFT_COLS.map((c) => columnBytes(nodeColumn(row, c)));
      const headCols = HEAD_COLS.map((c) => columnBytes(nodeColumn(row, c)));

      for (let arrow = 0; arrow < received.byteLength; arrow++) {
        if (received.getUint8(arrow) !== 0) {
          if (receivedDrawn >= receivedCapacity) continue;
          copyMatrix(shaftCols, arrow, receivedShaft, receivedDrawn);
          copyMatrix(headCols, arrow, receivedHead, receivedDrawn);
          receivedDrawn++;
          continue;
        }
        if (drawn >= capacity) continue;
        copyMatrix(shaftCols, arrow, shaft, drawn);
        copyMatrix(headCols, arrow, head, drawn);
        drawn++;
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
