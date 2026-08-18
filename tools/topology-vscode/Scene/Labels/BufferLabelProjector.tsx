import { useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { nodeLabel } from "../../src/webview/three/decode/buffer-decode-node";
import { ndcToPixel } from "../../src/webview/three/interaction/geometry-helpers";
import { columnF32 } from "../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../Buffer/column-owners";
import {
  COL_STREAM_NODE_LABEL_ANCHOR_X, COL_STREAM_NODE_LABEL_ANCHOR_Y, COL_STREAM_NODE_LABEL_ANCHOR_Z,
} from "../../Buffer/column-streams-gen";
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
    const { nodes: nodeCount } = ownerCounts();
    if (nodeCount <= 0) return;
    const positions: BufferLabelPos[] = [];
    for (let i = 0; i < nodeCount; i++) {

      _bufTopScratch.set(
        columnF32(nodeColumn(i, COL_STREAM_NODE_LABEL_ANCHOR_X)),
        columnF32(nodeColumn(i, COL_STREAM_NODE_LABEL_ANCHOR_Y)),
        columnF32(nodeColumn(i, COL_STREAM_NODE_LABEL_ANCHOR_Z)),
      ).project(camera);
      const topPx = ndcToPixel(_bufTopScratch.x, _bufTopScratch.y, size);
      positions.push({ row: i, label: nodeLabel(i), px: topPx.px, py: topPx.py });
    }
    onPositions(positions);
  });

  return null;
}
