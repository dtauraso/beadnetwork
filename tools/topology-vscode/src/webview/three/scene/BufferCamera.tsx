import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { getViewBlocks } from "./view-blocks";
import { anglesToWorldOffset } from "../nav/viewpoint-bridge";
import {
  readCameraPX, readCameraPY, readCameraPZ, readCameraR,
  readCameraPosPhi, readCameraPosTheta, readCameraUpPhi, readCameraUpTheta,
} from "../../../../../../Buffer/buffer-layout";

export function BufferCamera({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
}) {
  const { camera } = useThree();
  const pivotRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const cam = camera as THREE.PerspectiveCamera;
    if (cameraRef) cameraRef.current = cam; 

    const blocks = getViewBlocks();
    if (!blocks) return;
    const cv = blocks.cameraView;

    const r = readCameraR(cv);

    if (!(r > 0)) return;

    const pivot = pivotRef.current;
    pivot.set(readCameraPX(cv), readCameraPY(cv), readCameraPZ(cv));
    const posOffset = anglesToWorldOffset(r, readCameraPosPhi(cv), readCameraPosTheta(cv));
    cam.position.copy(pivot).add(posOffset);
    const upDir = anglesToWorldOffset(1, readCameraUpPhi(cv), readCameraUpTheta(cv)).normalize();
    cam.up.copy(upDir);
    cam.lookAt(pivot);
    cam.updateMatrixWorld(true);
  });

  return null;
}
