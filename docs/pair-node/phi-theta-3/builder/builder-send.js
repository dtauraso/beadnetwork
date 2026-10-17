const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (['gone', 'tree'].includes(k)) return undefined;
  if (k === 'trees') {
    const vars = p => Object.fromEntries(B_PAIR.filter(x => x in p).map(x => [x, p[x]]));
    return [...v].map(t => ({ ...t.ab.row, pairs: t.pairs.map(p => ({ with: p.ac.c.name, ...vars(p) })) }));
  }
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'sender' || k === 'item') return v.name;
  if (k === 'via') return [v.from.name, v.to.name];
  return v;
}

function bOldPairs(st, r) {
  const sib = bOut(st, r.from).find(x => bKind(x) === 'angle');
  const p = sib && sib.tree.pairs.find(q => q.ac.row === r);
  if (p) B_PAIR.filter(x => x in r).forEach(x => { p[x] = r[x]; });
  B_PAIR.forEach(x => delete r[x]);
}

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, n])), rows = st.trees || st.rows, via = new Map(), saved = new Map();
  st.nodes = new Set(by.values());
  st.trees = new Set();
  delete st.rows;
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    saved.set(r, r.pairs); delete r.pairs;
    if (r.from && r.to) { bAddRow(st, r); via.set(key, r); }
  });
  st.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), row = n.via && via.get(n.via.join('→'));
    delete n.ref; delete n.sender; delete n.via;
    if (from && row && row.to === n && row.from === from) { n.sender = from; n.via = row; }
    bRelink(st, n);
  });
  bRows(st).forEach(r => {
    const ps = saved.get(r);
    if (ps) ps.forEach(w => Object.assign(r.tree.pairs.find(p => p.ac.c.name === w.with) || {}, w, { with: undefined }));
    r.tree.pairs.forEach(p => delete p.with);
  });
  bRows(st).forEach(r => { if (bKind(r) === 'copy' && r.ref) bOldPairs(st, r); B_PAIR.filter(x => x !== 'base').forEach(x => delete r[x]); });
  bRows(st).forEach(r => { if (!bFloat(r)) bLink(st, r); });
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

const B_DATA = ['idx', 'start', 'end', 'base', 'ref', 'held', 'went', 'step'];

let bLog = null;
const bKeep = () => (bLog = new Map());
const bDone = () => { bLog = null; };

function bSave(o) {
  if (!bLog || bLog.has(o)) return;
  const c = { ...o };
  B_DATA.forEach(k => { if (c[k]) c[k] = { ...c[k] }; });
  bLog.set(o, c);
}

function bRestore(keep) {
  keep.forEach((c, o) => {
    Object.keys(o).forEach(k => { if (!(k in c)) delete o[k]; });
    Object.assign(o, c);
  });
  bDone();
}

function bLetGo(st, n) {
  bSave(n);
  delete n.sender;
  delete n.via;
  bRelink(st, n);
}

function bRefer(st, row) {
  const n = row.to;
  if (bFloat(row) || n.via === row) return;
  bLetGo(st, n);
  n.sender = row.from;
  n.via = row;
  bRelink(st, n);
}

function bLink(st, row) {
  if (!row.to.sender) bRefer(st, row);
}

function bDropNode(st, n) {
  [...bOut(st, n), ...bIn(st, n)].forEach(r => bDropRow(st, r));
  st.nodes.delete(n);
  n.gone = true;
  if (st.item === n) st.item = bFirst(st.nodes);
}

function bSetAt(st, n, p) {
  const J = bIndexOf(st, p);
  bSave(n);
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
}

const B_I = [1, 0, 0, 0, 1, 0, 0, 0, 1];
const bTimes = (m, v) => [0, 3, 6].map(i => m[i] * v[0] + m[i + 1] * v[1] + m[i + 2] * v[2]);
const bThen = (m, q) => [0, 3, 6].flatMap(i => [0, 1, 2].map(j => m[i] * q[j] + m[i + 1] * q[3 + j] + m[i + 2] * q[6 + j]));
const bFlip = m => [0, 3, 6, 1, 4, 7, 2, 5, 8].map(i => m[i]);
const bFrame = row => (row.turn ? row.turn.m : B_I);

function bPut(st, row) {
  const n = row.to, f = row.from;
  if (!n.known || !f.known) return false;
  const t = bDir(st, row);
  if (!t) return false;
  bSetAt(st, n, bAdd(bPlace(st, f.idx), bTimes(bFrame(row), bVec(st, bS(st, f), t))));
  return true;
}

function bRot(st, row, was, now) {
  const s = bS(st, row.from), k = 2 * Math.PI / (12 * s), a = was.theta * k, b = now.theta * k, f = (now.phi - was.phi) * k;
  const rot = v => {
    const x1 = v[0] * Math.cos(a) + v[2] * Math.sin(a), z1 = v[2] * Math.cos(a) - v[0] * Math.sin(a);
    const x2 = x1 * Math.cos(f) + v[1] * Math.sin(f), y2 = v[1] * Math.cos(f) - x1 * Math.sin(f);
    return [x2 * Math.cos(b) - z1 * Math.sin(b), y2, x2 * Math.sin(b) + z1 * Math.cos(b)];
  };
  const [x, y, z] = [[1, 0, 0], [0, 1, 0], [0, 0, 1]].map(rot), m = bFrame(row);
  return bThen(bThen(m, [x[0], y[0], z[0], x[1], y[1], z[1], x[2], y[2], z[2]]), bFlip(m));
}

function bRound(st, ac) {
  const P = bNodePos(st, ac.a), Q = bNodePos(st, ac.c);
  if (!P || !Q || !ac.rot) return false;
  bSetAt(st, ac.c, bAdd(P, bTimes(ac.rot, bSub(Q, P))));
  return true;
}

function bTurnVia(st, bc, rot) {
  if (!bc) return;
  bSave(bc.row);
  bc.row.turn = { m: bThen(rot, bFrame(bc.row)) };
}

function bCarry(st, p, rot) {
  const tree = p.ab.row.tree;
  p.ac.rot = rot;
  if (bRound(st, p.ac)) bTurnVia(st, tree.bcs[tree.pairs.indexOf(p)], rot);
}

function bHead(st, ab, was) {
  ab.was = was;
  bRefer(st, ab.row);
  ab.now = bDir(st, ab.row);
  bPut(st, ab.row);
}

function bPair(st, p) {
  const { was, now } = p.ab;
  if (!was || !now || bKind(p.ac.row) !== 'copy' || !bAuto(p)) return null;
  p.ac.rot = bRot(st, p.ab.row, was, bStep(st, p, was, now));
  return bRound(st, p.ac) ? p.ac.rot : null;
}

function bMove(st, row, was) {
  const tree = row.tree;
  bHead(st, tree.ab, was);
  tree.pairs.forEach((p, i) => { const rot = bPair(st, p); if (rot) bTurnVia(st, tree.bcs[i], rot); });
}
