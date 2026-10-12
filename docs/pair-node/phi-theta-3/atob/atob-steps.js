const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobAng = p => `r = ${atobF(p[0])}, φ = ${p[1].toFixed(1)}°, θ = ${p[2].toFixed(1)}°`;
const atobLen = v => Math.hypot(...v);

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobBoth({ driver, held, both, vec, to, gap, common, m }) {
  const gapRows = ['c', 'd'].map(n => ['s', `the gap, a and ${n}, local`,
    `d_${n}a + d_ab − d_${n}b = ${atobPt(to(n, 'a'))} + ${atobPt(vec.a)} − ${atobPt(vec[n])} = ${atobPt(gap[n])}, |gap| = ${atobF(atobLen(gap[n]))}`,
    atobLen(gap[n]) < 1e-6 ? `zero: b_a = b_${n}` : `not zero: b_a ≠ b_${n}; no centre used to find this`]);
  const ra = atobLen(vec.a), rc = atobLen(vec.c), D = atobLen(to('a', 'c'));
  const meet = ['s', 'do a\'s and c\'s spheres meet?', `|r_a − r_c| ≤ D ≤ r_a + r_c: |${atobF(ra)} − ${atobF(rc)}| ≤ ${atobF(D)} ≤ ${atobF(ra + rc)}`,
    held ? 'yes: they meet in the shared circle' : D < 1e-9 ? 'no: a and c sit on one point, D = 0' : 'no: no point is on both spheres, so there is no shared circle or sphere'];
  const head = [...gapRows, ['s', 'one b?', common ? 'all gaps zero' : 'a gap is not zero',
    common ? 'the three globals land on one point' : 'the globals land on more than one point'], meet];
  if (!held) return atobList(head);
  const where = both
    ? [['s', 'b on it', `Φ = ${both[1].toFixed(1)}°, Θ = ${both[2].toFixed(1)}°`,
        driver === 'both' ? 'from the both sliders' : 'read off d_ab − m_a; sets the both sliders'],
      ['s', 'where b is, from each node', ['a', 'c', 'd'].map(n => `${n}: m_${n} + polar_to_cart(ρ, Φ, Θ) = ${atobPt(vec[n])}`).join(';  '),
        'the same ρ, Φ, Θ from each node\'s own centre; each builds its own vector from it']]
    : [['s', 'b on it', '–', 'no common b: the both box is off, so each node aims on its own; tick it and c and d re-aim at a\'s b']];
  return atobList([...head,
    ['s', 'the shared sphere, a\'s and c\'s, from each node',
      `m_a = u·h = ${atobPt(m.a)}, m_c = d_ca + m_a = ${atobPt(m.c)}, m_d = d_da + m_a = ${atobPt(m.d)}, ρ = ${atobF(held.rho)}`,
      'one centre, seen from a, c and d; u: unit vector a→c; h = (D² + r_a² − r_c²) / 2D; held still while the both sliders drag'],
    ...where]);
}

function atobSteps(pair, p, d, head, { driver, mover, linked, vec, to, moved }) {
  const n = pair.node, own = `${n}b`;
  const build = ['local: build the vector', `d_${own} = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`,
    `${n} builds its own vector from its own r, φ, θ; starts at ${n}`];
  let target = null;
  if (driver === `pos-${n}`) {
    target = [`local: re-aim at b_${n}, from ${n}`, `d_${own} − Δ = ${atobPt(moved.old)} − ${atobPt(moved.delta)} = ${atobPt(d)}`,
      `${n} moved by Δ; it re-aims at the same b, so every gap stays the same; no centre used`];
  } else if (driver === 'both') {
    target = [`local: where b is, from ${n}`, `m_${n} + polar_to_cart(ρ, Φ, Θ) = ${atobPt(d)}`,
      `b's place on the shared sphere, from its centre as ${n} sees it`];
  } else if (linked && mover && mover !== n) {
    target = [`local: where b is, from ${n}`, `d_${n}${mover} + d_${mover}b = ${atobPt(to(n, mover))} + ${atobPt(vec[mover])} = ${atobPt(d)}`,
      `${n}'s vector to ${mover}, then ${mover}'s vector to b: two local vectors added, no centre used`];
  }
  const local = target
    ? [target, [`local: the ${n}→b angles, in ${n}'s own view`, `r = |v|, φ = acos(v_y / r), θ = atan2(v_z, v_x): ${atobAng(p)}`,
        `read off that vector v; sets the ${n}→b sliders`], build]
    : [[`local: the ${n}→b angles, in ${n}'s own view`, atobAng(p),
        driver === pair.id ? `from the ${n}→b sliders; r held` : `${n}'s own angles, unchanged: ${n} did not re-aim`], build];
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts'],
    ...local.map(([name, eq, note]) => [pair.cls, name, eq, note]),
    [pair.cls, 'global, last: the new centre', `b_${n} = C_${n} + d_${own} = ${atobPt(head)}`,
      `the only global step: O→${n} + ${n}→b = O→b_${n}`],
  ]);
}
