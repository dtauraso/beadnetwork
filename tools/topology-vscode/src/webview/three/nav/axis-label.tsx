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

  useEffect(() => {
    return () => { texture.dispose(); };
  }, [texture]);
  return (
    <sprite position={position} scale={[size * 4, size, 1]} raycast={() => null}>
      <spriteMaterial map={texture} transparent depthWrite={false} depthTest={false} />
    </sprite>
  );
}
