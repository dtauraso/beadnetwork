import { useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { ownerCounts } from "../../Scene/owner-counts";
import { nodeBytes } from "../node-leaves";

const TICK_COLOR = "#8a8a8a";

const COLUMNS = ["tickX0", "tickY0", "tickZ0", "tickX1", "tickY1", "tickZ1"] as const;

export function TickMarks() {
  const { geometry, holder } = useMemo(() => {
    const g = new THREE.BufferGeometry();
    return { geometry: g, holder: { positions: new Float32Array(0) } };
  }, []);

  useFrame(() => {
    const { nodes } = ownerCounts();
    let total = 0;
    for (let row = 0; row < nodes; row++) {
      const col = nodeBytes(row, "tickX0");
      if (col) total += col.byteLength / 4;
    }

    if (holder.positions.length < total * 6) {
      holder.positions = new Float32Array(Math.ceil(total * 1.5) * 6);
      geometry.setAttribute("position", new THREE.BufferAttribute(holder.positions, 3));
    }

    let out = 0;
    for (let row = 0; row < nodes; row++) {
      const cols = COLUMNS.map((n) => nodeBytes(row, n));
      const count = cols[0] ? cols[0].byteLength / 4 : 0;
      for (let i = 0; i < count; i++) {
        for (let c = 0; c < 6; c++) {
          const col = cols[c];
          holder.positions[out++] = col && col.byteLength >= i * 4 + 4 ? col.getFloat32(i * 4, true) : 0;
        }
      }
    }

    const attr = geometry.getAttribute("position");
    if (attr) attr.needsUpdate = true;
    geometry.setDrawRange(0, out / 3);
  });

  return (
    <lineSegments geometry={geometry} frustumCulled={false} raycast={() => null}>
      <lineBasicMaterial color={TICK_COLOR} />
    </lineSegments>
  );
}
