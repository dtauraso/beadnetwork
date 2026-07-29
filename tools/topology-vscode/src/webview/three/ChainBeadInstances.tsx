// ChainBeadInstances.tsx — the node-owned placeholder chain beads (docs/beads-are-the-edge.md).
//
// A node owns one chain per OUTGOING edge; the chain is what a traversal along that edge
// LOOKS like. It is NOT a picture of the node-to-node channels — those are the real
// connection and are never drawn — and its length is not a count of messages: a chain sits
// fully populated with nothing traversing it.
//
// Pure buffer→GPU, no state authority. Positions come from node-stream-blocks.ts's
// getChainBeads, which adds each node's OWN streamed center to the node-local offsets Go
// packed (the Interior block's convention). That single add is the only arithmetic here —
// no interpolation, no spacing decision, no layout: Go owns every offset.
//
// STAGE: step 2 of the plan. These beads are all drawn UNLIT and identical. Lighting them by
// percentage is the animation that replaces the moving bead, and it cannot land until the
// NODE owns traversal timing (step 3) — today that timing lives in PacedWire and reaches the
// editor only on the EDGE stream, so lighting from here would mean this layer deciding which
// bead a traversal has reached, which is exactly the domain state the drift rule forbids TS
// to hold.

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getChainBeads } from "./node-stream-blocks";

// Deliberately smaller than BeadInstances' BEAD_SPHERE_RADIUS (4): a chain bead is
// structure, not a transit bead, and at constant spacing along every edge there are many of
// them. Sized so a chain reads as a dotted line rather than a row of touching spheres.
const CHAIN_BEAD_RADIUS = 1.6;

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const meshRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());

  useFrame(() => {
    const mesh = meshRef.current;
    if (!mesh) return;

    const { positions, count } = getChainBeads();
    // Clamp to the allocated instance count. buffer-scene.tsx's capacity-growth table grows
    // `chainBeadCap` off this same count, so a clamp here is one frame at most — but it is
    // still a clamp, and it is why this block has its OWN row in that table rather than
    // borrowing another block's cap (the layout-link overlay silently lost links doing that).
    const drawn = Math.min(count, capacity);
    for (let i = 0; i < drawn; i++) {
      matRef.current.makeTranslation(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      mesh.setMatrixAt(i, matRef.current);
    }
    mesh.count = drawn;
    mesh.instanceMatrix.needsUpdate = true;
  });

  return (
    <instancedMesh ref={meshRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
      <sphereGeometry args={[CHAIN_BEAD_RADIUS, 8, 8]} />
      <meshStandardMaterial color="#7a8896" emissiveIntensity={0} />
    </instancedMesh>
  );
}
