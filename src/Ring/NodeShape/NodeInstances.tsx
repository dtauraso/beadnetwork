import { useRef, useContext } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { EnvTexContext } from "../../webview/scene/scene-env";
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
} from "../../Node/nodegeom/shading-params";
import {
  BUFFER_NODE_TAG, BUFFER_RING_TAG,
  RING_PICK_TUBE_RATIO, RING_PICK_COLOR, RING_PICK_OPACITY,
  RING_BAND_MAJOR, RING_BAND_TUBE, HOVER_COLOR, HOVER_RING_TUBE_RATIO,
} from "../../webview/scene/buffer-scene-shared";
import { updateNodeInstances } from "./node-instances-update";
import { getCanonicalRingSurfaceGeometry } from "../../Ring/NodeShape/ring-surface-geometry";
import {
  SELECTION_RING_TUBE_RATIO, SELECTION_RING_RADIAL_SEGMENTS, SELECTION_RING_TUBULAR_SEGMENTS,
  SELECTION_HALO_WIDTH_SEGMENTS, SELECTION_HALO_HEIGHT_SEGMENTS,
  SELECTION_COLOR, SELECTION_EMISSIVE_INTENSITY, HALO_COLOR, HALO_OPACITY,
} from "./node-highlight-shape";

export function NodeInstances({ capacity }: { capacity: number }) {
  const envTex = useContext(EnvTexContext);
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const ringPickRef = useRef<THREE.InstancedMesh>(null);
  const ringBandRef = useRef<THREE.InstancedMesh>(null);
  const selRingRef = useRef<THREE.InstancedMesh>(null);
  const selHaloRef = useRef<THREE.InstancedMesh>(null);
  const hoverRingRef = useRef<THREE.InstancedMesh>(null);
  const matRef  = useRef(new THREE.Matrix4());
  const posRef  = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef  = useRef(new THREE.Vector3());
  const colRef  = useRef(new THREE.Color());
  const ringGeomAppliedRef = useRef(false);

  useFrame(({ camera }) => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    const ringPick = ringPickRef.current;
    const ringBand = ringBandRef.current;
    const selRing = selRingRef.current;
    const selHalo = selHaloRef.current;
    const hoverRing = hoverRingRef.current;
    if (!body || !ring || !ringPick || !ringBand || !selRing || !selHalo || !hoverRing) return;

    if (!ringGeomAppliedRef.current) {
      const geom = getCanonicalRingSurfaceGeometry();
      if (geom) {
        ring.geometry = geom;
        ringGeomAppliedRef.current = true;
      }
    }

    updateNodeInstances(
      {
        body, ring, ringPick, ringBand, selRing, selHalo, hoverRing,
        mat: matRef.current,
        pos: posRef.current,
        quat: quatRef.current,
        scl: sclRef.current,
        col: colRef.current,
      },
      capacity,
      camera,
    );
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
        {}
        <bufferGeometry />
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
      {}
      <instancedMesh ref={selRingRef} args={[undefined, undefined, 1]} frustumCulled={false} raycast={() => null}>
        <torusGeometry args={[1, SELECTION_RING_TUBE_RATIO, SELECTION_RING_RADIAL_SEGMENTS, SELECTION_RING_TUBULAR_SEGMENTS]} />
        <meshStandardMaterial color={SELECTION_COLOR} emissive={SELECTION_COLOR} emissiveIntensity={SELECTION_EMISSIVE_INTENSITY} />
      </instancedMesh>
      <instancedMesh ref={selHaloRef} args={[undefined, undefined, 1]} frustumCulled={false} raycast={() => null}>
        <sphereGeometry args={[1, SELECTION_HALO_WIDTH_SEGMENTS, SELECTION_HALO_HEIGHT_SEGMENTS]} />
        <meshBasicMaterial color={HALO_COLOR} transparent opacity={HALO_OPACITY} side={THREE.DoubleSide} depthWrite={false} />
      </instancedMesh>
      <instancedMesh ref={hoverRingRef} args={[undefined, undefined, 1]} frustumCulled={false} raycast={() => null}>
        <torusGeometry args={[1, HOVER_RING_TUBE_RATIO, 8, 32]} />
        <meshStandardMaterial color={HOVER_COLOR} emissive={HOVER_COLOR} emissiveIntensity={SELECTION_EMISSIVE_INTENSITY} />
      </instancedMesh>
    </>
  );
}
