import * as THREE from "three";
import { postLog } from "../../log/post";

const FRAMES_AFTER_RESIZE = 10;

const PROBE_WORLD_SPAN = 100;

const across = new THREE.Vector3();
const p = new THREE.Vector3();
const q = new THREE.Vector3();

let framesLeft = 0;
let lastW = 0;
let lastH = 0;
let sequence = 0;

export function probeSceneSizeOnResize(
  cam: THREE.PerspectiveCamera, pivot: THREE.Vector3, focalPx: number,
  viewW: number, viewH: number,
): void {
  if (viewW !== lastW || viewH !== lastH) {
    lastW = viewW;
    lastH = viewH;
    framesLeft = FRAMES_AFTER_RESIZE;
    sequence++;
  }
  if (framesLeft <= 0) return;
  const frame = FRAMES_AFTER_RESIZE - framesLeft;
  framesLeft--;

  cam.getWorldDirection(across).cross(cam.up).normalize().multiplyScalar(PROBE_WORLD_SPAN);
  p.copy(pivot).project(cam);
  q.copy(pivot).add(across).project(cam);
  const px = Math.hypot(((p.x - q.x) * viewW) / 2, ((p.y - q.y) * viewH) / 2);

  postLog("scene-size-after-resize", {
    sequence,
    frame,
    view: `${viewW}x${viewH}`,
    focalPx: focalPx.toFixed(1),
    fovCamera: cam.fov.toFixed(3),
    aspect: cam.aspect.toFixed(4),
    pixelsPerWorldUnit: (px / PROBE_WORLD_SPAN).toFixed(4),
  });
}
