import { useMemo, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame, getNodeStreamFrameForRow } from "./node-frame-aggregate";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readNodeRingAxisPhi, readNodeRingAxisTheta, readNodeRingTubeRadius,
  readRingPointX, readRingPointY, readRingPointZ,
} from "../../../../schema/buffer-layout/buffer-layout";
import { SHADING_PARAM_RING_ROUGHNESS } from "../../../../schema/buffer-layout/shading-params";
import { nodeRowColors, poleAxis } from "../buffer-scene-shared";

const TUBE_RADIAL_SEGMENTS = 8;

export function NodeRingTube({ row }: { row: number }) {
  const meshRef = useRef<THREE.Mesh>(null);
  const geometry = useMemo(() => new THREE.BufferGeometry(), []);
  const colorRef = useRef(new THREE.Color());

  useFrame(() => {
    const mesh = meshRef.current;
    if (!mesh) return;

    const decodedNode = getNodeFrame();
    const streamFrame = getNodeStreamFrameForRow(row);
    if (!decodedNode || row >= decodedNode.nodeCount || !streamFrame || streamFrame.ringPointCount === 0) {
      mesh.visible = false;
      return;
    }

    const { nodeView } = decodedNode;
    const { ringPointView, ringPointCount: n } = streamFrame;

    const cx = readNodeCX(nodeView, row);
    const cy = readNodeCY(nodeView, row);
    const cz = readNodeCZ(nodeView, row);
    const a = readNodeRingTubeRadius(nodeView, row);
    if (a <= 0) {
      mesh.visible = false;
      return;
    }

    const [axX, axY, axZ] = poleAxis(readNodeRingAxisPhi(nodeView, row), readNodeRingAxisTheta(nodeView, row));

    const segs = TUBE_RADIAL_SEGMENTS;
    const vertexCount = n * segs;
    const positions = new Float32Array(vertexCount * 3);

    for (let k = 0; k < n; k++) {
      const px = readRingPointX(ringPointView, k); // world-space, streamed directly
      const py = readRingPointY(ringPointView, k);
      const pz = readRingPointZ(ringPointView, k);

      let ux = px - cx, uy = py - cy, uz = pz - cz;
      const uLen = Math.hypot(ux, uy, uz) || 1;
      ux /= uLen; uy /= uLen; uz /= uLen;

      for (let j = 0; j < segs; j++) {
        const angle = (j / segs) * Math.PI * 2;
        const c = Math.cos(angle), s = Math.sin(angle);
        const vx = px + a * (c * ux + s * axX);
        const vy = py + a * (c * uy + s * axY);
        const vz = pz + a * (c * uz + s * axZ);
        const vi = (k * segs + j) * 3;
        positions[vi] = vx;
        positions[vi + 1] = vy;
        positions[vi + 2] = vz;
      }
    }

    const quadCount = n * segs;
    const indices = new Uint32Array(quadCount * 6);
    let ii = 0;
    for (let k = 0; k < n; k++) {
      const kNext = (k + 1) % n;
      for (let j = 0; j < segs; j++) {
        const jNext = (j + 1) % segs;
        const a0 = k * segs + j;
        const a1 = k * segs + jNext;
        const b0 = kNext * segs + j;
        const b1 = kNext * segs + jNext;
        indices[ii++] = a0; indices[ii++] = b0; indices[ii++] = a1;
        indices[ii++] = a1; indices[ii++] = b0; indices[ii++] = b1;
      }
    }

    geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    geometry.setIndex(new THREE.BufferAttribute(indices, 1));
    geometry.computeVertexNormals();
    geometry.computeBoundingSphere();

    const { stroke } = nodeRowColors(nodeView, row);
    const mat = mesh.material as THREE.MeshStandardMaterial;
    mat.color.set(colorRef.current.set(stroke));

    mesh.visible = true;
  });

  return (
    <mesh ref={meshRef} geometry={geometry} frustumCulled={false} visible={false}>
      <meshStandardMaterial roughness={SHADING_PARAM_RING_ROUGHNESS} metalness={0} depthWrite={false} transparent={false} opacity={1} />
    </mesh>
  );
}
