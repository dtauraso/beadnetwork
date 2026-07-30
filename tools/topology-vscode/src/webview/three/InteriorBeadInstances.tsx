// InteriorBeadInstances.tsx — interior (held) beads rendered INSIDE each node, matching the
// JSON path's InteriorBeads (scene-beads.tsx). Split out of buffer-scene.tsx: pure
// buffer→GPU render, no state authority.

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getNodeFrame } from "./node-stream-blocks";
import { INTERIOR_SLOTS_PER_NODE } from "./buffer-decode";
import { interiorBeadStyleForValue } from "./bead-style";
import {
  readNodeCX, readNodeCY, readNodeCZ,
  readInteriorPresent, readInteriorValue, readInteriorOX, readInteriorOY, readInteriorOZ,
} from "../../schema/buffer-layout";

// Interior (held) bead sphere radius + ring tube ratio — mirror scene-beads.tsx's
// InteriorSlotBead (INTERIOR_BEAD_R=5, BEAD_RING_TUBE_RATIO=0.12) so the buffer path's
// interior beads match the JSON path's look exactly.
const INTERIOR_BEAD_R = 5;
const INTERIOR_RING_TUBE_RATIO = 0.12;

// The Interior block carries a fixed INTERIOR_SLOTS_PER_NODE rows per node (row =
// nodeRow*slots + slot); a slot draws only when Present=1 AND its value has a bead-style
// (0|1). World position = the node's buffer center + the Go-owned NODE-LOCAL slot offset
// (OX/OY/OZ) — the buffer path has no node group to inherit, so we add the center here (the
// JSON path composes it via the scene graph). Color is value-driven via
// interiorBeadStyleForValue (bead-style.ts, fill sphere + ring torus) — a registry SEPARATE
// from the on-wire beadStyleForValue, because an interior bead renders THROUGH the node's
// glassy transmissive shell (NodeInstances.tsx), which tints it; equality with a wire bead
// is achieved by authoring these constants against that tint, not by sharing a material.
export function InteriorBeadInstances({ capacity }: { capacity: number }) {
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef  = useRef(new THREE.Matrix4());
  const posRef  = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());
  const sclRef  = useRef(new THREE.Vector3());
  const colRef  = useRef(new THREE.Color());

  useFrame(() => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    if (!body || !ring) return;

    const decoded = getNodeFrame();
    if (!decoded) { body.count = 0; ring.count = 0; return; }
    const { nodeCount, nodeView, interiorView } = decoded;

    const q = quatRef.current; // identity (interior beads carry no rotation)
    sclRef.current.setScalar(INTERIOR_BEAD_R);
    let slot = 0;
    for (let i = 0; i < nodeCount && slot < capacity; i++) {
      const cx = readNodeCX(nodeView, i);
      const cy = readNodeCY(nodeView, i);
      const cz = readNodeCZ(nodeView, i);
      for (let s = 0; s < INTERIOR_SLOTS_PER_NODE && slot < capacity; s++) {
        const row = i * INTERIOR_SLOTS_PER_NODE + s;
        if (!readInteriorPresent(interiorView, row)) continue;
        const style = interiorBeadStyleForValue(readInteriorValue(interiorView, row));
        if (!style) continue; // non-0/1 value → hide (never paint a fallback)
        // World = node center + Go-owned node-local slot offset.
        posRef.current.set(
          cx + readInteriorOX(interiorView, row),
          cy + readInteriorOY(interiorView, row),
          cz + readInteriorOZ(interiorView, row),
        );
        matRef.current.compose(posRef.current, q, sclRef.current);
        body.setMatrixAt(slot, matRef.current);
        ring.setMatrixAt(slot, matRef.current);
        body.setColorAt(slot, colRef.current.set(style.fill));
        ring.setColorAt(slot, colRef.current.set(style.ring));
        slot++;
      }
    }
    body.count = slot;
    ring.count = slot;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      {/* Unit-radius geometry scaled per-instance to INTERIOR_BEAD_R; color is
          value-driven via setColorAt (fill sphere + ring torus).

          Both meshes are meshBasicMaterial + toneMapped={false}: an interior bead's fill is
          an AUTHORED constant (ShadingParamInteriorBeadFill0/1 via interiorBeadStyleForValue),
          so it opts out of lighting and out of the renderer's ACES tone mapping so that
          constant lands on screen verbatim.

          DRAW-ORDER FIX (glass must not tint the bead behind it): three.js draws the opaque
          group first, then the transparent group in `renderOrder` order. The node body in
          NodeInstances.tsx is meshPhysicalMaterial with transparent + opacity=
          SHADING_PARAM_NODE_OPACITY + depthWrite={false} (its renderOrder is unset, i.e. the
          default 0), which puts it in the transparent group. These bead meshes used to be
          plain meshBasicMaterial with no `transparent` flag, which put them in the OPAQUE
          group — drawn first — so the node shell (transparent, drawn after) alpha-blended ON
          TOP of an already-drawn bead and dimmed/tinted its authored color. An interior bead
          read dimmer than the identical bead sitting out on a wire for exactly this reason.
          Putting these meshes in the transparent group too (transparent + opacity={1} — NOT
          to make them see-through, opacity=1 is fully opaque, purely so they sort into that
          group) with a renderOrder ABOVE the node shell's means the beads are drawn AFTER
          the shell, so the bead's own authored pixel is the final pixel; the shell still
          blends normally everywhere it isn't covering a bead. depthWrite stays at its
          default (true) so these meshes still occlude each other and other opaque geometry
          correctly.

          KNOWN COST — do not "fix" this without re-reading the above: because the node
          shell doesn't write depth, and these beads now draw in the transparent pass ordered
          by renderOrder rather than by depth, when two nodes overlap on screen a bead
          belonging to the FARTHER node can draw over the glass of a NEARER node. This is a
          chosen tradeoff (interior-bead color fidelity over correct occlusion in that rare
          overlap case), not an oversight. */}
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} renderOrder={1} frustumCulled={false}>
        <sphereGeometry args={[1, 16, 16]} />
        <meshBasicMaterial toneMapped={false} transparent opacity={1} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} renderOrder={1} frustumCulled={false}>
        <torusGeometry args={[1, INTERIOR_RING_TUBE_RATIO, 8, 24]} />
        <meshBasicMaterial toneMapped={false} transparent opacity={1} />
      </instancedMesh>
    </>
  );
}
