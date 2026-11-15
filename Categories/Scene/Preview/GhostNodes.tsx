import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { previewF32Run } from "./preview-leaves";
import { nodeF32 } from "../../Node/node-leaves";

const GHOST_COLOR = "#9fd8ff";
const GHOST_OPACITY = 0.3;
const noRaycast = () => {};

export function GhostNodes({ capacity }: { capacity: number }) {
  const ref = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());

  useFrame(() => {
    const mesh = ref.current;
    if (!mesh) return;
    const xs = previewF32Run("ghostX");
    const ys = previewF32Run("ghostY");
    const zs = previewF32Run("ghostZ");
    if (!xs || !ys || !zs) { mesh.count = 0; return; }
    const n = Math.min(xs.length, ys.length, zs.length, capacity);
    const r = Math.hypot(nodeF32(0, "bodyM0"), nodeF32(0, "bodyM1"), nodeF32(0, "bodyM2")) || 1;
    scl.current.set(r, r, r);
    for (let i = 0; i < n; i++) {
      pos.current.set(xs[i] ?? 0, ys[i] ?? 0, zs[i] ?? 0);
      mat.current.compose(pos.current, quat.current, scl.current);
      mesh.setMatrixAt(i, mat.current);
    }
    mesh.count = n;
    mesh.instanceMatrix.needsUpdate = true;
  });

  return (
    <instancedMesh ref={ref} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={noRaycast} renderOrder={2}>
      <sphereGeometry args={[1, 16, 16]} />
      <meshBasicMaterial color={GHOST_COLOR} transparent opacity={GHOST_OPACITY} depthWrite={false} toneMapped={false} />
    </instancedMesh>
  );
}
