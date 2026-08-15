import { type NavNode } from "./buffer-nav";

export function navSignature(nav: NavNode[]): string {
  let s = "";
  for (const n of nav) {
    s += `${n.row}:${Math.round(n.center.x)},${Math.round(n.center.y)},${Math.round(n.center.z)},${Math.round(n.radius)},${n.selected ? 1 : 0},${n.latchedSel ? 1 : 0};`;
  }
  return s;
}
