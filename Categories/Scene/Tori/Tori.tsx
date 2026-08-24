import React, { useMemo, useEffect } from "react";
import * as THREE from "three";
import { SHADING_PARAM_TORUS_COLOR, SHADING_PARAM_TORUS_OPACITY } from "./shading-params";
import { overlayFlag } from "../../Overlay/overlay-flags";

export function Tori({ center, radius, tube }: {
  center: THREE.Vector3;
  radius: number;
  tube: number;
}) {
  const visible = overlayFlag("tori");

  const { geoA, geoB } = useMemo(
    () => ({
      geoA: new THREE.TorusGeometry(radius, tube, 12, 96),
      geoB: new THREE.TorusGeometry(radius, tube, 12, 96),
    }),
    [radius, tube],
  );

  useEffect(() => {
    return () => {
      geoA.dispose();
      geoB.dispose();
    };
  }, [geoA, geoB]);

  const rotB = useMemo(() => new THREE.Euler(Math.PI / 2, 0, 0), []);

  if (!visible) return null;

  return (
    <group position={[center.x, center.y, center.z]}>
      <mesh geometry={geoA} raycast={() => null}>
        <meshBasicMaterial color={SHADING_PARAM_TORUS_COLOR} transparent opacity={SHADING_PARAM_TORUS_OPACITY} depthWrite={false} />
      </mesh>
      <mesh geometry={geoB} rotation={rotB} raycast={() => null}>
        <meshBasicMaterial color={SHADING_PARAM_TORUS_COLOR} transparent opacity={SHADING_PARAM_TORUS_OPACITY} depthWrite={false} />
      </mesh>
    </group>
  );
}
