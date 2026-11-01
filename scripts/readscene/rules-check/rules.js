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

function wherePoints(link, s) {
  const T = turn(s);
  return {
    phi: wrap(link.start.phi + link.delta.phi, T),
    theta: wrap(link.start.theta + link.delta.theta, T),
    r: link.start.r + link.delta.r,
  };
}

function stepDelta(link, axis, sign, s) {
  const delta = { ...link.delta };
  delta[axis] = delta[axis] + sign * link.increment[axis];
  if (axis === "r") {
    if (link.start.r + delta.r < 0) return null;
  } else {
    delta[axis] = wrap(delta[axis], turn(s));
  }
  return delta;
}

function zFlips(points, s) {
  return points.phi > turn(s) / 2 && Math.abs(points.phi - points.theta) < SAME;
}

function linkVector(frame, link, s) {
  const points = wherePoints(link, s);
  const angle = t => t * TAU / turn(s);
  const length = points.r * frame.rho;
  const v = {
    x: length * Math.sin(angle(points.phi)) * Math.cos(angle(points.theta)),
    y: length * Math.cos(angle(points.phi)),
    z: length * Math.sin(angle(points.phi)) * Math.sin(angle(points.theta)),
  };
  if (zFlips(points, s)) v.z = -v.z;
  return v;
}

function nodePosition(frame, place) {
  const length = place.r * frame.constantR;
  const phi = place.phi * TAU / frame.maxIndexPhi;
  const theta = place.theta * TAU / frame.maxIndexTheta;
  return {
    x: length * Math.sin(phi) * Math.cos(theta),
    y: length * Math.cos(phi),
    z: length * Math.sin(phi) * Math.sin(theta),
  };
}

function linkCloses(frame, link, from, to) {
  const a = nodePosition(frame, from.place);
  const b = nodePosition(frame, to.place);
  const end = add(a, linkVector(frame, link, to.s));
  const gap = distance(end, b);
  return { end, b, gap, closes: gap <= ROUNDING * Math.max(1, size(b)) };
}

function isOneStep(before, after, s) {
  for (const axis of AXES) {
    for (const sign of [1, -1]) {
      const delta = stepDelta(before, axis, sign, s);
      if (delta && AXES.every(k => delta[k] === after.delta[k])) return true;
    }
  }
  return false;
}

function stepVerdict(before, after, s, aNodeMoved) {
  const stepped = AXES.some(k => before.delta[k] !== after.delta[k]);
  const startKept = AXES.every(k => Math.abs(before.start[k] - after.start[k]) < SAME);
  const verdict = (ok, why) => ({ ok, why, startKept, aNodeMoved });
  if (stepped) {
    if (!before.picked) return verdict(false, "Δ changed on a link that was not selected");
    if (!isOneStep(before, after, s)) return verdict(false, "Δ changed by something other than one increment");
    if (!startKept) return verdict(false, "Δ was stepped but the start changed too");
    return verdict(true, "stepped: Δ by one increment, start kept");
  }
  if (startKept) return verdict(true, "unchanged");
  if (!aNodeMoved) return verdict(false, "start changed, but neither of its nodes moved");
  return verdict(true, "not stepped: new start, because a node it touches moved");
}

function checkSequence(seq) {
  const nodes = {};
  const positions = {};
  for (const n of seq.nodes) {
    nodes[n.id] = n;
    if (n.place) positions[n.id] = nodePosition(seq.frame, n.place);
  }
  const links = seq.links.map(link => {
    const from = nodes[link.from];
    const to = nodes[link.to];
    const points = wherePoints(link, to.s);
    const row = { link, points, flipped: zFlips(points, to.s), vector: linkVector(seq.frame, link, to.s) };
    if (from && from.place && to.place) Object.assign(row, linkCloses(seq.frame, link, from, to));
    return row;
  });
  return { positions, links };
}

function checkStep(before, after) {
  const name = link => link.from + "→" + link.to;
  const was = {};
  for (const link of before.links) was[name(link)] = link;
  const p0 = checkSequence(before).positions;
  const p1 = checkSequence(after).positions;
  const moved = id => Boolean(p0[id] && p1[id] && distance(p0[id], p1[id]) > ROUNDING * Math.max(1, size(p1[id])));
  const s = {};
  for (const n of after.nodes) s[n.id] = n.s;
  return after.links.filter(link => was[name(link)]).map(link => {
    const old = was[name(link)];
    return { link, old, ...stepVerdict(old, link, s[link.to], moved(link.from) || moved(link.to)) };
  });
}

const RULES_SOURCE = [
  "const TAU = 2 * Math.PI;\nconst ROUNDING = 1e-6;\nconst SAME = 1e-9;\nconst AXES = [\"phi\", \"theta\", \"r\"];",
  turn, wrap, add, distance, size, wherePoints, stepDelta, zFlips, linkVector, nodePosition,
  linkCloses, isOneStep, stepVerdict, checkSequence, checkStep,
].map(String).join("\n\n");
