import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeBeads } from "../edges/edge-bead-blocks";
import { beadStyleForValue, COMM_EDGE_LINE_COLOR } from "./bead-style";
import { getCommNodeRows } from "../nodes/comm-nodes";
import { overlayOn } from "../../controls/flags/overlay-flags";
import { readOverlayCommEdges } from "../../../../schema/buffer-layout/buffer-layout";
import {
  SHADING_PARAM_BEAD_RADIUS,
  SHADING_PARAM_BEAD_RING_TUBE_RATIO,
} from "../../../../schema/buffer-layout/shading-params";

const RING_COLOR = beadStyleForValue(1)!.ring;

const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);
const BEAD_UNIT_SCALE = new THREE.Vector3(1, 1, 1);

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const litBodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const ringMatRef = useRef(new THREE.Matrix4());
  const quatRef = useRef(new THREE.Quaternion());
  const axisRef = useRef(new THREE.Vector3());
  const posRef = useRef(new THREE.Vector3());
  const colRef = useRef(new THREE.Color());

  useFrame(() => {
    const litBody = litBodyRef.current;
    const ring = ringRef.current;
    if (!litBody || !ring) return;

    const { positions, ringAxis, value, srcNodeRow, count } = getEdgeBeads();
    const showComm = overlayOn(readOverlayCommEdges);
    const commRows = showComm ? getCommNodeRows() : null;

    const drawn = Math.min(count, capacity);

    let litCount = 0;
    for (let i = 0; i < drawn; i++) {
      // A comm edge is drawn as a line and an arrow, so its beads are not
      // drawn at all while that overlay is on.
      if (commRows !== null && commRows.has(srcNodeRow[i]!)) continue;

      const style = beadStyleForValue(value[i]);
      if (!style) continue;

      posRef.current.set(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      matRef.current.makeTranslation(posRef.current.x, posRef.current.y, posRef.current.z);
      litBody.setMatrixAt(litCount, matRef.current);
      litBody.setColorAt(litCount, colRef.current.set(style.fill));

      // The torus faces along the way the bead is going, which for a bead on
      // an edge is that edge's own direction.
      axisRef.current.set(ringAxis[i * 3]!, ringAxis[i * 3 + 1]!, ringAxis[i * 3 + 2]!);
      quatRef.current.setFromUnitVectors(TORUS_DEFAULT_NORMAL, axisRef.current);
      ringMatRef.current.compose(posRef.current, quatRef.current, BEAD_UNIT_SCALE);
      ring.setMatrixAt(litCount, ringMatRef.current);
      ring.setColorAt(litCount, colRef.current.set(RING_COLOR));

      litCount++;
    }
    litBody.count = litCount;
    ring.count = litCount;
    litBody.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (litBody.instanceColor) litBody.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      <instancedMesh ref={litBodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <torusGeometry
          args={[SHADING_PARAM_BEAD_RADIUS, SHADING_PARAM_BEAD_RADIUS * SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8, 24]}
        />
        <meshBasicMaterial toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
