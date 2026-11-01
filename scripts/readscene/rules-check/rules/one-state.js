function wherePoints(link, s) {
  const T = turn(s);
  return {
    phi: wrap(link.start.phi + link.delta.phi, T),
    theta: wrap(link.start.theta + link.delta.theta, T),
    r: link.start.r + link.delta.r,
  };
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
