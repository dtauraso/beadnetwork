const bAdd = (p, q) => p.map((v, i) => v + q[i]);
const bSub = (p, q) => p.map((v, i) => v - q[i]);
const bLen = v => Math.hypot(...v);
const bCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
const B_O = [0, 0, 0];
const bWrap = (s, t) => ((t % (12 * s)) + 12 * s) % (12 * s);
const bWhole = x => Math.abs(x - Math.round(x)) < 1e-6;
