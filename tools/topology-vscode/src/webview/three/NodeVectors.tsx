// NodeVectors.tsx — one arrow per node, from its centre to its own top.
//
// The direction is the node's own RING AXIS (Buffer/layout.go's RingAxisTheta/RingAxisPhi),
// the same axis its torus is drawn on, so the arrow points straight out of the ring's plane
// rather than along some separately-chosen direction. In a scene whose axis is up
// (scene_tabs.go's UpAxis) that means the rings lie flat and the arrows stand up out of them.
//
// WHETHER a node draws one is Go's answer too, and it rides the SAME value as how long:
// VectorLen is zero for a node that draws none. There is no toggle and no per-scene branch
// here — a scene that wants no vectors streams zeros and this component draws nothing.
//
// Two instanced meshes, one draw call each: a thin cylinder for the shaft and a cone for the
// head. Both are authored pointing along +Y (three.js's own cylinder/cone axis) and rotated
// onto each node's axis, the same setFromUnitVectors pattern the rings use.
//
// Pure buffer → GPU: this reads the node frames every frame and writes instance matrices
// imperatively, holding no state of its own.

import React, { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "./node-stream-blocks";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readNodeRingAxisTheta, readNodeRingAxisPhi, readNodeVectorLen,
} from "../../schema/buffer-layout";

// The shaft's thickness and the head's size, as fractions of the vector's own length, so an
// arrow keeps its proportions whatever a node's radius is.
const SHAFT_RADIUS_FRAC = 0.035;
const HEAD_LEN_FRAC = 0.22;
const HEAD_RADIUS_FRAC = 0.09;
const VECTOR_COLOR = "#7CFF9E";

// three.js authors both a cylinder and a cone along +Y, so that is the axis rotated FROM.
const GEOMETRY_AXIS = new THREE.Vector3(0, 1, 0);

export function NodeVectors({ capacity }: { capacity: number }) {
  const shaftRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const posRef = useRef(new THREE.Vector3());
  const axisRef = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const shaft = shaftRef.current;
    const head = headRef.current;
    if (!shaft || !head) return;

    const decoded = getNodeFrame();
    if (!decoded) {
      shaft.count = 0;
      head.count = 0;
      return;
    }
    const { nodeCount, nodeView } = decoded;

    let drawn = 0;
    for (let row = 0; row < nodeCount && drawn < capacity; row++) {
      const len = readNodeVectorLen(nodeView, row);
      if (!(len > 0)) continue; // Go says this node draws no vector

      // Direction = this node's own ring axis. (0,0) is the "no position yet" value, which
      // means world +y — the same reading NodeInstances gives it.
      const theta = readNodeRingAxisTheta(nodeView, row);
      const phi = readNodeRingAxisPhi(nodeView, row);
      if (theta === 0 && phi === 0) {
        axisRef.current.set(0, 1, 0);
      } else {
        const st = Math.sin(theta);
        axisRef.current.set(st * Math.cos(phi), Math.cos(theta), st * Math.sin(phi));
      }
      quatRef.current.setFromUnitVectors(GEOMETRY_AXIS, axisRef.current);

      const cx = readNodeCX(nodeView, row);
      const cy = readNodeCY(nodeView, row);
      const cz = readNodeCZ(nodeView, row);

      // SHAFT: a unit cylinder is centred on its own origin, so it sits at the MIDPOINT of
      // the span it covers — from the node's centre to where the head begins.
      const shaftLen = len * (1 - HEAD_LEN_FRAC);
      posRef.current.set(
        cx + axisRef.current.x * (shaftLen / 2),
        cy + axisRef.current.y * (shaftLen / 2),
        cz + axisRef.current.z * (shaftLen / 2),
      );
      sclRef.current.set(len * SHAFT_RADIUS_FRAC, shaftLen, len * SHAFT_RADIUS_FRAC);
      matRef.current.compose(posRef.current, quatRef.current, sclRef.current);
      shaft.setMatrixAt(drawn, matRef.current);

      // HEAD: likewise centred, so it sits half a head-length back from the tip — which
      // lands the tip exactly at the node's top, at distance `len` from its centre.
      const headLen = len * HEAD_LEN_FRAC;
      const headCentre = len - headLen / 2;
      posRef.current.set(
        cx + axisRef.current.x * headCentre,
        cy + axisRef.current.y * headCentre,
        cz + axisRef.current.z * headCentre,
      );
      sclRef.current.set(len * HEAD_RADIUS_FRAC, headLen, len * HEAD_RADIUS_FRAC);
      matRef.current.compose(posRef.current, quatRef.current, sclRef.current);
      head.setMatrixAt(drawn, matRef.current);

      drawn++;
    }

    shaft.count = drawn;
    head.count = drawn;
    shaft.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
  });

  return (
    <>
      {/* Unit geometries — radius 1, height 1 — scaled per instance above, so one
          geometry serves every node whatever its size. */}
      <instancedMesh ref={shaftRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[1, 1, 1, 12]} />
        <meshBasicMaterial color={VECTOR_COLOR} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[1, 1, 14]} />
        <meshBasicMaterial color={VECTOR_COLOR} />
      </instancedMesh>
    </>
  );
}
