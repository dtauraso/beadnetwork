import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { nodeLabel } from "../../src/webview/three/decode/buffer-decode-node";
import { getNodeFrame } from "../../src/webview/three/scene/nodes/node-frame-aggregate";
import { ndcToPixel } from "../../src/webview/three/interaction/geometry-helpers";
import { readNodeLabelAnchorX, readNodeLabelAnchorY, readNodeLabelAnchorZ } from "../../Buffer/buffer-layout";
import type { BufferLabelPos } from "../../src/webview/three/scene/buffer-scene-shared";

const _bufTopScratch = new THREE.Vector3();

export function BufferLabelProjector({ onPositions }: {
  onPositions: (positions: BufferLabelPos[]) => void;
}) {
  const { camera, size } = useThree();
  const frameCountRef = useRef(0);

  useFrame(() => {
    frameCountRef.current++;
    if (frameCountRef.current % 2 !== 0) return;
    const decoded = getNodeFrame();
    if (!decoded) return;
    const { nodeCount, nodeView } = decoded;
    const positions: BufferLabelPos[] = [];
    for (let i = 0; i < nodeCount; i++) {

      _bufTopScratch.set(
        readNodeLabelAnchorX(nodeView, i),
        readNodeLabelAnchorY(nodeView, i),
        readNodeLabelAnchorZ(nodeView, i),
      ).project(camera);
      const topPx = ndcToPixel(_bufTopScratch.x, _bufTopScratch.y, size);
      positions.push({ row: i, label: nodeLabel(decoded, i), px: topPx.px, py: topPx.py });
    }
    onPositions(positions);
  });

  return null;
}
