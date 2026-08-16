import React, { useMemo, useEffect } from "react";
import * as THREE from "three";

const LABEL_FONT_PX = 44;
const LABEL_HEIGHT_PX = 64;
const LABEL_PAD_PX = 12;

export function AxisLabel({ text, color, position, size }: {
  text: string; color: string; position: [number, number, number]; size: number;
}) {
  const { texture, aspect } = useMemo(() => {
    const font = `bold ${LABEL_FONT_PX}px sans-serif`;
    const c = document.createElement("canvas");
    const ctx = c.getContext("2d");
    if (!ctx) return { texture: new THREE.CanvasTexture(c), aspect: 1 };

    ctx.font = font;
    const width = Math.max(Math.ceil(ctx.measureText(text).width) + LABEL_PAD_PX * 2, LABEL_HEIGHT_PX);
    c.width = width;
    c.height = LABEL_HEIGHT_PX;

    ctx.font = font;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillStyle = color;
    ctx.fillText(text, width / 2, LABEL_HEIGHT_PX / 2);

    const t = new THREE.CanvasTexture(c);
    t.needsUpdate = true;
    return { texture: t, aspect: width / LABEL_HEIGHT_PX };
  }, [text, color]);

  useEffect(() => {
    return () => { texture.dispose(); };
  }, [texture]);
  return (
    <sprite position={position} scale={[size * aspect, size, 1]} raycast={() => null}>
      <spriteMaterial map={texture} transparent depthWrite={false} depthTest={false} />
    </sprite>
  );
}
