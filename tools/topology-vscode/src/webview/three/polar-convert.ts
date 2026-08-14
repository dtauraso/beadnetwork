export function polarToCart(r: number, phi: number, theta: number): [number, number, number] {
  const sp = Math.sin(phi);
  return [r * sp * Math.cos(theta), r * Math.cos(phi), r * sp * Math.sin(theta)];
}
