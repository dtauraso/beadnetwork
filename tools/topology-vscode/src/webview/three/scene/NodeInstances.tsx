import { useRef, useContext } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "./node-frame-aggregate";
import { getViewBlocks } from "./view-blocks";
import { EnvTexContext } from "./scene-env";
import {
  SHADING_PARAM_NODE_TRANSMISSION,
  SHADING_PARAM_NODE_THICKNESS,
  SHADING_PARAM_NODE_ROUGHNESS,
  SHADING_PARAM_NODE_IOR,
  SHADING_PARAM_NODE_METALNESS,
  SHADING_PARAM_NODE_CLEARCOAT,
  SHADING_PARAM_NODE_CLEARCOAT_ROUGHNESS,
  SHADING_PARAM_NODE_ENV_MAP_INTENSITY,
  SHADING_PARAM_NODE_OPACITY,
  SHADING_PARAM_RING_ROUGHNESS,
} from "../../../schema/shading-params";
import {
  readNodeRingAxisTheta,
  readNodeRingAxisPhi,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOverlaySelSpherePoles,
  readOverlayNodeBody, readOverlayNodeRing, readOverlayRingPick,
} from "../../../schema/buffer-layout";
import {
  BUFFER_NODE_TAG, BUFFER_RING_TAG, NODE_SPHERE_RADIUS,
  NODE_RING_TUBE_RATIO, RING_PICK_TUBE_RATIO, RING_PICK_COLOR, RING_PICK_OPACITY,
  RING_BAND_MAJOR, RING_BAND_TUBE, nodeRowColors, poleAxis,
} from "./buffer-scene-shared";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";

const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

export function NodeInstances({ capacity }: { capacity: number }) {
  const envTex = useContext(EnvTexContext);
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const ringPickRef = useRef<THREE.InstancedMesh>(null);
  const ringBandRef = useRef<THREE.InstancedMesh>(null);
  const matRef  = useRef(new THREE.Matrix4());
  const posRef  = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const ringQuatRef = useRef(new THREE.Quaternion());
  const ringAxisRef = useRef(new THREE.Vector3());
  const sclRef  = useRef(new THREE.Vector3());
  const colRef  = useRef(new THREE.Color());

  useFrame(({ camera }) => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    const ringPick = ringPickRef.current;
    const ringBand = ringBandRef.current;
    if (!body || !ring || !ringPick || !ringBand) return;

    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!decodedNode || !blocks) {
      body.count = 0; ring.count = 0; ringPick.count = 0; ringBand.count = 0;
      return;
    }
    const { overlayView } = blocks;
    const { nodeCount, nodeView } = decodedNode;

    const selectModeOn = readOverlaySelSpherePoles(overlayView) !== 0;

    const showBody = readOverlayNodeBody(overlayView) !== 0;
    const showRing = readOverlayNodeRing(overlayView) !== 0;

    const showPickBand = readOverlayRingPick(overlayView) !== 0;

    const n = Math.min(nodeCount, capacity);

    const q = quatRef.current;

    const order = computeNodeDepthOrder(
      n,
      (row) => readNodeCX(nodeView, row),
      (row) => readNodeCY(nodeView, row),
      (row) => readNodeCZ(nodeView, row),
      camera.position.x, camera.position.y, camera.position.z,
    );
    setNodeDrawOrder(order);
    const ringAxis = ringAxisRef.current;
    for (let slot = 0; slot < n; slot++) {
      const row = order[slot]!;
      const r = readNodeRadius(nodeView, row) || NODE_SPHERE_RADIUS;
      posRef.current.set(
        readNodeCX(nodeView, row),
        readNodeCY(nodeView, row),
        readNodeCZ(nodeView, row),
      );

      sclRef.current.setScalar(r);
      matRef.current.compose(posRef.current, q, sclRef.current);
      body.setMatrixAt(slot, matRef.current);

      const poleTheta = readNodeRingAxisTheta(nodeView, row);
      const polePhi = readNodeRingAxisPhi(nodeView, row);
      const [ax, ay, az] = poleAxis(poleTheta, polePhi);
      ringAxis.set(ax, ay, az);
      ringQuatRef.current.setFromUnitVectors(TORUS_DEFAULT_NORMAL, ringAxis);
      matRef.current.compose(posRef.current, ringQuatRef.current, sclRef.current);
      ring.setMatrixAt(slot, matRef.current);

      ringPick.setMatrixAt(slot, matRef.current);
      ringBand.setMatrixAt(slot, matRef.current);

      const { fill, stroke } = nodeRowColors(nodeView, row);
      body.setColorAt(slot, colRef.current.set(fill));
      ring.setColorAt(slot, colRef.current.set(stroke));
    }
    body.count = showBody ? n : 0;
    ring.count = showRing ? n : 0;

    ringPick.count = selectModeOn ? n : 0;

    ringBand.count = showPickBand ? n : 0;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    ringPick.instanceMatrix.needsUpdate = true;
    ringBand.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;

    if (showBody) body.computeBoundingSphere();
    if (selectModeOn) ringPick.computeBoundingSphere();
  });

  return (
    <>
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} userData={{ [BUFFER_NODE_TAG]: true }} frustumCulled={false}>
        <sphereGeometry args={[1, 16, 16]} />
        {}
        <meshPhysicalMaterial
          transmission={SHADING_PARAM_NODE_TRANSMISSION}
          thickness={SHADING_PARAM_NODE_THICKNESS}
          roughness={SHADING_PARAM_NODE_ROUGHNESS}
          ior={SHADING_PARAM_NODE_IOR}
          metalness={SHADING_PARAM_NODE_METALNESS}
          clearcoat={SHADING_PARAM_NODE_CLEARCOAT}
          clearcoatRoughness={SHADING_PARAM_NODE_CLEARCOAT_ROUGHNESS}
          envMap={envTex ?? undefined}
          envMapIntensity={SHADING_PARAM_NODE_ENV_MAP_INTENSITY}
          transparent
          opacity={SHADING_PARAM_NODE_OPACITY}
          depthWrite={false}
        />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} userData={{ [BUFFER_RING_TAG]: true }} frustumCulled={false}>
        <torusGeometry args={[1, NODE_RING_TUBE_RATIO, 8, 32]} />
        <meshStandardMaterial roughness={SHADING_PARAM_RING_ROUGHNESS} metalness={0} depthWrite={false} transparent={false} opacity={1} />
      </instancedMesh>
      {}
      <instancedMesh ref={ringPickRef} args={[undefined, undefined, capacity]} userData={{ [BUFFER_RING_TAG]: true }} frustumCulled={false}>
        <torusGeometry args={[1, RING_PICK_TUBE_RATIO, 8, 32]} />
        <meshBasicMaterial transparent opacity={0} colorWrite={false} depthWrite={false} />
      </instancedMesh>
      {}
      <instancedMesh ref={ringBandRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <torusGeometry args={[RING_BAND_MAJOR, RING_BAND_TUBE, 8, 32]} />
        <meshBasicMaterial
          color={RING_PICK_COLOR}
          transparent
          opacity={RING_PICK_OPACITY}
          depthWrite={false}
        />
      </instancedMesh>
    </>
  );
}
