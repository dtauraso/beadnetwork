export function formatAngle(idx: number, points: number): string {
  const denom = Math.max(1, Math.round(points / 2));
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${denom}`;
}

export function widestAngle(points: number): string {
  const denom = Math.max(1, Math.round(points / 2));
  return `-${Math.max(1, points)}π/${denom}`;
}
