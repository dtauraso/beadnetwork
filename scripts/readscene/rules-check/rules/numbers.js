const TAU = 2 * Math.PI;
const ROUNDING = 1e-6;
const SAME = 1e-9;
const AXES = ["phi", "theta", "r"];

function turn(s) {
  return 12 * s;
}

function wrap(t, T) {
  return ((t % T) + T) % T;
}

function add(p, q) {
  return { x: p.x + q.x, y: p.y + q.y, z: p.z + q.z };
}

function distance(p, q) {
  return Math.hypot(p.x - q.x, p.y - q.y, p.z - q.z);
}

function size(p) {
  return Math.hypot(p.x, p.y, p.z);
}

function samePlace(p, q) {
  return distance(p, q) <= ROUNDING * Math.max(1, size(p));
}

function sameTicks(a, b) {
  return AXES.every(k => a[k] === b[k]);
}
