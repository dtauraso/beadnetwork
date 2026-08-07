// tilt-vector-angle-format.ts — the pure display-format derivation TiltVectorAnglePanel.tsx
// uses for its per-node θ readout. Split out from the .tsx so it has no react/vscode-api
// dependency and can be unit-tested (and imported) without pulling in a webview module
// (task/pair-lattice-points).
//
// θ is displayed as an INTEGER MULTIPLE of THIS NODE'S OWN lattice step — 2π/points, where
// `points` is the LIVE streamed lattice point count (Buffer/layout.go's LatticePoints), not
// the fixed compile-time CurveParamTiltVectorAngleStep/CURVE_PARAM_TILT_VECTOR_ANGLE_STEP
// (π/12, a 24-point default). That fixed constant is only right at 24 points; deriving from
// the streamed count instead keeps the index and its shown fraction denominator correct at
// whatever count the scene setting currently holds (6 of 24 shows "6π/12", 3 of 12 shows
// "3π/6" — same index, half the denominator, at half the points).
export function formatAngle(radians: number, points: number): string {
  const denom = Math.max(1, Math.round(points / 2));
  const step = (2 * Math.PI) / Math.max(1, points);
  const idx = Math.round(radians / step);
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${denom}`;
}
