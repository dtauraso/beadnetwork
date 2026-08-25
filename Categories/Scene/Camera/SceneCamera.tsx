import { useEffect, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { anglesToWorldOffset } from "./viewpoint-bridge";
import { postLog } from "../../../Start/extension/webview/log/post";
import { loadCameraBlockPath, readCameraPose, type CameraPose } from "./camera-leaves";

export function SceneCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera, gl } = useThree();
  const pivotRef = useRef(new THREE.Vector3());
  const poseRef = useRef<CameraPose | null>(null);

  useEffect(() => {
    let live = true;
    let blockPath: string | undefined;
    let reads = 0, sumMs = 0, maxMs = 0, windowStart = performance.now();
    const pump = async () => {
      while (live) {
        blockPath ??= await loadCameraBlockPath();
        if (blockPath !== undefined) {
          const t0 = performance.now();
          const pose = await readCameraPose(blockPath);
          const ms = performance.now() - t0;
          reads++; sumMs += ms; if (ms > maxMs) maxMs = ms;
          if (pose) poseRef.current = pose;
        }
        const now = performance.now();
        if (now - windowStart >= 1000) {
          postLog("camera-pose-reads", {
            reads,
            meanMs: reads ? (sumMs / reads).toFixed(1) : "0",
            maxMs: maxMs.toFixed(1),
          });
          reads = 0; sumMs = 0; maxMs = 0; windowStart = now;
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
