import { useRef, useContext } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { EnvTexContext } from "../../../Start/extension/webview/scene/scene-env";
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
} from "./shading-params";
import { SCENE_NODE_TAG, SCENE_RING_TAG } from "../../../Start/extension/webview/scene/scene-tags";
import { updateNodeInstances } from "./node-instances-update";
import { getCanonicalRingSurfaceGeometry } from "../../Ring/NodeShape/ring-surface-geometry";

export function NodeInstances({ capacity }: { capacity: number }) {
  const envTex = useContext(EnvTexContext);
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const colRef  = useRef(new THREE.Color());
  const ringGeomAppliedRef = useRef(false);

  useFrame(({ camera }) => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    if (!body || !ring) return;

    if (!ringGeomAppliedRef.current) {
      const geom = getCanonicalRingSurfaceGeometry();
      if (geom) {
        ring.geometry = geom;
        ringGeomAppliedRef.current = true;
      }
    }

    updateNodeInstances({ body, ring, col: colRef.current }, capacity, camera);
  });

  return (
    <>
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} userData={{ [SCENE_NODE_TAG]: true }} frustumCulled={false}>
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
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} userData={{ [SCENE_RING_TAG]: true }} frustumCulled={false}>
        {}
        <bufferGeometry />
        <meshStandardMaterial roughness={SHADING_PARAM_RING_ROUGHNESS} metalness={0} depthWrite={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
