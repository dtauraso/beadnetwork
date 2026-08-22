import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { nodeLabel } from "../../Node/node-label";
import { ndcToPixel } from "../../Input/Drag/ndc";
import { ownerCounts } from "../../Scene/owner-counts";
import { nodeF32 } from "../../Node/node-leaves";
import { overlayFlag } from "../../Overlay/overlay-flags";
import { setLabelPositions } from "./label-canvas";
import type { LabelPos } from "../../webview/scene/scene-tags";

const _bufTopScratch = new THREE.Vector3();

export function LabelProjector() {
  const { camera, gl } = useThree();

  useFrame(() => {
    if (!overlayFlag("labelsGlobal")) {
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

    const positions: LabelPos[] = [];
    for (let i = 0; i < nodeCount; i++) {
      _bufTopScratch.set(
        nodeF32(i, "labelAnchorX"),
        nodeF32(i, "labelAnchorY"),
        nodeF32(i, "labelAnchorZ"),
      ).project(camera);
      const topPx = ndcToPixel(_bufTopScratch.x, _bufTopScratch.y, size);
      positions.push({ row: i, label: nodeLabel(i), px: topPx.px, py: topPx.py });
    }
    setLabelPositions(positions);
  });

  return null;
}
