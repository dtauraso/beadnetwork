import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { anglesToWorldOffset } from "../nav/viewpoint-bridge";
import { columnF32 } from "../../../../Buffer/column-values";
import {
  COL_STREAM_CAMERA_PX, COL_STREAM_CAMERA_PY, COL_STREAM_CAMERA_PZ, COL_STREAM_CAMERA_R,
  COL_STREAM_CAMERA_POS_PHI, COL_STREAM_CAMERA_POS_THETA,
  COL_STREAM_CAMERA_UP_PHI, COL_STREAM_CAMERA_UP_THETA,
  COL_STREAM_CAMERA_FOCAL_PX,
} from "../../../../Buffer/column-streams-gen";
import { probeSceneSizeOnResize } from "./scene-size-probe";

export function BufferCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera, gl } = useThree();
  const pivotRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const cam = camera as THREE.PerspectiveCamera;
    if (cameraRef) cameraRef.current = cam;

    const r = columnF32(COL_STREAM_CAMERA_R);

    if (!(r > 0)) return;

    const pivot = pivotRef.current;
    pivot.set(
      columnF32(COL_STREAM_CAMERA_PX),
      columnF32(COL_STREAM_CAMERA_PY),
      columnF32(COL_STREAM_CAMERA_PZ),
    );
    const posOffset = anglesToWorldOffset(
      r, columnF32(COL_STREAM_CAMERA_POS_PHI), columnF32(COL_STREAM_CAMERA_POS_THETA));
    cam.position.copy(pivot).add(posOffset);
    const upDir = anglesToWorldOffset(
      1, columnF32(COL_STREAM_CAMERA_UP_PHI), columnF32(COL_STREAM_CAMERA_UP_THETA)).normalize();
    cam.up.copy(upDir);
    cam.lookAt(pivot);

    const focalPx = columnF32(COL_STREAM_CAMERA_FOCAL_PX);
    const el = gl.domElement;
    const widthPx = Math.max(1, el.clientWidth);
    const heightPx = Math.max(1, el.clientHeight);
    if (focalPx > 0) {
      const fov = 2 * Math.atan(heightPx / (2 * focalPx)) * 180 / Math.PI;
      const aspect = widthPx / heightPx;
      if (fov !== cam.fov || aspect !== cam.aspect) {
        cam.fov = fov;
        cam.aspect = aspect;
        cam.updateProjectionMatrix();
      }
    }

    cam.updateMatrixWorld(true);

    probeSceneSizeOnResize(cam, pivot, focalPx, widthPx, heightPx);
  });

  return null;
}
