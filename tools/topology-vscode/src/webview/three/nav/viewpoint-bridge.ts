import * as THREE from "three";

export function anglesToWorldOffset(r: number, theta: number, phi: number): THREE.Vector3 {
  const sinTheta = Math.sin(theta);
  return new THREE.Vector3(
    r * sinTheta * Math.cos(phi),
    r * Math.cos(theta),
    r * sinTheta * Math.sin(phi),
  );
}
