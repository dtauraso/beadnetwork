// TiltVectors.tsx — up to three arrows per node, from its centre outward: its OWN TILT
// direction, the coplanar normal a quarter turn from it, and (a THIRD, separately
// coloured arrow) the direction that last ARRIVED on this node's tilt-vector channel.
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
// TiltVectorLen is zero for a node that draws none, and ReceivedVectorLen is zero for a
// node that has received nothing on its vector channel yet (or was reset) — see that
// column's own doc comment in Buffer/layout.go for why zero-length is distinguishable
// from an actually-received (0,0) direction.
//
// Two instanced-mesh PAIRS: the first (shaft/head) draws BOTH the tilt vector and the
// coplanar normal, sharing one material colour, at up to two instances per node. The
// second (receivedShaft/receivedHead) draws only the third, received-direction arrow, in
// its OWN colour, at up to one instance per node — kept as a separate mesh pair rather
// than per-instance colours on the first, since only the first pair ever needs more than
// one colour and per-instance colour buffers would size (and dirty) that state for every
// arrow just to serve this one case.
//
// Both authored pointing along +Y (three.js's own cylinder/cone axis) and rotated onto
// each node's axis, the same setFromUnitVectors pattern the rings use.
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
  readNodeCoplanarNormalTheta, readNodeCoplanarNormalPhi,
  readNodeReceivedVectorLen, readNodeReceivedVectorTheta, readNodeReceivedVectorPhi,
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
// The THIRD arrow's colour — the last-received direction, kept by the RECEIVING node
// until the next arrival replaces it. Chosen by the same test as VECTOR_COLOR above: what
// it actually crosses, not isolation. It overlaps the same two pale, near-white node
// bodies (Node1 #fff8e1, Node2 #e8eaf6) against the same dark background, and it must
// also read as visually DISTINCT from the magenta the other two arrows already share on
// this same node. A saturated cyan is on the opposite side of the hue wheel from magenta
// (as far as two saturated colours can be), so the two are never confusable even
// overlapping at a node's centre; it is just as dark relative to the pale bodies as the
// magenta is, so it reads there too; and it stays bright against the dark background over
// the rest of its length, for the same reason the magenta does.
const RECEIVED_VECTOR_COLOR = "#00E5FF";

// three.js authors both a cylinder and a cone along +Y, so that is the axis rotated FROM.
const GEOMETRY_AXIS = new THREE.Vector3(0, 1, 0);

export function TiltVectors({ capacity, receivedCapacity }: { capacity: number; receivedCapacity: number }) {
  const shaftRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const receivedShaftRef = useRef<THREE.InstancedMesh>(null);
  const receivedHeadRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const posRef = useRef(new THREE.Vector3());
  const axisRef = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const shaft = shaftRef.current;
    const head = headRef.current;
    const receivedShaft = receivedShaftRef.current;
    const receivedHead = receivedHeadRef.current;
    if (!shaft || !head || !receivedShaft || !receivedHead) return;

    const decoded = getNodeFrame();
    if (!decoded) {
      shaft.count = 0;
      head.count = 0;
      receivedShaft.count = 0;
      receivedHead.count = 0;
      return;
    }
    const { nodeCount, nodeView } = decoded;

    // writeArrowInto composes one arrow's shaft+head matrices into whichever mesh pair
    // and instance index the caller supplies — shared by both the tilt/normal pair below
    // and the third, received-direction pair, so the geometry math lives in one place.
    const writeArrowInto = (
      targetShaft: THREE.InstancedMesh, targetHead: THREE.InstancedMesh, idx: number,
      cx: number, cy: number, cz: number, len: number, theta: number, phi: number,
    ) => {
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
      targetShaft.setMatrixAt(idx, matRef.current);

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
      targetHead.setMatrixAt(idx, matRef.current);
    };

    // Each node draws up to TWO arrows in the first pair: its own tilt vector, and a
    // second one a quarter turn away inside the same ring plane
    // (Buffer/layout.go's CoplanarNormalTheta/CoplanarNormalPhi). Both come from Go as
    // directions; nothing here decides where either points. They are written into the
    // SAME instanced meshes, so one draw call still covers both.
    let drawn = 0;
    // The THIRD arrow — the last-received direction — draws at most ONE per node, into
    // its own separate mesh pair so it can carry its own colour.
    let receivedDrawn = 0;

    for (let row = 0; row < nodeCount; row++) {
      const cx = readNodeCX(nodeView, row);
      const cy = readNodeCY(nodeView, row);
      const cz = readNodeCZ(nodeView, row);

      const len = readNodeTiltVectorLen(nodeView, row);
      if (len > 0 && drawn + 1 < capacity) {
        writeArrowInto(shaft, head, drawn, cx, cy, cz, len, readNodeTiltVectorTheta(nodeView, row), readNodeTiltVectorPhi(nodeView, row));
        drawn++;
        writeArrowInto(shaft, head, drawn, cx, cy, cz, len, readNodeCoplanarNormalTheta(nodeView, row), readNodeCoplanarNormalPhi(nodeView, row));
        drawn++;
      }

      // 0 = nothing received yet on this node's vector channel (or a reset cleared it) —
      // see ReceivedVectorLen's own doc comment (Buffer/layout.go) for why zero-length is
      // distinguishable from an actually-received (0,0) direction.
      const receivedLen = readNodeReceivedVectorLen(nodeView, row);
      if (receivedLen > 0 && receivedDrawn < receivedCapacity) {
        writeArrowInto(
          receivedShaft, receivedHead, receivedDrawn,
          cx, cy, cz, receivedLen,
          readNodeReceivedVectorTheta(nodeView, row), readNodeReceivedVectorPhi(nodeView, row),
        );
        receivedDrawn++;
      }
    }

    shaft.count = drawn;
    head.count = drawn;
    shaft.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    receivedShaft.count = receivedDrawn;
    receivedHead.count = receivedDrawn;
    receivedShaft.instanceMatrix.needsUpdate = true;
    receivedHead.instanceMatrix.needsUpdate = true;
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
      {/* The THIRD arrow's own mesh pair, in RECEIVED_VECTOR_COLOR, at most one instance
          per node. */}
      <instancedMesh ref={receivedShaftRef} args={[undefined, undefined, receivedCapacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[1, 1, 1, 12]} />
        <meshBasicMaterial color={RECEIVED_VECTOR_COLOR} />
      </instancedMesh>
      <instancedMesh ref={receivedHeadRef} args={[undefined, undefined, receivedCapacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[1, 1, 14]} />
        <meshBasicMaterial color={RECEIVED_VECTOR_COLOR} />
      </instancedMesh>
    </>
  );
}
