// ChainBeadInstances.tsx — the node-owned bead chain that IS the edge (docs/beads-are-the-edge.md).
//
// A node owns one chain per OUTGOING edge. The chain is what a traversal along that edge LOOKS
// like. It is NOT a picture of the node-to-node channels — those are the real connection and
// are never drawn — and its length is not a count of messages: a chain sits fully populated
// with nothing traversing it.
//
// A chain bead is BEAD 1 IN A PALE CYAN: same radius, same fill-sphere + ring-torus structure,
// with a resting bead's fill set to ShadingParamChainBeadFill, drawn with an UNLIT material so
// that constant lands on screen verbatim (see the mesh below). That is DELIBERATELY not the wire
// tube's colour — see that constant's comment, which records why, because an earlier version
// matched the tube exactly and the reasoning for it is persuasive enough to be re-applied by
// mistake. Beads sit one DIAMETER apart so they TOUCH — a chain is a solid line of beads, not a
// dotted one. Both the radius and the spacing come from the
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

// No tube emissive here any more: an unlit bead is deliberately NOT the tube's colour (see
// ShadingParamChainBeadFill), and adding the tube's blue emissive on top of a pale cyan base
// would drag it back toward the blue it is meant to differ from.

export function ChainBeadInstances({ capacity }: { capacity: number }) {
  const unlitBodyRef = useRef<THREE.InstancedMesh>(null);
  const litBodyRef = useRef<THREE.InstancedMesh>(null);
  const ringRef = useRef<THREE.InstancedMesh>(null);
  const matRef = useRef(new THREE.Matrix4());
  const colRef = useRef(new THREE.Color());

  useFrame(() => {
    const unlitBody = unlitBodyRef.current;
    const litBody = litBodyRef.current;
    const ring = ringRef.current;
    if (!unlitBody || !litBody || !ring) return;

    const { positions, count, lit, litValue } = getChainBeads();
    // Clamp to the allocated instance count. buffer-scene.tsx's capacity-growth table grows
    // chainBeadCap off this same count, so a clamp lasts one frame at most — but it is still a
    // clamp, and it is why this block has its OWN row in that table rather than borrowing
    // another block's cap (the layout-link overlay silently lost links doing that).
    const drawn = Math.min(count, capacity);

    // Emissive is a per-MATERIAL property, not a per-instance one (setColorAt only ever
    // reaches the base colour), so one instanced mesh cannot hold both a glowing unlit bead
    // (matching the tube's material exactly) and a flat black/white lit bead. The body is
    // therefore split into two meshes: unlit beads compact into unlitBody, lit beads (with a
    // style) compact into litBody. The ring keeps its single mesh — it never glows either way.
    let unlitCount = 0;
    let litCount = 0;
    for (let i = 0; i < drawn; i++) {
      matRef.current.makeTranslation(positions[i * 3]!, positions[i * 3 + 1]!, positions[i * 3 + 2]!);
      ring.setMatrixAt(i, matRef.current);
      ring.setColorAt(i, colRef.current.set(RING_COLOR));

      // The ONE visual difference: an occupied bead wears its traversal's own fill, an empty
      // one wears the edge tube's own material. A lit bead whose value is not 0|1 has no
      // style — that is a Go bug rather than a colour to invent, so it stays edge-coloured
      // instead of painting a fake one (bead-style.ts's own stance on a non-0/1 value).
      const style = lit[i] === 1 ? beadStyleForValue(litValue[i]) : undefined;
      if (style) {
        litBody.setMatrixAt(litCount, matRef.current);
        litBody.setColorAt(litCount, colRef.current.set(style.fill));
        litCount++;
      } else {
        unlitBody.setMatrixAt(unlitCount, matRef.current);
        unlitCount++;
      }
    }
    unlitBody.count = unlitCount;
    litBody.count = litCount;
    ring.count = drawn;
    unlitBody.instanceMatrix.needsUpdate = true;
    litBody.instanceMatrix.needsUpdate = true;
    ring.instanceMatrix.needsUpdate = true;
    if (litBody.instanceColor) litBody.instanceColor.needsUpdate = true;
    if (ring.instanceColor) ring.instanceColor.needsUpdate = true;
  });

  return (
    <>
      {/* Unlit body: every resting bead is identical, so this mesh needs no
          setColorAt/instanceColor at all — one flat material carries the whole resting chain.

          EVERY bead mesh in this file — this one, the lit body below, and the ring — is
          meshBasicMaterial + toneMapped={false}, and that pair is load-bearing on all of
          them: it makes the authored fill constant EQUAL the pixel. Two stages would
          otherwise alter it — a lit material multiplies the base by incoming light (#a7dfe5
          arrived as ~#8daaad), and the renderer's default ACES tone mapping compresses on top.
          A bead in transit is authored to an exact value, so it opts out of both, same as a
          resting bead.

          A bead inside a node (InteriorBeadInstances.tsx) matches this one by way of its OWN
          authored constants (ShadingParamInteriorBeadFill0/1), NOT by sharing this material —
          an interior bead sits behind the node's glassy transmissive shell, so it needs a
          separately-tuned constant to look the same on screen, not the same material props.
          Cost here: no shading gradient, so a resting bead reads flat and its ring carries the
          silhouette. */}
      <instancedMesh ref={unlitBodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshBasicMaterial color={SHADING_PARAM_CHAIN_BEAD_FILL} toneMapped={false} />
      </instancedMesh>
      {/* Lit body: the 0/1 bead colours via instanceColor — material colour stays white so
          instanceColor applies verbatim.

          meshBasicMaterial + toneMapped={false}, same as every other mesh in this file: the
          streamed 0/1 colour (bead-style.ts) is an authored constant, so it opts out of
          lighting and tone mapping to land on screen verbatim. This does NOT need to match
          InteriorBeadInstances' material — the two look the same by way of SEPARATE authored
          constants (this file's on-wire fills vs. ShadingParamInteriorBeadFill0/1), because an
          interior bead sits behind the node's glassy transmissive shell and a wire bead does
          not; sharing a material could never make them equal on screen. */}
      <instancedMesh ref={litBodyRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <sphereGeometry args={[SHADING_PARAM_BEAD_RADIUS, 16, 16]} />
        <meshBasicMaterial toneMapped={false} />
      </instancedMesh>
      <instancedMesh ref={ringRef} args={[undefined, undefined, capacity]} frustumCulled={false}>
        <torusGeometry
          args={[SHADING_PARAM_BEAD_RADIUS, SHADING_PARAM_BEAD_RADIUS * SHADING_PARAM_BEAD_RING_TUBE_RATIO, 8, 24]}
        />
        <meshBasicMaterial toneMapped={false} />
      </instancedMesh>
    </>
  );
}
