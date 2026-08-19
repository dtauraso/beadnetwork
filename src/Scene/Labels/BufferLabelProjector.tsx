import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { nodeLabel } from "../../Node/buffer-decode-node";
import { ndcToPixel } from "../../webview/three/interaction/geometry-helpers";
import { columnF32 } from "../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../Buffer/column-owners";
import {
  COL_STREAM_NODE_LABEL_ANCHOR_X, COL_STREAM_NODE_LABEL_ANCHOR_Y,
  COL_STREAM_NODE_LABEL_ANCHOR_Z,
} from "../../Node/columns-gen";
import { overlayFlag } from "../../webview/three/controls/flags/overlay-flags";
import { setLabelPositions } from "./label-canvas";
import type { BufferLabelPos } from "../../webview/three/scene/buffer-scene-shared";

const _bufTopScratch = new THREE.Vector3();

export function BufferLabelProjector() {
  const { camera, gl } = useThree();

  useFrame(() => {
    if (overlayFlag("labelsGlobal")) {
      setLabelPositions([]);
      return;
    }
    const { nodes: nodeCount } = ownerCounts();
    if (nodeCount <= 0) return;

    const el = gl.domElement;
    const size = {
      width: Math.max(1, el.clientWidth),
      height: Math.max(1, el.clientHeight),
    };

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
    setLabelPositions(positions);
  });

  return null;
}
