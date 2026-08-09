// EdgeLines.tsx — the edge, drawn: one thin line per edge row from the source node's
// surface to the target's, with an ARROWHEAD at the target end saying which way the edge
// runs. Pure buffer→GPU render, no state authority.
//
// This restores a drawn edge. The chain of placeholder beads used to BE the edge's picture
// (docs/beads-are-the-edge.md); it no longer is — the chain still exists in Go, it is simply
// not what you see. What renders a traversal now is a single pulse bead moving along this
// line (ChainBeadInstances), which is why the line and the pulse share one colour: they are
// the same edge, one at rest and one carrying a value.
//
// TS COMPUTES NO GEOMETRY here. The two endpoints are Go's (the Edge block's SX..EZ, the
// edgeMover's own recomputed segment); this file orients a unit cylinder and a cone onto
// them, the same "plot what Go streams" shape SphereRings uses for its streamed normals.

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getEdgeStreamAccessor } from "./edge-stream-blocks";
import { EDGE_LINE_COLOR } from "./bead-style";
// Shading VALUES live in Go (nodes/Wiring/shading_params.go, guard: check-ts-shading-from-go).
// The line reuses the node ring's own roughness rather than introducing a second one: it is
// the same kind of surface, and a literal here would be exactly the drift that guard exists
// to stop.
import {
  SHADING_PARAM_BEAD_RADIUS,
  SHADING_PARAM_BEAD_RING_TUBE_RATIO,
} from "../../schema/shading-params";
import { beadStyleForValue } from "./bead-style";

// The chain bead's ring colour, taken the way ChainBeadInstances takes it: off the value-1
// style rather than as a second black literal.
const BEAD_RING_COLOR = beadStyleForValue(1)!.ring;
import { DIRECTION_ZERO_EPS } from "./buffer-scene-shared";

// Line thickness and arrowhead size in world units. The head is sized off the line so the
// two stay in proportion if the line changes.
// The ORIGINAL drawn edge's own numbers, recovered from the deleted EdgeTube rather than
// re-picked by eye: core tube radius 1.5, arrowhead radius 3. An edge that comes back should
// come back the size it was.
const EDGE_LINE_RADIUS = 1.5;
const ARROW_HEAD_RADIUS = 3;
const ARROW_HEAD_LENGTH = ARROW_HEAD_RADIUS * 2;

// A three.js cylinder and cone are both authored along +Y, so orienting either one means
// rotating THIS onto the segment's own direction.
const AXIS_DEFAULT = new THREE.Vector3(0, 1, 0);
// A torusGeometry lies in the XY plane, so ITS own normal is +Z — a different default from
// the cylinder/cone above, which is why both constants exist rather than one.
const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

