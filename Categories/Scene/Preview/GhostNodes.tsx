import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { previewF32Run } from "./preview-leaves";
import { nodeF32 } from "../../Node/node-leaves";

const GHOST_COLOR = "#3d9bff";
const GHOST_OPACITY = 0.55;
const RIM_COLOR = "#0a5cd6";
const RIM_OPACITY = 0.9;
const RIM_SCALE = 1.12;
const noRaycast = () => {};

export function GhostNodes({ capacity }: { capacity: number }) {
  const body = useRef<THREE.InstancedMesh>(null);
  const rim = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());

  useFrame(() => {
    const b = body.current;
    const o = rim.current;
    if (!b || !o) return;
    const xs = previewF32Run("ghostX");
    const ys = previewF32Run("ghostY");
    const zs = previewF32Run("ghostZ");
    if (!xs || !ys || !zs) { b.count = 0; o.count = 0; return; }
    const n = Math.min(xs.length, ys.length, zs.length, capacity);
    const r = Math.hypot(nodeF32(0, "bodyM0"), nodeF32(0, "bodyM1"), nodeF32(0, "bodyM2")) || 1;
    for (let i = 0; i < n; i++) {
      pos.current.set(xs[i] ?? 0, ys[i] ?? 0, zs[i] ?? 0);
      scl.current.set(r, r, r);
      mat.current.compose(pos.current, quat.current, scl.current);
      b.setMatrixAt(i, mat.current);
      scl.current.multiplyScalar(RIM_SCALE);
      mat.current.compose(pos.current, quat.current, scl.current);
      o.setMatrixAt(i, mat.current);
    }
    b.count = n;
    o.count = n;
    b.instanceMatrix.needsUpdate = true;
    o.instanceMatrix.needsUpdate = true;
  });

  return (
    <>
      <instancedMesh ref={rim} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={noRaycast} renderOrder={2}>
        <sphereGeometry args={[1, 24, 24]} />
        <meshBasicMaterial color={RIM_COLOR} transparent opacity={RIM_OPACITY} side={THREE.BackSide} depthWrite={false} toneMapped={false} />
      </instancedMesh>
      <instancedMesh ref={body} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={noRaycast} renderOrder={3}>
        <sphereGeometry args={[1, 24, 24]} />
        <meshBasicMaterial color={GHOST_COLOR} transparent opacity={GHOST_OPACITY} depthWrite={false} toneMapped={false} />
      </instancedMesh>
    </>
  );
}
