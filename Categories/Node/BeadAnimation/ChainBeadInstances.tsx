import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { ownerCounts } from "../../Scene/owner-counts";
import { beadBytes, BEAD_RING_NAMES } from "../../Ring/Bead/bead-leaves";
import { beadStyleForValue } from "../../Ring/Bead/bead-style";
import { getCanonicalBeadRingSurfaceGeometry } from "../../Ring/Bead/bead-ring-surface-geometry";
import { SHADING_PARAM_BEAD_RADIUS } from "../../Node/nodegeom/shading-params";

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
      const xs = beadBytes(row, "x");
      if (!xs || xs.byteLength === 0) continue;
      const beads = xs.byteLength >> 2;
      seen += beads;

      const ys = beadBytes(row, "y");
      const zs = beadBytes(row, "z");
      const vs = beadBytes(row, "value");
      if (!ys || !zs || !vs) continue;
      if (ys.byteLength < beads * 4 || zs.byteLength < beads * 4 || vs.byteLength < beads * 4) continue;
      const rings = BEAD_RING_NAMES.map((n) => beadBytes(row, n));

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