export function EdgeLines({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const beadRef = useRef<THREE.InstancedMesh>(null);
  const beadRingRef = useRef<THREE.InstancedMesh>(null);
  const ringQuat = useRef(new THREE.Quaternion());
  const unit = useRef(new THREE.Vector3(1, 1, 1));
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const dir = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    const bead = beadRef.current;
    const beadRing = beadRingRef.current;
    if (!line || !head || !bead || !beadRing) return;

    const edges = getEdgeStreamAccessor();
    if (!edges) {
      line.count = 0; head.count = 0; bead.count = 0; beadRing.count = 0;
      return;
    }

    const n = Math.min(edges.edgeCount, capacity);
    let drawn = 0;
    for (let row = 0; row < n; row++) {
      const [sx, sy, sz, ex, ey, ez] = edges.segment(row);
      dir.current.set(ex - sx, ey - sy, ez - sz);
      const len = dir.current.length();
      // An edge row whose stream frame has not arrived yet reads as 0,0,0 -> 0,0,0. That is
      // a zero-length segment, not an edge at the origin: skip it rather than drawing a
      // degenerate stub, and never normalize it (divide by ~0).
      if (len <= DIRECTION_ZERO_EPS) continue;
      dir.current.divideScalar(len);
      quat.current.setFromUnitVectors(AXIS_DEFAULT, dir.current);

      // The LINE: a unit cylinder (height 1, centred on its own origin) scaled to the
      // segment's length and placed at its midpoint. It stops short of the target end by
      // the head's length so the arrowhead is the tip of the edge rather than an ornament
      // stuck through it.
      const shaft = Math.max(len - ARROW_HEAD_LENGTH, 0);
      pos.current.set(sx, sy, sz).addScaledVector(dir.current, shaft / 2);
      scl.current.set(1, shaft, 1);
      mat.current.compose(pos.current, quat.current, scl.current);
      line.setMatrixAt(drawn, mat.current);

      // The HEAD, at the TARGET end and pointing into it — which end this sits on is the
      // whole point of drawing it, so it is placed from the segment's own direction rather
      // than from a per-edge flag that could disagree with the geometry.
      pos.current.set(ex, ey, ez).addScaledVector(dir.current, -ARROW_HEAD_LENGTH / 2);
      scl.current.set(1, 1, 1);
      mat.current.compose(pos.current, quat.current, scl.current);
      head.setMatrixAt(drawn, mat.current);

      // ONE RESTING EDGE BEAD, at the segment's midpoint — a chain bead, in the chain bead's
      // own fill, sitting between the two nodes. Every other bead on screen is a pulse in a
      // value colour (black/white), so this is the only place the resting tone appears
      // besides the line, which is what makes the line's colour judgeable against it: same
      // scene, same lighting, side by side.
      pos.current.set(sx, sy, sz).addScaledVector(dir.current, len / 2);
      mat.current.makeTranslation(pos.current.x, pos.current.y, pos.current.z);
      bead.setMatrixAt(drawn, mat.current);
      // …and its RING. A chain bead is a sphere INSIDE a torus, never a bare sphere — the
      // ring is most of what it looks like. A torusGeometry's own normal is +Z, so it is
      // rotated onto the edge direction here, which puts the ring around the line the bead
      // sits on (on a real chain the same ring lies in its node's ring plane, resolved from
      // the streamed per-bead axis this reference bead has no equivalent of).
      ringQuat.current.setFromUnitVectors(TORUS_DEFAULT_NORMAL, dir.current);
      mat.current.compose(pos.current, ringQuat.current, unit.current);
      beadRing.setMatrixAt(drawn, mat.current);

      drawn++;
    }
    line.count = drawn;
    head.count = drawn;
    bead.count = drawn;
    beadRing.count = drawn;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
    bead.instanceMatrix.needsUpdate = true;
    beadRing.instanceMatrix.needsUpdate = true;
    if (drawn > 0) {
      line.computeBoundingSphere();
      head.computeBoundingSphere();
    }
  });

  return (
    <>
      {/* raycast disabled on both: the edge is not a pick target — the pick halo that once
          made it one was removed with the old EdgeTube, and nothing on the bridge can
          receive an edge hit any more (edge-stream-blocks.ts's `selected` note). Drawing it
          again does not silently re-introduce a hit path. */}
      <instancedMesh ref={lineRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <cylinderGeometry args={[EDGE_LINE_RADIUS, EDGE_LINE_RADIUS, 1, 8]} />
        {/* UNLIT (meshBasic), see EDGE_LINE_COLOR: this tone is a rendered appearance, and a
            lit material would render it a second time and come out darker than the beads. */}
        {/* meshBasicMaterial + toneMapped={false}, the SAME pair every mesh in
            ChainBeadInstances uses, and load-bearing for the same reason: it makes the
            authored fill constant equal the pixel. Basic alone was not enough — the
            renderer's default ACES tone mapping still compressed it, so the line came out a
            different colour from the beads despite reading the identical constant. */}
        <meshBasicMaterial color={EDGE_LINE_COLOR} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[ARROW_HEAD_RADIUS, ARROW_HEAD_LENGTH, 12]} />
        {/* UNLIT (meshBasic), see EDGE_LINE_COLOR: this tone is a rendered appearance, and a
            lit material would render it a second time and come out darker than the beads. */}
        {/* meshBasicMaterial + toneMapped={false}, the SAME pair every mesh in
            ChainBeadInstances uses, and load-bearing for the same reason: it makes the
            authored fill constant equal the pixel. Basic alone was not enough — the
            renderer's default ACES tone mapping still compressed it, so the line came out a
            different colour from the beads despite reading the identical constant. */}
        <meshBasicMaterial color={EDGE_LINE_COLOR} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      {/* The RESTING EDGE BEAD, one per edge at its midpoint. Authored at the bead's own
          radius and the EDGE's own fill (EDGE_LINE_COLOR) with the same
          unlit + tone-mapping-exempt pair ChainBeadInstances uses — a chain bead in every
          respect, just placed between the nodes rather than filling the edge.
          raycast disabled: it is scenery, not a target. */}
      <instancedMesh ref={beadRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        {/* EDGE_LINE_COLOR, not the chain fill: this bead belongs to the edge, so it wears
            the edge's own (brighter) tone and stays the line's like-for-like reference. */}
        <meshBasicMaterial color={EDGE_LINE_COLOR} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
      {/* Its RING — geometry and colour copied from ChainBeadInstances verbatim (bead radius,
          tube = radius × SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8×24 segments, the black
          beadStyleForValue(1).ring). Without it this was a bare sphere, which is not what an
          edge bead looks like: the ring carries most of its silhouette. */}
      <instancedMesh ref={beadRingRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <torusGeometry
          args={[SHADING_PARAM_BEAD_RADIUS, SHADING_PARAM_BEAD_RADIUS * SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8, 24]}
        />
        <meshBasicMaterial color={BEAD_RING_COLOR} toneMapped={false} transparent={false} opacity={1} />
      </instancedMesh>
    </>
  );
}
