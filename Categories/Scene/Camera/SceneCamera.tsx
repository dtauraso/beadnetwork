import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { anglesToWorldOffset } from "./viewpoint-bridge";
import { readCameraPose } from "./camera-leaves";

export function SceneCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera, gl } = useThree();
  const pivotRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const cam = camera as THREE.PerspectiveCamera;
    if (cameraRef) cameraRef.current = cam;

    const pose = readCameraPose();
    if (!pose || !(pose.r > 0)) return;

    const pivot = pivotRef.current;
    pivot.set(pose.pivotX, pose.pivotY, pose.pivotZ);
    const posOffset = anglesToWorldOffset(pose.r, pose.posPhi, pose.posTheta);
    cam.position.copy(pivot).add(posOffset);
    const upDir = anglesToWorldOffset(1, pose.upPhi, pose.upTheta).normalize();
    cam.up.copy(upDir);
    cam.lookAt(pivot);

    const el = gl.domElement;
    const widthPx = Math.max(1, el.clientWidth);
    const heightPx = Math.max(1, el.clientHeight);
    if (pose.focalPx > 0) {
      const fov = 2 * Math.atan(heightPx / (2 * pose.focalPx)) * 180 / Math.PI;
      const aspect = widthPx / heightPx;
      if (fov !== cam.fov || aspect !== cam.aspect) {
        cam.fov = fov;
        cam.aspect = aspect;
        cam.updateProjectionMatrix();
      }
    }

    cam.updateMatrixWorld(true);
  });

  return null;
}
