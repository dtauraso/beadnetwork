// NodeInstances.tsx — solid node render matching GraphNode's look: a SOLID sphere per node
// (fill from NODE_DEFS[kind].fill) plus a border torus ring (stroke from NODE_DEFS[kind].stroke),
// plus the invisible pick-proxy ring used to author a `port ∈ torus` lock. Split out of
// buffer-scene.tsx: pure buffer→GPU render, no state authority.

import { useRef, useContext } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "./node-stream-blocks";
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
} from "../../schema/shading-params";
import {
  readNodeRingAxisTheta,
  readNodeRingAxisPhi,
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOverlaySelSpherePoles,
} from "../../schema/buffer-layout";
import {
  BUFFER_NODE_TAG, BUFFER_RING_TAG, NODE_SPHERE_RADIUS,
  NODE_RING_TUBE_RATIO, RING_PICK_TUBE_RATIO, nodeRowColors,
} from "./buffer-scene-shared";
import { computeNodeDepthOrder, setNodeDrawOrder } from "./node-depth-order";

// A three.js torusGeometry lies in the XY plane, so its own normal is +Z. Orienting a ring
// means rotating THIS onto the axis Go streams for that node.
const TORUS_DEFAULT_NORMAL = new THREE.Vector3(0, 0, 1);

export function NodeInstances({ capacity }: { capacity: number }) {
  const envTex = useContext(EnvTexContext);
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const ringPickRef = useRef<THREE.InstancedMesh>(null);
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
    if (!body || !ring || !ringPick) return;

    const blocks = getViewBlocks();
    const decodedNode = getNodeFrame();
    if (!decodedNode || !blocks) { body.count = 0; ring.count = 0; ringPick.count = 0; return; }
    const { overlayView } = blocks;
    const { nodeCount, nodeView } = decodedNode;
    // The invisible ring pick-proxy (RING_PICK_TUBE_RATIO) is a pick target ONLY while the
    // selSpherePoles ("select") overlay is on — that's the one mode where a torus click
    // authors a `port ∈ torus` equation. When the overlay is off, count=0 removes it from
    // the raycast scene entirely, so it never steals hits from the node body (BUFFER_NODE_TAG)
    // underneath it. The VISIBLE ring (ringRef, NODE_RING_TUBE_RATIO) is unaffected — it stays
    // rendered in both modes; only the fat invisible pick torus is gated.
    const selectModeOn = readOverlaySelSpherePoles(overlayView) !== 0;

    const n = Math.min(nodeCount, capacity);
    // A node's ring is oriented by the AXIS Go streams for it (PoleTheta/PolePhi), not left
    // at identity. A torusGeometry lies in the XY plane with its normal along +Z, so an
    // unrotated ring sits in the world XY plane wherever the node is and whatever its frame
    // — which is why an edge could run straight through the hole instead of lying in the
    // plane. Same setFromUnitVectors pattern SphereRings.tsx already uses for the scene
    // rings; the axis itself is Go's (nodes/Wiring's inwardPole, projected perpendicular to
    // the edge in a scene that asks for coplanar edges).
    const q = quatRef.current;
    // Depth-sort node rows back-to-front against the live camera THIS frame, in the SAME
    // useFrame that writes the instance matrices — that's what keeps a moved node's new
    // position and its new draw order landing on the same frame (the TIMING CONTRACT this
    // file's imperative useFrame update exists for). order[slot] = nodeRow; writing
    // instance `slot` with row `row`'s transform means the nearest node is written
    // (and drawn) LAST, so it wins the pixel under depthWrite=false (node-depth-order.ts).
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
      // Body: unit sphere scaled to the node radius.
      sclRef.current.setScalar(r);
      matRef.current.compose(posRef.current, q, sclRef.current);
      body.setMatrixAt(slot, matRef.current);
      // Ring: unit torus (major radius 1) scaled to the node radius; tube thickness
      // is baked into the geometry as a fraction of that radius (NODE_RING_TUBE_RATIO).
      // Written at the SAME drawSlot as the body above, so the ring never separates from
      // its node's body once the draw order departs from row order.
      // Ring orientation: rotate the torus's own +Z normal onto the axis Go streams for
      // this node. The BODY stays unrotated (a sphere has no orientation to get wrong), so
      // this is composed separately rather than reusing the body's matrix.
      ringAxis.set(0, 1, 0);
      const poleTheta = readNodeRingAxisTheta(nodeView, row);
      const polePhi = readNodeRingAxisPhi(nodeView, row);
      if (poleTheta !== 0 || polePhi !== 0) {
        const st = Math.sin(poleTheta);
        ringAxis.set(st * Math.cos(polePhi), Math.cos(poleTheta), st * Math.sin(polePhi));
      }
      ringQuatRef.current.setFromUnitVectors(TORUS_DEFAULT_NORMAL, ringAxis);
      matRef.current.compose(posRef.current, ringQuatRef.current, sclRef.current);
      ring.setMatrixAt(slot, matRef.current);
      // Invisible pick-proxy: identical transform to the visible ring, just a thicker
      // raycast target (see RING_PICK_TUBE_RATIO comment). Same drawSlot for the same reason.
      ringPick.setMatrixAt(slot, matRef.current);

      const { fill, stroke } = nodeRowColors(nodeView, row);
      body.setColorAt(slot, colRef.current.set(fill));
      ring.setColorAt(slot, colRef.current.set(stroke));
    }
    body.count = n;
    ring.count = n;
    ringPick.count = selectModeOn ? n : 0;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    ringPick.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
    // Refresh the InstancedMesh bounding sphere so raycast picking stays accurate as
    // nodes move (three.js early-outs a ray against a cached union sphere; a dragged
    // node outside the stale sphere would otherwise be un-pickable). Cheap for the
    // small node counts here.
    body.computeBoundingSphere();
    if (selectModeOn) ringPick.computeBoundingSphere();
  });

  return (
    <>
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} userData={{ [BUFFER_NODE_TAG]: true }} frustumCulled={false}>
        <sphereGeometry args={[1, 16, 16]} />
        {/* Match GraphNode's glassy translucent body EXACTLY (scene-graph.tsx): a
            meshPhysicalMaterial with transmission + depthWrite=false + opacity 0.92 so
            the node interior (held/interior beads) shows through. Per-node fill is the
            instanceColor (setColorAt below); the shared material color stays white so
            instanceColor is applied verbatim. envMap comes from the same PMREM context
            the JSON path uses (BufferScene is wrapped in ProceduralEnvProvider). */}
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
      {/* Invisible pick-proxy torus: same per-instance transform as the visible ring above,
          but a much thicker tube (RING_PICK_TUBE_RATIO) so the ring band is a generous
          raycast target. colorWrite/depthWrite false + zero opacity means it draws nothing;
          it must stay visible (not visible={false}) or three.js Raycaster skips it entirely.
          Tagged BUFFER_RING_TAG so pickBufferRing (scene-content.tsx) resolves its instanceId
          to the same node row as the visible ring. */}
      <instancedMesh ref={ringPickRef} args={[undefined, undefined, capacity]} userData={{ [BUFFER_RING_TAG]: true }} frustumCulled={false}>
        <torusGeometry args={[1, RING_PICK_TUBE_RATIO, 8, 32]} />
        <meshBasicMaterial transparent opacity={0} colorWrite={false} depthWrite={false} />
      </instancedMesh>
    </>
  );
}
