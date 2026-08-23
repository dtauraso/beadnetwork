import { useState } from "react";
import { useFrame } from "@react-three/fiber";
import type * as THREE from "three";
import { ownerCounts } from "../../../../Categories/Scene/owner-counts";
import { INTERIOR_SLOTS_PER_NODE } from "../../../../Categories/Node/Interior/interior-values-gen";

import { ChainBeadInstances } from "../../../../Categories/Node/BeadAnimation/ChainBeadInstances";
import { EdgeLines } from "../../../../Categories/Node/Edge/EdgeLines";
import { getEdgeStreamAccessor } from "../../../../Categories/Node/Edge/edge-stream-blocks";
import { TiltVectors } from "../../../../Categories/Node/TiltVectors/TiltVectors";
import { NodeInstances } from "../../../../Categories/Ring/NodeShape/NodeInstances";
import { RuleChannelLines } from "../../../../Categories/Node/ChannelVectors/RuleChannelLines";
import { ChromeCanvas } from "../../../../Categories/Chrome/Panels/ChromeCanvas";
import { InteriorBeadInstances } from "../../../../Categories/Node/Interior/InteriorBeadInstances";
import { SceneCamera } from "../../../../Categories/Scene/Camera/SceneCamera";
import { LabelProjector } from "../../../../Categories/Scene/Labels/LabelProjector";

export type { LabelPos } from "./scene-tags";
export {
  SCENE_NODE_TAG,
  SCENE_RING_TAG,
  SCENE_EDGE_TAG,
} from "./scene-tags";
export { LabelProjector };

const INITIAL_NODE_CAP  = 32;

const INITIAL_CHAINBEAD_CAP = 256;
const INITIAL_EDGE_CAP = 32; 

export function SceneRoot({ cameraRef }: {
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
      <SceneCamera cameraRef={cameraRef} />
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
      <ChromeCanvas />
    </>
  );
}
