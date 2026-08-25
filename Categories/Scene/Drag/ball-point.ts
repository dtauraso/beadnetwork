import * as THREE from "three";
import { sceneValue } from "../scene-leaves";
import {
  SHADING_PARAM_HANDHOLD_RADIUS_RATIO, SHADING_PARAM_HANDHOLD_MIN_RADIUS,
} from "../Handholds/shading-params";

const eye = new THREE.Vector3();
const dir = new THREE.Vector3();
const centre = new THREE.Vector3();
const toCentre = new THREE.Vector3();
const at = new THREE.Vector3();

const BALL_MAX_EYE_FRACTION = 0.9;

export function ballPoint(
  cam: THREE.PerspectiveCamera,
  ndcX: number,
  ndcY: number,
  keepToRim = false,
): { x: number; y: number; z: number; onRim: boolean } {
  const sceneR = sceneValue("radius");
  if (!(sceneR > 0)) return { x: 0, y: 0, z: 0, onRim: false };
  centre.set(sceneValue("cx"), sceneValue("cy"), sceneValue("cz"));

  eye.setFromMatrixPosition(cam.matrixWorld);
  dir.set(ndcX, ndcY, 0.5).unproject(cam).sub(eye).normalize();

  toCentre.copy(centre).sub(eye);

  const r = Math.min(sceneR, toCentre.length() * BALL_MAX_EYE_FRACTION);
  if (!(r > 0)) return { x: 0, y: 0, z: 0, onRim: false };

  const along = toCentre.dot(dir);
  const perp = Math.sqrt(Math.max(0, toCentre.lengthSq() - along * along));

  const grip = Math.max(r * SHADING_PARAM_HANDHOLD_RADIUS_RATIO, SHADING_PARAM_HANDHOLD_MIN_RADIUS);

  if (!keepToRim && perp < r - grip) {
    at.copy(eye).addScaledVector(dir, along - Math.sqrt(r * r - perp * perp));
    return { x: at.x, y: at.y, z: at.z, onRim: false };
  }
  at.copy(eye).addScaledVector(dir, along).sub(centre).setLength(r).add(centre);
  return { x: at.x, y: at.y, z: at.z, onRim: true };
}
