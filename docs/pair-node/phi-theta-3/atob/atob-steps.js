const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobAng = p => `r = ${atobF(p[0])}, φ = ${p[1].toFixed(1)}°, θ = ${p[2].toFixed(1)}°`;

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobBoth({ driver, held, both, dAB, dCB, toA, gap, common, ra, rc }) {
  const D = Math.hypot(...toA);
  const gapRow = ['s', 'the gap, local', `d_ca + d_ab − d_cb = ${atobPt(toA)} + ${atobPt(dAB)} − ${atobPt(dCB)} = ${atobPt(gap)}, |gap| = ${atobF(Math.hypot(...gap))}`,
    common ? 'zero: the two globals land on one point, b_a = b_c' : 'not zero: the two globals land on two points, b_a ≠ b_c; no centre used to find this'];
  const meet = ['s', 'do the spheres meet?', `|r_a − r_c| ≤ D ≤ r_a + r_c: |${atobF(ra)} − ${atobF(rc)}| ≤ ${atobF(D)} ≤ ${atobF(ra + rc)}`,
    held ? 'yes: they meet in the shared circle' : D < 1e-9 ? 'no: a and c sit on one point, D = 0' : 'no: no point is on both spheres, so there is no shared circle or sphere'];
  if (!held) return atobList([gapRow, meet]);
  const where = both
    ? [['s', 'b on it', `Φ = ${both[1].toFixed(1)}°, Θ = ${both[2].toFixed(1)}°`,
        driver === 'both' ? 'from the both sliders' : 'read off d_ab − m; sets the both sliders'],
      ['s', 'where b is, from a', `m + polar_to_cart(ρ, Φ, Θ) = ${atobPt(dAB)}`, 'a and c each read their angles off this and build their own vector']]
    : [['s', 'b on it', '–', 'no common b to place; moving the both sliders puts one on the shared sphere and aims both vectors at it']];
  return atobList([gapRow, meet,
    ['s', 'the shared sphere, in a\'s view', `m = u·h = ${atobPt(held.m)}, ρ = ${atobF(held.rho)}`,
      'u: unit vector a→c; h = (D² + r_a² − r_c²) / 2D; held still while the both sliders drag'],
    ...where]);
}

function atobSteps(pair, p, d, tip, { driver, dAB, toA }) {
  const n = pair.node, own = `${n}b`, mine = driver !== 'both';
  const build = ['local: build the vector', `d_${own} = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`,
    `${n} builds its own vector from its own r, φ, θ; starts at ${n}`];
  let local;
  if (mine) {
    local = [[`local: the ${n}→b angles, in ${n}'s own view`, atobAng(p), `from the ${n}→b sliders; r held`], build];
  } else {
    const target = n === 'c'
      ? ['local: where b is, from c', `d_ca + d_ab = ${atobPt(toA)} + ${atobPt(dAB)} = ${atobPt(d)}`,
        'c\'s vector to a, then a\'s vector to b: two local vectors added, no centre used']
      : ['local: where b is, from a', `m + polar_to_cart(ρ, Φ, Θ) = ${atobPt(d)}`, 'b\'s place on the shared sphere'];
    local = [target,
      [`local: the ${n}→b angles, in ${n}'s own view`, `r = |v|, φ = acos(v_y / r), θ = atan2(v_z, v_x): ${atobAng(p)}`,
        `read off that vector v; sets the ${n}→b sliders`], build];
  }
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts'],
    ...local.map(([name, eq, note]) => [pair.cls, name, eq, note]),
    [pair.cls, 'global, last: the new centre', `b_${n} = C_${n} + d_${own} = ${atobPt(tip)}`,
      `the only global step: O→${n} + ${n}→b = O→b_${n}`],
  ]);
}
