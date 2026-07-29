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
// Lighting is the animation that replaces the moving bead: the chain is fixed and the LIT
// bead advances along it. Which bead is lit is Go's decision entirely (the source node drives
// its own outgoing wires and reads their in-flight fraction — nodeMover.chainBeads); this
// layer only colours by the streamed Lit flag. A chain with nothing traversing it is fully
// populated and entirely unlit, which is the normal resting state, not missing data.

import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { getChainBeads } from "./node-stream-blocks";

// Unlit beads are the chain's resting structure; a lit bead is where a traversal currently
// is. Both come from the same instanced mesh, differing only in per-instance colour and
// scale, so a chain never renders as two separate objects that could drift apart.
const CHAIN_UNLIT = new THREE.Color("#7a8896");
const CHAIN_LIT = new THREE.Color("#ffd479");
// A lit bead is drawn larger so the animation reads at a glance, not only as a hue shift.
const LIT_SCALE = 2.2;

// Deliberately smaller than BeadInstances' BEAD_SPHERE_RADIUS (4): a chain bead is
// structure, not a transit bead, and at constant spacing along every edge there are many of
// them. Sized so a chain reads as a dotted line rather than a row of touching spheres.
const CHAIN_BEAD_RADIUS = 1.6;

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const meshRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const sclRef = useRef(new THREE.Vector3());
  const posRef = useRef(new THREE.Vector3());
  const quatRef = useRef(new THREE.Quaternion());

  useFrame(() => {
    const mesh = meshRef.current;
    if (!mesh) return;

    const { positions, count, lit } = getChainBeads();
    // Clamp to the allocated instance count. buffer-scene.tsx's capacity-growth table grows
    // `chainBeadCap` off this same count, so a clamp here is one frame at most — but it is
    // still a clamp, and it is why this block has its OWN row in that table rather than
    // borrowing another block's cap (the layout-link overlay silently lost links doing that).
    const drawn = Math.min(count, capacity);
    for (let i = 0; i < drawn; i++) {
      const isLit = lit[i] === 1;
      posRef.current.set(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      const s = isLit ? LIT_SCALE : 1;
      sclRef.current.set(s, s, s);
      matRef.current.compose(posRef.current, quatRef.current, sclRef.current);
      mesh.setMatrixAt(i, matRef.current);
      mesh.setColorAt(i, isLit ? CHAIN_LIT : CHAIN_UNLIT);
    }
    mesh.count = drawn;
    mesh.instanceMatrix.needsUpdate = true;
    if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
  });

  return (
    <instancedMesh ref={meshRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
      <sphereGeometry args={[CHAIN_BEAD_RADIUS, 8, 8]} />
      <meshStandardMaterial emissiveIntensity={0} />
    </instancedMesh>
  );
}
