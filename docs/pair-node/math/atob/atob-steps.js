function atobSteps(pair, phi, theta, d, b) {
  const deg = a => a * 180 / Math.PI;
  const f = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
  const pt = p => `(${f(p[0])}, ${f(p[1])}, ${f(p[2])})`;
  const n = pair.node, v = `${n}→b`, R = Math.hypot(...b);
  const read = R < 1e-9 ? `b_${n} is at O — no angle`
    : `R = |b_${n} − O| = ${f(R)}, Φ = acos(b_y / R) = ${deg(Math.acos(Math.max(-1, Math.min(1, b[1] / R)))).toFixed(1)}°, ` +
      `Θ = atan2(b_z, b_x) = ${deg(Math.atan2(b[2], b[0])).toFixed(1)}°`;
  return [
    [pair.cls, `the ${v} vector`, `r = ${f(pair.r)}, φ = ${phi}°, θ = ${theta}°`, `from the ${v} sliders only`],
    [pair.cls, `node ${n}`, `C_${n} = ${pt(pair.C)}`, 'where the vector starts: a known centre, not another pair\'s result'],
    [pair.cls, 'the vector', `d_${n}b = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${pt(d)}`, `starts at ${n}`],
    [pair.cls, 'b goes to its tip', `b_${n} = C_${n} + d_${n}b = ${pt(b)}`, `closes the triangle O, ${n}, b_${n}: O→${n} + ${n}→b = O→b_${n}`],
    ['g', 'read from O', read, 'output only; never fed back into 1–4, and never into the other pair'],
  ].map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}
