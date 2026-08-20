import { useEffect, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { anglesToWorldOffset } from "./viewpoint-bridge";
import { loadCameraPaths, readCameraPose, type CameraPose, type CameraPrimitive } from "./camera-leaves";

export function BufferCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera, gl } = useThree();
  const pivotRef = useRef(new THREE.Vector3());
  const poseRef = useRef<CameraPose | null>(null);

  useEffect(() => {
    let live = true;
    let paths: Map<CameraPrimitive, string> | undefined;
    const pump = async () => {
      while (live) {
        paths ??= await loadCameraPaths();
        if (paths) {
          const pose = await readCameraPose(paths);
          if (pose) poseRef.current = pose;
        }
        await new Promise((r) => requestAnimationFrame(() => r(undefined)));
      }
    };
    void pump();
    return () => { live = false; };
  }, []);

  useFrame(() => {
    const cam = camera as THREE.PerspectiveCamera;
    if (cameraRef) cameraRef.current = cam;

    const pose = poseRef.current;
    if (!pose || !(pose.r > 0)) return;

    const pivot = pivotRef.current;
    pivot.set(pose["pivot-x"], pose["pivot-y"], pose["pivot-z"]);
    const posOffset = anglesToWorldOffset(pose.r, pose["pos-phi"], pose["pos-theta"]);
    cam.position.copy(pivot).add(posOffset);
    const upDir = anglesToWorldOffset(1, pose["up-phi"], pose["up-theta"]).normalize();
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
