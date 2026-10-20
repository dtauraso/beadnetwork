const bF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const bN = n => (Number.isInteger(n) ? `${n}` : bF(n));
const bPt = p => `(${bF(p[0])}, ${bF(p[1])}, ${bF(p[2])})`;
const bIdx = I => `(I_r ${bF(I.r)}, I_φ ${bF(I.phi)}, I_θ ${bF(I.theta)})`;
const bPlaceText = I => `O + polar_to_cart(${bF(I.r)}·c_r, ${bF(I.phi)}·c_φ, ${bF(I.theta)}·c_θ)`;
const bAt = r => `${r.from.name}→${r.to.name}`;
