// ChainBeadInstances.tsx — the node-owned bead chain that IS the edge (docs/beads-are-the-edge.md).
//
// A node owns one chain per OUTGOING edge. The chain is what a traversal along that edge LOOKS
// like. It is NOT a picture of the node-to-node channels — those are the real connection and
// are never drawn — and its length is not a count of messages: a chain sits fully populated
// with nothing traversing it.
//
// A chain bead is a GREY VERSION OF BEAD 1: same radius, same two-mesh structure (fill sphere
// + ring torus), only the fill differs. Beads sit one DIAMETER apart so they TOUCH — a chain
// is a solid line of beads, not a dotted one. Both the radius and the spacing come from the
// same ShadingParamBeadRadius constant (Go-owned, mirrored into TS), so "no gaps" cannot drift
// into a gap by one side editing its own copy.
//
// The animation is exactly ONE visual difference: the occupied bead's FILL becomes bead 0's or
// bead 1's own fill (bead-style.ts). No size change, no ring change, nothing appears or
// disappears. Which bead is lit, and with which VALUE, is Go's decision entirely (the source
// node drives its own outgoing wires and reads their in-flight fraction —
// nodeMover.chainBeads); this layer only applies the streamed colour.
//
// Pure buffer→GPU, no state authority. Positions come from getChainBeads, which adds each
// node's OWN streamed center to the node-local offsets Go packed (the Interior block's
// convention). That single add is the only arithmetic here.

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getChainBeads } from "./node-stream-blocks";
import { beadStyleForValue } from "./bead-style";
import {
  SHADING_PARAM_BEAD_RADIUS,
  SHADING_PARAM_BEAD_RING_TUBE_RATIO,
  SHADING_PARAM_CHAIN_BEAD_FILL,
} from "../../schema/shading-params";

// Bead 1's own ring, worn by every chain bead whether lit or not — the ring is not part of the
// animation, so it never changes and is read once here.
const RING_COLOR = beadStyleForValue(1)!.ring;

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const bodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const colRef = useRef(new THREE.Color());

  useFrame(() => {
    const body = bodyRef.current;
    const ring = ringRef.current;
    if (!body || !ring) return;

    const { positions, count, lit, litValue } = getChainBeads();
    // Clamp to the allocated instance count. buffer-scene.tsx's capacity-growth table grows
    // chainBeadCap off this same count, so a clamp lasts one frame at most — but it is still a
    // clamp, and it is why this block has its OWN row in that table rather than borrowing
    // another block's cap (the layout-link overlay silently lost links doing that).
    const drawn = Math.min(count, capacity);
    for (let i = 0; i < drawn; i++) {
      matRef.current.makeTranslation(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      body.setMatrixAt(i, matRef.current);
      ring.setMatrixAt(i, matRef.current);

      // The ONE visual difference: an occupied bead wears its traversal's own fill, an empty
      // one wears the grey. A Lit bead whose value is not 0|1 has no style — that is a Go bug
      // rather than a colour to invent, so it stays grey instead of painting a fake one
      // (bead-style.ts's own stance on a non-0/1 value).
      const style = lit[i] === 1 ? beadStyleForValue(litValue[i]) : undefined;
      body.setColorAt(i, colRef.current.set(style ? style.fill : SHADING_PARAM_CHAIN_BEAD_FILL));
      ring.setColorAt(i, colRef.current.set(RING_COLOR));
    }
    body.count = drawn;
    ring.count = drawn;
    body.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (body.instanceColor) body.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      {/* Body + ring: the SAME two-mesh structure and the SAME radius as the 0/1 beads, from
          the shared ShadingParamBead* constants — a chain bead is a recoloured bead, not a
          different marker. The material colour stays white so instanceColor applies verbatim
          (same reasoning as NodeInstances). */}
      <instancedMesh ref={bodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshStandardMaterial emissiveIntensity={0} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <torusGeometry
          args={[SHADING_PARAM_BEAD_RADIUS, SHADING_PARAM_BEAD_RADIUS * SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8, 24]}
        />
        <meshStandardMaterial emissiveIntensity={0} />
      </instancedMesh>
    </>
  );
}
