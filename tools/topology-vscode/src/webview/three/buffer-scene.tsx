// buffer-scene.tsx — buffer-driven render path orchestrator.
//
// Reads the latest binary snapshot each frame and renders:
//   - Beads: InstancedMesh updated from bead column (positions, live flag).
//   - Nodes: InstancedMesh updated from node column (center positions).
//   - Edges: LineSegments updated from edge column (start/end endpoints).
//
// This component does NOT write to any Zustand store. It reads the snapshot
// buffer directly (zero-copy DataView slices via buffer-decode.ts) and fills
// GPU attribute arrays imperatively via useFrame. No domain state flows out.
//
// The actual per-block renderers live in sibling files (BeadInstances, NodeInstances,
// SelectionHighlight/HoverHighlight, SphereRings, InteriorBeadInstances,
// EdgeTube, BufferCamera, BufferLabelProjector) — this file is just the capacity-manager
// orchestrator that mounts them, plus the shared pick-tag re-exports scene-content.tsx and
// ThreeView.tsx still import from here. There is no PortInstances any more
// (docs/channels-not-ports.md): a port is a load-time channel-binding ROLE, never drawn or
// hit-testable.

import { useState } from "react";
import { useFrame } from "@react-three/fiber";
import type * as THREE from "three";
import { getNodeFrame, getLayoutLinks, getChainBeads } from "./node-stream-blocks";
import { INTERIOR_SLOTS_PER_NODE } from "./buffer-decode";
// BeadInstances (the single MOVING transit bead per wire) is gone: the animation is now the
// LIT bead advancing along a node-owned fixed chain (ChainBeadInstances,
// docs/beads-are-the-edge.md). Two representations of one traversal would drift.
import { ChainBeadInstances } from "./ChainBeadInstances";
import { NodeInstances } from "./NodeInstances";
import { SelectionHighlight, HoverHighlight } from "./SelectionHighlight";
import { SphereRings } from "./SphereRings";
import { InteriorBeadInstances } from "./InteriorBeadInstances";
import { EdgeTubes } from "./EdgeTube";
import { BufferCamera } from "./BufferCamera";
import { BufferLabelProjector } from "./BufferLabelProjector";

export type { BufferLabelPos } from "./buffer-scene-shared";
export {
  BUFFER_NODE_TAG,
  BUFFER_RING_TAG,
  BUFFER_EDGE_TAG,
} from "./buffer-scene-shared";
export { BufferLabelProjector };

// ── Sizing constants ──────────────────────────────────────────────────────────
const INITIAL_NODE_CAP  = 32;
const INITIAL_CHAINBEAD_CAP = 256; // node-owned placeholder chain beads (docs/beads-are-the-edge.md): count is len/spacing summed over every node's OUTGOING edges, so it is far larger than any other block's and independent of every other cap
const INITIAL_LAYOUTLINK_CAP = 32; // layout cascade-link overlay pairs — from LocalPolars filtered to the cascade-link set, NOT the Edge block, so its count is independent of edgeCount and needs its OWN cap

// ── BufferScene ───────────────────────────────────────────────────────────────
// Capacity manager: checks the latest snapshot each frame and grows per-block
// capacities when counts exceed current allocation, triggering a React re-render
// (which remounts the InstancedMesh/LineSegments with a larger buffer).

export function BufferScene({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
} = {}) {
  const [nodeCap,  setNodeCap]  = useState(INITIAL_NODE_CAP);
  const [layoutLinkCap, setLayoutLinkCap] = useState(INITIAL_LAYOUTLINK_CAP);
  const [chainBeadCap, setChainBeadCap] = useState(INITIAL_CHAINBEAD_CAP);

  // Capacity-growth guard: runs every frame to detect need for reallocation. EVERY
  // variable-length streamed block must have a row here — a block whose count outgrows a
  // cap that isn't tracked is silently clamped (the layout-link overlay lost links this
  // way, borrowing edgeCap). Listing them in ONE table (not scattered ifs) makes a new
  // block's capacity a single obvious edit and its omission a visible gap in this list.
  useFrame(() => {
    const grow: { count: number; cap: number; set: (n: number) => void }[] = [];

    // Layout links are aggregated from the per-node dedicated streams' own outbound
    // layout-links (node-stream-blocks.ts's getLayoutLinks) — independent of edge/bead/node
    // stream arrival.
    const { layoutLinkCount } = getLayoutLinks();
    grow.push({ count: layoutLinkCount, cap: layoutLinkCap, set: setLayoutLinkCap });

    // Chain beads are aggregated from the per-node dedicated streams too (getChainBeads) —
    // each node contributes the chains on its OWN outgoing edges. Its own row here, not a
    // share of beadCap: that cap tracks in-flight transit beads on the edge streams, an
    // unrelated and much smaller count.
    const { count: chainBeadCount } = getChainBeads();
    grow.push({ count: chainBeadCount, cap: chainBeadCap, set: setChainBeadCap });

    // No edgeCap/beadCap row any more: nothing renders per-edge geometry or the per-edge
    // transit beads. Each edge's own dedicated stream frame is still read directly by row
    // (edge-stream-blocks.ts, for segment/label decode and the .probe debug log) — that read
    // is un-capacitied (a Map keyed by row), not an instanced-mesh pool sized ahead of time,
    // so there is nothing here to grow.

    // Node/Interior + Label bytes are aggregated from every node row's own dedicated
    // stream frame (node-stream-blocks.ts) — grow nodeCap off that aggregate's count,
    // independent of edge/bead stream arrival.
    const decodedNode = getNodeFrame();
    if (decodedNode) {
      grow.push({ count: decodedNode.nodeCount, cap: nodeCap, set: setNodeCap });
    }

    for (const g of grow) {
      if (g.count > g.cap) g.set(Math.ceil(g.count * 1.5));
    }
  });

  return (
    <>
      <BufferCamera cameraRef={cameraRef} />
      <ChainBeadInstances capacity={chainBeadCap} />
      <NodeInstances capacity={nodeCap} />
      <InteriorBeadInstances capacity={nodeCap * INTERIOR_SLOTS_PER_NODE} />
      <SelectionHighlight />
      <HoverHighlight />
      <SphereRings />
      <EdgeTubes     layoutLinkCapacity={layoutLinkCap} />
    </>
  );
}
