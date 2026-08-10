// axis-label.tsx — AxisLabel: canvas-texture Sprite billboard used by PolarFrame's axis/arc
// labels (and reusable by other buffer-driven 3D overlays, e.g. PortLabels' port-name tags).
// Always faces the camera, no font asset needed.

import React, { useMemo, useEffect } from "react";
import * as THREE from "three";

export function AxisLabel({ text, color, position, size }: {
  text: string; color: string; position: [number, number, number]; size: number;
}) {
  const texture = useMemo(() => {
    const c = document.createElement("canvas");
    c.width = 256; c.height = 64;
    const ctx = c.getContext("2d");
    if (!ctx) return new THREE.CanvasTexture(c);
    ctx.font = "bold 44px sans-serif";
    ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.fillStyle = color;
    ctx.fillText(text, 128, 32);
    const t = new THREE.CanvasTexture(c);
    t.needsUpdate = true;
    return t;
  }, [text, color]);
  // Dispose the previous texture when deps change and on unmount to prevent GPU memory leaks.
  useEffect(() => {
    return () => { texture.dispose(); };
  }, [texture]);
  return (
    <sprite position={position} scale={[size * 4, size, 1]} raycast={() => null}>
      <spriteMaterial map={texture} transparent depthWrite={false} depthTest={false} />
    </sprite>
  );
}
