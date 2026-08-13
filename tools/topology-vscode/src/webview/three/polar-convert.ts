/**
 * The ONE place the webview turns a polar coordinate into a cartesian one.
 *
 * It mirrors Go's polar.Polar2cart exactly — same y-up convention, and the same
 * three.js angle names: PHI is the angle down from +y, THETA the angle around
 * it from +x toward +z.
 *
 *     x = r*sin(phi)*cos(theta)
 *     y = r*cos(phi)
 *     z = r*sin(phi)*sin(theta)
 *
 * This formula had been written out four times across the webview — in
 * viewpoint-bridge, buffer-nav, buffer-scene-shared, and once with phi fixed at
 * 0 inside TiltVectors. Four copies of one convention is four places for a
 * convention change to be missed.
 *
 * The renderer does not COMPUTE geometry here: Go streams theta and phi and
 * this reads them back as a direction to point a mesh along. Go remains the
 * only place a position is decided.
 */
export function polarToCart(r: number, phi: number, theta: number): [number, number, number] {
  const sp = Math.sin(phi);
  return [r * sp * Math.cos(theta), r * Math.cos(phi), r * sp * Math.sin(theta)];
}
