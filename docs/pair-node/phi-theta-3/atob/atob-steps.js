const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobAng = p => `r = ${atobF(p[0])}, φ = ${p[1].toFixed(1)}°, θ = ${p[2].toFixed(1)}°`;
const ATOB_BY = { ab: 'the a→b sliders', cb: 'the c→b sliders', both: 'the both sliders' };

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobBoth({ driver, held, both, dAB }) {
  if (!held) return atobList([['s', 'no shared sphere', 'C_a = C_c', 'a and c sit on one point, so there is no line between them to hold a shared circle']]);
  const read = driver === 'both' ? 'from the both sliders' : `read off d_ab − m, after ${ATOB_BY[driver]} moved b; sets the both sliders`;
  return atobList([
    ['s', 'the shared sphere, in a\'s view', `m = u·h = ${atobPt(held.m)}, ρ = ${atobF(held.rho)}`,
      'centre on the a–c line, radius of the shared circle; held still while the both sliders drag'],
    ['s', 'b on it', `Φ = ${both[1].toFixed(1)}°, Θ = ${both[2].toFixed(1)}°`, read],
    ['s', 'where b is, from a', `m + polar_to_cart(ρ, Φ, Θ) = ${atobPt(dAB)}`, 'a and c each read their angles off this and build their own vector'],
  ]);
}

function atobSteps(pair, p, d, tip, { driver, dAB, dCB, toC, toA }) {
  const n = pair.node, own = `${n}b`, mine = driver === pair.id;
  const build = ['local: build the vector', `d_${own} = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${atobPt(d)}`,
    `${n} builds its own vector from its own r, φ, θ; starts at ${n}`];
  let local;
  if (mine) {
    local = [[`local: the ${n}→b angles, in ${n}'s own view`, atobAng(p), `from the ${n}→b sliders; r held`], build];
  } else {
    const target = n === 'c'
      ? ['local: where b went, from c', `d_ca + d_ab = ${atobPt(toA)} + ${atobPt(dAB)} = ${atobPt(d)}`,
        'c\'s vector to a, then a\'s vector to b: two local vectors added, no centre used']
      : driver === 'cb'
        ? ['local: where b went, from a', `d_ac + d_cb = ${atobPt(toC)} + ${atobPt(dCB)} = ${atobPt(d)}`,
          'a\'s vector to c, then c\'s vector to b: two local vectors added, no centre used']
        : ['local: where b went, from a', `m + polar_to_cart(ρ, Φ, Θ) = ${atobPt(d)}`, 'b\'s place on the shared sphere'];
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
