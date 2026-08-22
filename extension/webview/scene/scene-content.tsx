import React, { useEffect, useRef } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import type { PickRef } from "../../../Input/Drag/pick-types";
import {
  SHADING_PARAM_SCENE_AMBIENT_INTENSITY,
  SHADING_PARAM_SCENE_DIR_INTENSITY,
} from "../../../Node/nodegeom/shading-params";
import { SCENE_NODE_TAG, SCENE_EDGE_TAG, SCENE_RING_TAG } from "./scene-root";
import { resolveNodeDrawSlot } from "../../../Ring/NodeShape/node-depth-order";

function pickEdge(hits: THREE.Intersection[]): string | null {
  for (const hit of hits) {
    const row: unknown = (hit.object as THREE.Mesh).userData?.[SCENE_EDGE_TAG];
    if (typeof row !== "number") continue;
    return String(row);
  }
  return null;
}

function pickHandhold(hits: THREE.Intersection[]): string | null {
  for (const hit of hits) {
    const data = (hit.object as THREE.Mesh).userData;
    if (data?.handhold === true) return "1";
  }
  return null;
}

function pickRing(hits: THREE.Intersection[]): string | null {
  for (const hit of hits) {
    if ((hit.object as THREE.Mesh).userData?.[SCENE_RING_TAG] !== true) continue;
    if (hit.instanceId === undefined) continue;
    return String(resolveNodeDrawSlot(hit.instanceId));
  }
  return null;
}

function pickNode(hits: THREE.Intersection[], excludeRow?: string): string | null {
  for (const hit of hits) {
    if ((hit.object as THREE.Mesh).userData?.[SCENE_NODE_TAG] !== true) continue;
    if (hit.instanceId === undefined) continue;
    const row = String(resolveNodeDrawSlot(hit.instanceId));
    if (excludeRow && row === excludeRow) continue;
    return row;
  }
  return null;
}

function RaycasterHelper({
  onPickRequest,
}: {
  onPickRequest: PickRef;
}) {
  const { camera, scene } = useThree();
  const raycaster = useRef(new THREE.Raycaster());

  useEffect(() => {

    onPickRequest.current = (ndcX, ndcY, opts) => {
      const ndc = new THREE.Vector2(ndcX, ndcY);
      raycaster.current.setFromCamera(ndc, camera);
      const allHits = raycaster.current.intersectObject(scene, true);
      const hits =
        allHits.length === 0
          ? allHits
          : allHits.filter((h) => (h.object as THREE.Mesh).isMesh);
      if (hits.length === 0) return null;

      if (opts?.handholdOnly) return pickHandhold(hits);
      if (opts?.edgeOnly) return pickEdge(hits);
      if (opts?.ringOnly) return pickRing(hits);
      return pickNode(hits, opts?.nodesOnly ? opts.excludeId : undefined);
    };

  }, [camera, scene, onPickRequest]);

  return null;
}

export function Scene({
  onPickRequest,
}: {
  onPickRequest: PickRef;
}) {
  return (
    <>
      <RaycasterHelper onPickRequest={onPickRequest} />
      <ambientLight intensity={SHADING_PARAM_SCENE_AMBIENT_INTENSITY} />
      <directionalLight position={[0, 0, 10]} intensity={SHADING_PARAM_SCENE_DIR_INTENSITY} />
    </>
  );
}
