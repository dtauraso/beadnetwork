import React, { useMemo, useEffect } from "react";
import * as THREE from "three";

export function SceneGuides({ center, radius, tube, showTori, showHandholds }: {
  center: THREE.Vector3;
  radius: number;
  tube: number;
  showTori: boolean;
  showHandholds: boolean;
}) {
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

  const hhAngles = [0, Math.PI / 2, Math.PI, (3 * Math.PI) / 2];
  const hhRadius = Math.max(radius * 0.04, 3);
  const handholds = (rotation?: THREE.Euler) => (
    <group rotation={rotation}>
      {hhAngles.map((a) => (
        <mesh key={a} position={[radius * Math.cos(a), radius * Math.sin(a), 0]} userData={{ handhold: true }}>
          <sphereGeometry args={[hhRadius, 16, 16]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} transparent opacity={0.9} />
        </mesh>
      ))}
    </group>
  );

  return (
    <group position={[center.x, center.y, center.z]}>
      {showTori && (
        <>
          <mesh geometry={geoA} raycast={() => null}>
            <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
          </mesh>
          <mesh geometry={geoB} rotation={rotB} raycast={() => null}>
            <meshBasicMaterial color="#cc8844" transparent opacity={0.4} depthWrite={false} />
          </mesh>
        </>
      )}
      {showHandholds && handholds()}
      {showHandholds && handholds(rotB)}
    </group>
  );
}
