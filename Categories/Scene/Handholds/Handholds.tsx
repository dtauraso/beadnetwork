import React, { useMemo } from "react";
import * as THREE from "three";
import {
  SHADING_PARAM_HANDHOLD_COLOR, SHADING_PARAM_HANDHOLD_EMISSIVE,
  SHADING_PARAM_HANDHOLD_EMISSIVE_INTENSITY, SHADING_PARAM_HANDHOLD_OPACITY,
  SHADING_PARAM_HANDHOLD_RADIUS_RATIO, SHADING_PARAM_HANDHOLD_MIN_RADIUS,
} from "./shading-params";
import { overlayFlag } from "../View/Flags/overlay-flags";

const ANGLES = [0, Math.PI / 2, Math.PI, (3 * Math.PI) / 2];

export function Handholds({ center, radius }: {
  center: THREE.Vector3;
  radius: number;
}) {
  const visible = overlayFlag("handholds");
  const rotB = useMemo(() => new THREE.Euler(Math.PI / 2, 0, 0), []);
  const hhRadius = Math.max(radius * SHADING_PARAM_HANDHOLD_RADIUS_RATIO, SHADING_PARAM_HANDHOLD_MIN_RADIUS);

  const ring = (rotation?: THREE.Euler) => (
    <group rotation={rotation}>
      {ANGLES.map((a) => (
        <mesh key={a} position={[radius * Math.cos(a), radius * Math.sin(a), 0]} userData={{ handhold: true }}>
          <sphereGeometry args={[hhRadius, 16, 16]} />
          <meshStandardMaterial color={SHADING_PARAM_HANDHOLD_COLOR} emissive={SHADING_PARAM_HANDHOLD_EMISSIVE} emissiveIntensity={SHADING_PARAM_HANDHOLD_EMISSIVE_INTENSITY} transparent opacity={SHADING_PARAM_HANDHOLD_OPACITY} />
        </mesh>
      ))}
    </group>
  );

  if (!visible) return null;

  return (
    <group position={[center.x, center.y, center.z]}>
      {ring()}
      {ring(rotB)}
    </group>
  );
}
