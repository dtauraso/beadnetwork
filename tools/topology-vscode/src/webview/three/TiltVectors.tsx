// TiltVectors.tsx — one arrow per node, from its centre outward along its OWN TILT
// direction.
//
// The direction is Go-owned: TiltVectorTheta/TiltVectorPhi (Buffer/layout.go), the SAME
// θ-from-world-+y / φ-azimuth-around-+y convention as the ring axis and every other angle
// pair on the buffer. Default (0,0) is world +y, matching the tilt vector's original
// hardcoded-up arrangement — a scene/topology that never sets an angle looks unchanged. Go
// holds the angle as an INTEGER index × nodes/Wiring.CurveParamTiltVectorAngleStep
// (memory/feedback_abc_times_constant_not_rederive.md); this component reads only the
// already-multiplied radians the buffer carries, same as it reads RingAxisTheta/Phi — no
// index/step arithmetic on this side.
//
// WHETHER a node draws one is Go's answer too, and it rides the SAME value as how long:
// TiltVectorLen is zero for a node that draws none. There is no toggle and no per-scene
// branch here — a scene that wants no tilt vectors streams zeros and this component draws
// nothing.
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
  readNodeTiltVectorLen, readNodeTiltVectorTheta, readNodeTiltVectorPhi,
  readNodeVector2Theta, readNodeVector2Phi,
} from "../../schema/buffer-layout";

// The shaft's thickness and the head's size, as fractions of the vector's own length, so an
// arrow keeps its proportions whatever a node's radius is.
const SHAFT_RADIUS_FRAC = 0.035;
const HEAD_LEN_FRAC = 0.22;
const HEAD_RADIUS_FRAC = 0.09;
// Chosen against what the arrow actually crosses, not in isolation: it overlaps its own
// node's body, and the pair's two kinds are both PALE (Node1 #fff8e1, a near-white yellow;
// Node2 #e8eaf6, a near-white blue), while the scene behind them is dark. A light colour —
// the pale green this replaced — vanished against the node bodies exactly where the arrow
// starts. A saturated magenta is far from both pale tints in hue AND much darker than them,
// so it reads on the bodies, and it stays bright enough to read against the dark background
// over the rest of its length.
const VECTOR_COLOR = "#FF2E88";

// three.js authors both a cylinder and a cone along +Y, so that is the axis rotated FROM.
const GEOMETRY_AXIS = new THREE.Vector3(0, 1, 0);

export function TiltVectors({ capacity }: { capacity: number }) {
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

    // Each node draws TWO arrows: its own tilt vector, and a second one a quarter turn
    // away inside the same ring plane (Buffer/layout.go's Vector2Theta/Vector2Phi). Both
    // come from Go as directions; nothing here decides where either points. They are
    // written into the SAME instanced meshes, so one draw call still covers every arrow.
    let drawn = 0;
    const writeArrow = (cx: number, cy: number, cz: number, len: number, theta: number, phi: number) => {
      // (0,0) decodes to world +y — the default the tilt vector had before it carried its
      // own direction, so an unedited node's arrow is unchanged. Same θ-from-+y /
      // φ-azimuth-around-+y conversion NodeInstances uses for the ring axis.
      if (theta === 0 && phi === 0) {
        axisRef.current.set(0, 1, 0);
      } else {
        const st = Math.sin(theta);
        axisRef.current.set(st * Math.cos(phi), Math.cos(theta), st * Math.sin(phi));
      }
      quatRef.current.setFromUnitVectors(GEOMETRY_AXIS, axisRef.current);

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
      // lands the tip exactly at distance `len` from the node's centre.
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
    };

    for (let row = 0; row < nodeCount && drawn + 1 < capacity; row++) {
      const len = readNodeTiltVectorLen(nodeView, row);
      if (!(len > 0)) continue; // Go says this node draws no tilt vectors

      const cx = readNodeCX(nodeView, row);
      const cy = readNodeCY(nodeView, row);
      const cz = readNodeCZ(nodeView, row);

      writeArrow(cx, cy, cz, len, readNodeTiltVectorTheta(nodeView, row), readNodeTiltVectorPhi(nodeView, row));
      writeArrow(cx, cy, cz, len, readNodeVector2Theta(nodeView, row), readNodeVector2Phi(nodeView, row));
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
