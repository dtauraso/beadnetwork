const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobTurn = (to, from) => { const t = ((to - from) % 360 + 540) % 360 - 180; return `${t >= 0 ? '+' : ''}${t.toFixed(1)}°`; };

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobSteps(pair, r, phi, theta, d, tip, dAB) {
  const n = pair.node;
  const local = n === 'a'
    ? [['local: the a→b vector, in a\'s own view', `r = ${atobF(r)}, φ = ${phi.toFixed(1)}°, θ = ${theta.toFixed(1)}°`,
        `from the a→b sliders; turned from the start (φ ${pair.start[0]}°, θ ${pair.start[1]}°) by Δφ = ${atobTurn(phi, pair.start[0])}, Δθ = ${atobTurn(theta, pair.start[1])}`],
       ['local: as a vector from a', `d_ab = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`, 'starts at a']]
    : [['local: c\'s vector to a', `d_ca = ${atobPt(pair.toA)}`, 'what c holds for its partner a, in c\'s own view'],
       ['local: the c→b vector', `d_cb = d_ca + d_ab = ${atobPt(pair.toA)} + ${atobPt(dAB)} = ${atobPt(d)}`,
        'c\'s vector to a, then a\'s vector to b: two local vectors added, no centre used'],
       ['local: as polar, in c\'s own view', `r = |d_cb| = ${atobF(r)}, φ = acos(y / r) = ${phi.toFixed(1)}°, θ = atan2(z, x) = ${theta.toFixed(1)}°`,
        'read off d_cb itself, no centre used: the radius and circles of c\'s sphere']];
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts'],
    ...local.map(([name, eq, note]) => [pair.cls, name, eq, note]),
    [pair.cls, 'global, last: the new centre', `b_${n} = C_${n} + d_${n}b = ${atobPt(tip)}`,
      `the only global step: O→${n} + ${n}→b = O→b_${n}`],
  ]);
}
