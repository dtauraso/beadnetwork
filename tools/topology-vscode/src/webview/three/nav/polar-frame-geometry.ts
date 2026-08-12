// polar-frame-geometry.ts — the pure arithmetic PolarFrame derives from its `scale` prop:
// every stick/cone/arc/handhold size below is a fraction of that one number, with no
// react/three/ref dependency, so it is split out the same way tilt-vector-angle-format.ts
// is split from its panel. This is LOCAL DRAWING SCALE for a decorative overlay frame
// (stick lengths, cone sizes, handhold radii), never a bead/node/edge position — `scale`
// and `center` both arrive already computed by Go.

/** One PolarFrame's derived sizes, computed once from its `scale` prop. */
export interface PolarFrameGeometry {
  poleLen: number;
  poleRadius: number;
  coneH: number;
  coneBaseR: number;
  arcR: number;
  arcTube: number;
  arcMid: number;
  hhR: number;
  arcHH: number;
}

/** computePolarFrameGeometry — derives every stick/cone/arc/handhold size PolarFrame draws
 *  from its `scale` prop alone. Pure arithmetic, no THREE/react dependency. */
export function computePolarFrameGeometry(scale: number): PolarFrameGeometry {
  const radiusKey = Math.max(Math.round(scale), 1);
  const poleLen = radiusKey * 1.3;
  const poleRadius = Math.max(radiusKey * 0.01, 1);
  const coneH = radiusKey * 0.12;
  const coneBaseR = radiusKey * 0.05;
  const arcR = poleLen * 0.68;
  const arcTube = Math.max(radiusKey * 0.012, 1.2);
  const arcMid = arcR * 1.12 * Math.SQRT1_2;
  const hhR = Math.max(radiusKey * 0.04, 3);   // handhold sphere radius (matches the tori handholds)
  const arcHH = arcR * Math.SQRT1_2;           // a quarter-arc's midpoint radius (45° in its plane)
  return { poleLen, poleRadius, coneH, coneBaseR, arcR, arcTube, arcMid, hhR, arcHH };
}
