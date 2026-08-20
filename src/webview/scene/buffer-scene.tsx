import { useState } from "react";
import { useFrame } from "@react-three/fiber";
import type * as THREE from "three";
import { ownerCounts } from "../../schema/buffer-layout/column-owners";
import { INTERIOR_SLOTS_PER_NODE } from "../../Node/Interior/buffer-decode-interior";

import { ChainBeadInstances } from "../../Node/ChainBeadInstances";
import { EdgeLines } from "../../Node/Edge/EdgeLines";
import { getEdgeStreamAccessor } from "../../Node/Edge/edge-stream-blocks";
import { TiltVectors } from "../../Scene/TiltVectors/TiltVectors";
import { NodeInstances } from "../../Node/Shape/NodeInstances";
import { RuleChannelLines } from "../../Scene/Vectors/RuleChannelLines";
import { PanelOverlay } from "../../Chrome/PanelOverlay/PanelOverlay";
import { InteriorBeadInstances } from "../../Node/Interior/InteriorBeadInstances";
import { BufferCamera } from "../../Camera/BufferCamera";
import { BufferLabelProjector } from "../../Scene/Labels/BufferLabelProjector";

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

    const edges = getEdgeStreamAccessor();
    if (edges) {
      grow.push({ count: edges.edgeCount, cap: edgeCap, set: setEdgeCap });
    }

    const nodeTotal = ownerCounts().nodes;
    if (nodeTotal > 0) {
      grow.push({ count: nodeTotal, cap: nodeCap, set: setNodeCap });
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
      <ChainBeadInstances
        capacity={chainBeadCap}
        onCount={(n) => { if (n > chainBeadCap) setChainBeadCap(Math.ceil(n * 1.5)); }}
      />
      <NodeInstances capacity={nodeCap} />
      {}
      {}
      <TiltVectors capacity={nodeCap * 3} receivedCapacity={nodeCap} />
      <InteriorBeadInstances capacity={nodeCap * INTERIOR_SLOTS_PER_NODE} />
      <RuleChannelLines capacity={2 * nodeCap * (nodeCap - 1)} />
      <PanelOverlay />
    </>
  );
}
