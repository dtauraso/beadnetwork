import * as THREE from "three";
import { polarToCart } from "../polar-convert";

export function anglesToWorldOffset(r: number, theta: number, phi: number): THREE.Vector3 {
  return new THREE.Vector3(...polarToCart(r, theta, phi));
}
