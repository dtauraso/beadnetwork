const atobF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const atobPt = p => `(${atobF(p[0])}, ${atobF(p[1])}, ${atobF(p[2])})`;
const atobAng = p => `r = ${atobF(p[0])}, φ = ${p[1].toFixed(1)}°, θ = ${p[2].toFixed(1)}°`;
const ATOB_BY = { ab: 'the a→b sliders', cb: 'the c→b sliders', both: 'the both sliders' };

function atobList(rows) {
  return rows.map(([c, name, eq, note], i) =>
    `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
}

function atobBoth({ driver, held, both }) {
  if (!held) return atobList([['s', 'no shared sphere', 'C_a = C_c', 'a and c sit on one point, so there is no line between them to hold a shared circle']]);
  const read = driver === 'both' ? 'from the both sliders' : `read off d_ab − m, after ${ATOB_BY[driver]} moved b; sets the both sliders`;
  return atobList([
    ['s', 'the shared sphere, in a\'s view', `m = u·h = ${atobPt(held.m)}, ρ = ${atobF(held.rho)}`,
      'centre on the a–c line, radius of the shared circle; held still while the both sliders drag'],
    ['s', 'b on it', `Φ = ${both[1].toFixed(1)}°, Θ = ${both[2].toFixed(1)}°`, read],
    ['s', 'as a vector from a', `d_ab = m + polar_to_cart(ρ, Φ, Θ)`, 'the shared sphere moves b; a and c follow'],
  ]);
}

function atobSteps(pair, p, d, tip, { driver, dAB, dCB, toC, toA }) {
  const n = pair.node, own = `${n}b`, mine = driver === pair.id;
  let local;
  if (mine) {
    local = [[`local: the ${n}→b vector, in ${n}'s own view`, atobAng(p), `from the ${n}→b sliders; r held`],
      ['local: as a vector', `d_${own} = polar_to_cart(r, φ, θ) = ${atobPt(d)}`, `starts at ${n}`]];
  } else {
    const derive = n === 'c'
      ? ['local: the c→b vector', `d_cb = d_ca + d_ab = ${atobPt(toA)} + ${atobPt(dAB)} = ${atobPt(d)}`,
        'c\'s vector to a, then a\'s vector to b: two local vectors added, no centre used']
      : driver === 'cb'
        ? ['local: the a→b vector', `d_ab = d_ac + d_cb = ${atobPt(toC)} + ${atobPt(dCB)} = ${atobPt(d)}`,
          'a\'s vector to c, then c\'s vector to b: two local vectors added, no centre used']
        : ['local: the a→b vector', `d_ab = m + polar_to_cart(ρ, Φ, Θ) = ${atobPt(d)}`, 'from the shared sphere'];
    local = [derive,
      [`local: as polar, in ${n}'s own view`, `${atobAng(p)}`, `read off d_${own} itself; sets the ${n}→b sliders`]];
  }
  return atobList([
    [pair.cls, `node ${n}`, `C_${n} = ${atobPt(pair.C)}`, 'where the vector starts'],
    ...local.map(([name, eq, note]) => [pair.cls, name, eq, note]),
    [pair.cls, 'global, last: the new centre', `b_${n} = C_${n} + d_${own} = ${atobPt(tip)}`,
      `the only global step: O→${n} + ${n}→b = O→b_${n}`],
  ]);
}
