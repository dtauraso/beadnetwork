import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { columnBytes } from "../../../../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../../../../Buffer/column-owners";
import {
  COL_STREAM_EDGE_BEAD_X, COL_STREAM_EDGE_BEAD_Y, COL_STREAM_EDGE_BEAD_Z,
  COL_STREAM_EDGE_BEAD_VALUE,
  COL_STREAM_EDGE_BEAD_RING_M0, COL_STREAM_EDGE_BEAD_RING_M1,
  COL_STREAM_EDGE_BEAD_RING_M2, COL_STREAM_EDGE_BEAD_RING_M3,
  COL_STREAM_EDGE_BEAD_RING_M4, COL_STREAM_EDGE_BEAD_RING_M5,
  COL_STREAM_EDGE_BEAD_RING_M6, COL_STREAM_EDGE_BEAD_RING_M7,
  COL_STREAM_EDGE_BEAD_RING_M8, COL_STREAM_EDGE_BEAD_RING_M9,
  COL_STREAM_EDGE_BEAD_RING_M10, COL_STREAM_EDGE_BEAD_RING_M11,
  COL_STREAM_EDGE_BEAD_RING_M12, COL_STREAM_EDGE_BEAD_RING_M13,
  COL_STREAM_EDGE_BEAD_RING_M14, COL_STREAM_EDGE_BEAD_RING_M15,
} from "../../../../../Buffer/column-streams-gen";
import { beadStyleForValue } from "./bead-style";
import { getCanonicalBeadRingSurfaceGeometry } from "./bead-ring-surface-geometry";
import { SHADING_PARAM_BEAD_RADIUS } from "../../../../../Buffer/shading-params";

const RING_COLS = [
  COL_STREAM_EDGE_BEAD_RING_M0, COL_STREAM_EDGE_BEAD_RING_M1,
  COL_STREAM_EDGE_BEAD_RING_M2, COL_STREAM_EDGE_BEAD_RING_M3,
  COL_STREAM_EDGE_BEAD_RING_M4, COL_STREAM_EDGE_BEAD_RING_M5,
  COL_STREAM_EDGE_BEAD_RING_M6, COL_STREAM_EDGE_BEAD_RING_M7,
  COL_STREAM_EDGE_BEAD_RING_M8, COL_STREAM_EDGE_BEAD_RING_M9,
  COL_STREAM_EDGE_BEAD_RING_M10, COL_STREAM_EDGE_BEAD_RING_M11,
  COL_STREAM_EDGE_BEAD_RING_M12, COL_STREAM_EDGE_BEAD_RING_M13,
  COL_STREAM_EDGE_BEAD_RING_M14, COL_STREAM_EDGE_BEAD_RING_M15,
];

const RING_COLOR = beadStyleForValue(1)!.ring;

export function ChainBeadInstances({ capacity, onCount }: {
  capacity: number;
  onCount: (n: number) => void;
}) {
  const litBodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
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

    const ringOut = ring.instanceMatrix.array;
    const { nodes } = ownerCounts();

    let seen = 0;
    let litCount = 0;
    for (let row = 0; row < nodes; row++) {
      const xs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_X));
      if (!xs || xs.byteLength === 0) continue;
      const beads = xs.byteLength >> 2;
      seen += beads;

      const ys = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_Y));
      const zs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_Z));
      const vs = columnBytes(nodeColumn(row, COL_STREAM_EDGE_BEAD_VALUE));
      if (!ys || !zs || !vs) continue;
      if (ys.byteLength < beads * 4 || zs.byteLength < beads * 4 || vs.byteLength < beads * 4) continue;
      const rings = RING_COLS.map((c) => columnBytes(nodeColumn(row, c)));

      for (let i = 0; i < beads && litCount < capacity; i++) {
        const o = i * 4;
        const style = beadStyleForValue(vs.getInt32(o, true));
        if (!style) continue;

        matRef.current.makeTranslation(
          xs.getFloat32(o, true), ys.getFloat32(o, true), zs.getFloat32(o, true),
        );
        litBody.setMatrixAt(litCount, matRef.current);
        litBody.setColorAt(litCount, colRef.current.set(style.fill));

        const base = litCount * 16;
        for (let m = 0; m < 16; m++) {
          const col = rings[m];
          ringOut[base + m] = col && col.byteLength >= o + 4 ? col.getFloat32(o, true) : 0;
        }
        ring.setColorAt(litCount, colRef.current.set(RING_COLOR));

        litCount++;
      }
    }
    onCount(seen);

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
