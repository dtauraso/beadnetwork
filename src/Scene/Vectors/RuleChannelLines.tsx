import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { columnBytes } from "../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../Buffer/column-owners";
import { overlayFlag } from "../../webview/flags/overlay-flags";
import {
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M0, COL_STREAM_CHANNEL_VECTOR_SHAFT_M1,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M2, COL_STREAM_CHANNEL_VECTOR_SHAFT_M3,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M4, COL_STREAM_CHANNEL_VECTOR_SHAFT_M5,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M6, COL_STREAM_CHANNEL_VECTOR_SHAFT_M7,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M8, COL_STREAM_CHANNEL_VECTOR_SHAFT_M9,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M10, COL_STREAM_CHANNEL_VECTOR_SHAFT_M11,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M12, COL_STREAM_CHANNEL_VECTOR_SHAFT_M13,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M14, COL_STREAM_CHANNEL_VECTOR_SHAFT_M15,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M0, COL_STREAM_CHANNEL_VECTOR_HEAD_M1,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M2, COL_STREAM_CHANNEL_VECTOR_HEAD_M3,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M4, COL_STREAM_CHANNEL_VECTOR_HEAD_M5,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M6, COL_STREAM_CHANNEL_VECTOR_HEAD_M7,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M8, COL_STREAM_CHANNEL_VECTOR_HEAD_M9,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M10, COL_STREAM_CHANNEL_VECTOR_HEAD_M11,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M12, COL_STREAM_CHANNEL_VECTOR_HEAD_M13,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M14, COL_STREAM_CHANNEL_VECTOR_HEAD_M15,
} from "./columns-gen";
import {
  SHADING_PARAM_CHANNEL_LINE_RADIUS,
  SHADING_PARAM_CHANNEL_HEAD_RADIUS,
  SHADING_PARAM_CHANNEL_HEAD_LENGTH,
} from "../../Buffer/shading-params";

const CHANNEL_COLOR = "#7b6bd6";

const CHANNEL_LINE_OPACITY = 0.2584;
const CHANNEL_HEAD_OPACITY = 0.3675;

const SHAFT_COLS = [
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M0, COL_STREAM_CHANNEL_VECTOR_SHAFT_M1,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M2, COL_STREAM_CHANNEL_VECTOR_SHAFT_M3,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M4, COL_STREAM_CHANNEL_VECTOR_SHAFT_M5,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M6, COL_STREAM_CHANNEL_VECTOR_SHAFT_M7,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M8, COL_STREAM_CHANNEL_VECTOR_SHAFT_M9,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M10, COL_STREAM_CHANNEL_VECTOR_SHAFT_M11,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M12, COL_STREAM_CHANNEL_VECTOR_SHAFT_M13,
  COL_STREAM_CHANNEL_VECTOR_SHAFT_M14, COL_STREAM_CHANNEL_VECTOR_SHAFT_M15,
];

const HEAD_COLS = [
  COL_STREAM_CHANNEL_VECTOR_HEAD_M0, COL_STREAM_CHANNEL_VECTOR_HEAD_M1,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M2, COL_STREAM_CHANNEL_VECTOR_HEAD_M3,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M4, COL_STREAM_CHANNEL_VECTOR_HEAD_M5,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M6, COL_STREAM_CHANNEL_VECTOR_HEAD_M7,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M8, COL_STREAM_CHANNEL_VECTOR_HEAD_M9,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M10, COL_STREAM_CHANNEL_VECTOR_HEAD_M11,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M12, COL_STREAM_CHANNEL_VECTOR_HEAD_M13,
  COL_STREAM_CHANNEL_VECTOR_HEAD_M14, COL_STREAM_CHANNEL_VECTOR_HEAD_M15,
];

function copyMatrix(
  cols: Array<DataView | undefined>, vector: number,
  mesh: THREE.InstancedMesh, slot: number,
): void {
  const out = mesh.instanceMatrix.array;
  const b = slot * 16;
  const o = vector * 4;
  for (let m = 0; m < 16; m++) {
    const col = cols[m];
    out[b + m] = col && col.byteLength >= o + 4 ? col.getFloat32(o, true) : 0;
  }
}

export function RuleChannelLines({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    if (!line || !head) return;

    if (!overlayFlag("ruleChannels")) {
      line.count = 0;
      head.count = 0;
      return;
    }

    let drawn = 0;
    const { nodes } = ownerCounts();
    for (let row = 0; row < nodes && drawn < capacity; row++) {
      const shaftCols = SHAFT_COLS.map((c) => columnBytes(nodeColumn(row, c)));
      const headCols = HEAD_COLS.map((c) => columnBytes(nodeColumn(row, c)));
      const first = shaftCols[0];
      if (!first || first.byteLength === 0) continue;

      const vectors = first.byteLength >> 2;
      for (let v = 0; v < vectors && drawn < capacity; v++) {
        copyMatrix(shaftCols, v, line, drawn);
        copyMatrix(headCols, v, head, drawn);
        drawn++;
      }
    }

    line.count = drawn;
    head.count = drawn;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    if (drawn > 0) {
      line.computeBoundingSphere();
      head.computeBoundingSphere();
    }
  });

  return (
    <>
      <instancedMesh ref={lineRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[SHADING_PARAM_CHANNEL_LINE_RADIUS, SHADING_PARAM_CHANNEL_LINE_RADIUS, 1, 6]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={CHANNEL_LINE_OPACITY} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[SHADING_PARAM_CHANNEL_HEAD_RADIUS, SHADING_PARAM_CHANNEL_HEAD_LENGTH, 8]} />
        <meshBasicMaterial color={CHANNEL_COLOR} toneMapped={false} transparent opacity={CHANNEL_HEAD_OPACITY} />
      </instancedMesh>
    </>
  );
}
