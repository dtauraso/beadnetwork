import React, { useMemo } from "react";
import * as THREE from "three";
import { type NavNode } from "../../Node/nav-nodes";
import { overlayFlag } from "../../Overlay/overlay-flags";

const UP = new THREE.Vector3(0, 1, 0);

export function SceneVectors({ center, nodes, tube }: {
  center: THREE.Vector3; nodes: NavNode[]; tube: number;
}) {
  if (!overlayFlag("sceneVectors")) return null;

  return (
    <>
      {nodes.map((n) => (
        <SceneVector key={n.row} from={center} to={n.center} tube={tube} />
      ))}
    </>
  );
}

function SceneVector({ from, to, tube }: {
  from: THREE.Vector3; to: THREE.Vector3; tube: number;
}) {
  const { mid, quat, len } = useMemo(() => {
    const d = to.clone().sub(from);
    const l = d.length();
    const q = new THREE.Quaternion();
    if (l > 0) q.setFromUnitVectors(UP, d.clone().normalize());
    return { mid: from.clone().add(d.multiplyScalar(0.5)), quat: q, len: l };
  }, [from, to]);

  if (len <= 0) return null;
  return (
    <group position={[mid.x, mid.y, mid.z]} quaternion={quat}>
      <mesh raycast={() => null}>
        <cylinderGeometry args={[tube, tube, len, 8]} />
        <meshBasicMaterial color="#a678e0" transparent opacity={0.6} depthWrite={false} />
      </mesh>
    </group>
  );
}
