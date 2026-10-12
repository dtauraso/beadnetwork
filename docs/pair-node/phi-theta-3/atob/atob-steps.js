const atobDeg = a => a * 180 / Math.PI;
const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobTurn = (to, from) => { const t = ((to - from) % 360 + 540) % 360 - 180; return `${t >= 0 ? '+' : ''}${t.toFixed(1)}°`; };

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobSteps(pair, r, phi, theta, d, tip, tipA) {
  const n = pair.node, R = Math.hypot(...tip);
  const turned = `turned from the start (φ ${pair.start[0]}°, θ ${pair.start[1]}°) by Δφ = ${atobTurn(phi, pair.start[0])}, Δθ = ${atobTurn(theta, pair.start[1])}`;
  const local = n === 'a'
    ? [`r = ${atobF(r)}, φ = ${phi.toFixed(1)}°, θ = ${theta.toFixed(1)}°`, `from the a→b sliders; ${turned}`]
    : [`(r, φ, θ) = cart_to_polar(C_a + d_ab − C_c) = cart_to_polar(${atobPt(tipA.map((v, i) => v - pair.C[i]))}) = ` +
        `(${atobF(r)}, ${phi.toFixed(1)}°, ${theta.toFixed(1)}°)`,
      `the turn in c's own view that lands where a's does: ${turned}`];
  const read = R < 1e-9 ? `b_${n} is at O — no angle`
    : `R = |b_${n} − O| = ${atobF(R)}, Φ = acos(b_y / R) = ${atobDeg(Math.acos(Math.max(-1, Math.min(1, tip[1] / R)))).toFixed(1)}°, ` +
      `Θ = (atan2(b_z, b_x) + 360°) mod 360° = ${((atobDeg(Math.atan2(tip[2], tip[0])) + 360) % 360).toFixed(1)}°`;
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts'],
    [pair.cls, `local: the ${n}→b vector, in ${n}'s own view`, local[0], local[1]],
    [pair.cls, 'local: as a vector from the node', `d_${n}b = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`, `starts at ${n}`],
    [pair.cls, 'global, last: the new centre', `b_${n} = C_${n} + d_${n}b = ${atobPt(tip)}`, `drawn from the global centre: O→${n} + ${n}→b = O→b_${n}`],
    ['g', 'read from O', read, 'output only; never fed back into 1–4'],
  ]);
}
