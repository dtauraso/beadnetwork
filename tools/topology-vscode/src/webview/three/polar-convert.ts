/**
 * The ONE place the webview turns a polar coordinate into a cartesian one.
 *
 * It mirrors Go's polar.Polar2cart exactly — same y-up convention, theta down
 * from +y, phi around it from +x toward +z:
 *
 *     x = r*sin(theta)*cos(phi)
 *     y = r*cos(theta)
 *     z = r*sin(theta)*sin(phi)
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
export function polarToCart(r: number, theta: number, phi: number): [number, number, number] {
  const st = Math.sin(theta);
  return [r * st * Math.cos(phi), r * Math.cos(theta), r * st * Math.sin(phi)];
}
