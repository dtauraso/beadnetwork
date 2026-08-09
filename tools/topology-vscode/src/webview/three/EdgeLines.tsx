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
import { SHADING_PARAM_RING_ROUGHNESS } from "../../schema/shading-params";
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

export function EdgeLines({ capacity }: { capacity: number }) {
  const lineRef = useRef<THREE.InstancedMesh>(null);
  const headRef = useRef<THREE.InstancedMesh>(null);
  const mat = useRef(new THREE.Matrix4());
  const pos = useRef(new THREE.Vector3());
  const dir = useRef(new THREE.Vector3());
  const quat = useRef(new THREE.Quaternion());
  const scl = useRef(new THREE.Vector3());

  useFrame(() => {
    const line = lineRef.current;
    const head = headRef.current;
    if (!line || !head) return;

    const edges = getEdgeStreamAccessor();
    if (!edges) { line.count = 0; head.count = 0; return; }

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

      drawn++;
    }
    line.count = drawn;
    head.count = drawn;
    line.instanceMatrix.needsUpdate = true;
    head.instanceMatrix.needsUpdate = true;
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
        <meshBasicMaterial color={EDGE_LINE_COLOR} />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[undefined, undefined, capacity]} frustumCulled={false} raycast={() => null}>
        <coneGeometry args={[ARROW_HEAD_RADIUS, ARROW_HEAD_LENGTH, 12]} />
        {/* UNLIT (meshBasic), see EDGE_LINE_COLOR: this tone is a rendered appearance, and a
            lit material would render it a second time and come out darker than the beads. */}
        <meshBasicMaterial color={EDGE_LINE_COLOR} />
      </instancedMesh>
    </>
  );
}
