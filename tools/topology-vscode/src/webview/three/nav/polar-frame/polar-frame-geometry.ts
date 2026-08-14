export interface PolarFrameGeometry {
  poleLen: number;
  poleRadius: number;
  coneH: number;
  coneBaseR: number;
  arcR: number;
  arcTube: number;
  arcMid: number;
}

export function computePolarFrameGeometry(scale: number): PolarFrameGeometry {
  const radiusKey = Math.max(Math.round(scale), 1);
  const poleLen = radiusKey * 1.3;
  const poleRadius = Math.max(radiusKey * 0.01, 1);
  const coneH = radiusKey * 0.12;
  const coneBaseR = radiusKey * 0.05;
  const arcR = poleLen * 0.68;
  const arcTube = Math.max(radiusKey * 0.012, 1.2);
  const arcMid = arcR * 1.12 * Math.SQRT1_2;
  return { poleLen, poleRadius, coneH, coneBaseR, arcR, arcTube, arcMid };
}
