import * as THREE from "three";
import { polarToCart } from "../../Vectors/polar-convert";

export function anglesToWorldOffset(r: number, phi: number, theta: number): THREE.Vector3 {
  return new THREE.Vector3(...polarToCart(r, phi, theta));
}
