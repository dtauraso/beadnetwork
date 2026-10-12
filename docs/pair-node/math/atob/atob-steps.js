const atobDeg = a => a * 180 / Math.PI;
const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobTarget(R, Phi, Theta, b) {
  return atobList([
    ['g', 'where b is', `b = polar_to_cart(R, Φ, Θ) = polar_to_cart(${atobF(R)}, ${Phi}°, ${Theta}°) = ${atobPt(b)}`,
      'Φ and Θ from the both sliders, about O; R stays √2. Both pairs aim at this one b'],
  ]);
}

function atobSteps(pair, b, r, phi, theta, d, tip) {
  const n = pair.node, v = `${n}→b`, R = Math.hypot(...tip);
  const g = b.map((x, i) => x - pair.C[i]);
  const read = R < 1e-9 ? `b_${n} is at O — no angle`
    : `R = |b_${n} − O| = ${atobF(R)}, Φ = acos(b_y / R) = ${atobDeg(Math.acos(Math.max(-1, Math.min(1, tip[1] / R)))).toFixed(1)}°, ` +
      `Θ = (atan2(b_z, b_x) + 360°) mod 360° = ${((atobDeg(Math.atan2(tip[2], tip[0])) + 360) % 360).toFixed(1)}°`;
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts: a known centre, not another pair\'s result'],
    [pair.cls, `the ${v} vector`,
      `b − C_${n} = ${atobPt(g)}, r = |b − C_${n}| = ${atobF(r)}, φ = acos((b − C_${n})_y / r) = ${phi.toFixed(1)}°, ` +
      `θ = (atan2((b − C_${n})_z, (b − C_${n})_x) + 360°) mod 360° = ${theta.toFixed(1)}°`,
      `aimed at b from ${n}; uses only b and C_${n}, nothing from the other pair`],
    [pair.cls, 'the vector', `d_${n}b = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`, `starts at ${n}`],
    [pair.cls, 'b goes to its tip', `b_${n} = C_${n} + d_${n}b = ${atobPt(tip)}`,
      `closes the triangle O, ${n}, b_${n}: O→${n} + ${n}→b = O→b_${n}, which is the shared b`],
    ['g', 'read from O', read, 'output only; never fed back into 1–4, and never into the other pair'],
  ]);
}
