import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { anglesToWorldOffset } from "../nav/viewpoint-bridge";
import { columnF32 } from "../../../../Buffer/column-values";
import {
  COL_STREAM_CAMERA_PX, COL_STREAM_CAMERA_PY, COL_STREAM_CAMERA_PZ, COL_STREAM_CAMERA_R,
  COL_STREAM_CAMERA_POS_PHI, COL_STREAM_CAMERA_POS_THETA,
  COL_STREAM_CAMERA_UP_PHI, COL_STREAM_CAMERA_UP_THETA,
} from "../../../../Buffer/column-streams-gen";

export function BufferCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera } = useThree();
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
    cam.updateMatrixWorld(true);
  });

  return null;
}
