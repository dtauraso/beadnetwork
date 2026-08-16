import { useState } from "react";
import { useFrame } from "@react-three/fiber";
import type * as THREE from "three";
import { getEdgeBeads } from "./edges/edge-bead-blocks";
import { getNodeFrame } from "./nodes/node-frame-aggregate";
import { INTERIOR_SLOTS_PER_NODE } from "../decode/buffer-decode-interior";

import { ChainBeadInstances } from "./beads/ChainBeadInstances";
import { EdgeLines } from "./edges/EdgeLines";
import { getEdgeStreamAccessor } from "./edges/edge-stream-blocks";
import { TiltVectors } from "../nav/TiltVectors";
import { NodeInstances } from "./nodes/NodeInstances";
import { SelectionHighlight, HoverHighlight } from "./overlays/SelectionHighlight";
import { RuleChannelLines } from "./overlays/RuleChannelLines";
import { InteriorBeadInstances } from "./beads/InteriorBeadInstances";
import { BufferCamera } from "./BufferCamera";
import { BufferLabelProjector } from "./labels/BufferLabelProjector";

export type { BufferLabelPos } from "./buffer-scene-shared";
export {
  BUFFER_NODE_TAG,
  BUFFER_RING_TAG,
  BUFFER_EDGE_TAG,
} from "./buffer-scene-shared";
export { BufferLabelProjector };

const INITIAL_NODE_CAP  = 32;

const INITIAL_CHAINBEAD_CAP = 256;
const INITIAL_EDGE_CAP = 32; 

export function BufferScene({ cameraRef }: {
  cameraRef?: React.MutableRefObject<THREE.PerspectiveCamera | null>;
} = {}) {
  const [nodeCap,  setNodeCap]  = useState(INITIAL_NODE_CAP);
  const [chainBeadCap, setChainBeadCap] = useState(INITIAL_CHAINBEAD_CAP);
  const [edgeCap, setEdgeCap] = useState(INITIAL_EDGE_CAP);

  useFrame(() => {
    const grow: { count: number; cap: number; set: (n: number) => void }[] = [];

    const { count: chainBeadCount } = getEdgeBeads();
    grow.push({ count: chainBeadCount, cap: chainBeadCap, set: setChainBeadCap });

    const edges = getEdgeStreamAccessor();
    if (edges) {
      grow.push({ count: edges.edgeCount, cap: edgeCap, set: setEdgeCap });
    }

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
      {}
      <EdgeLines capacity={edgeCap} />
      <ChainBeadInstances capacity={chainBeadCap} />
      <NodeInstances capacity={nodeCap} />
      {}
      {}
      <TiltVectors capacity={nodeCap * 3} receivedCapacity={nodeCap} />
      <InteriorBeadInstances capacity={nodeCap * INTERIOR_SLOTS_PER_NODE} />
      <SelectionHighlight />
      <HoverHighlight />
      <RuleChannelLines capacity={(nodeCap * (nodeCap - 1)) / 2} />
    </>
  );
}
