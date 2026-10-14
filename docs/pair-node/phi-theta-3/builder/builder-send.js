const bNode = (st, name) => st.nodes.find(x => x.name === name);
const bSOf = (st, n) => n.s || st.s;

function bPack(k, v) {
  if (k === 'sender') return v.name;
  if (k === 'via') return [v.from, v.to];
  if (k === 'out') return v.map(x => x.name);
  return v;
}

const bKeep = st => JSON.stringify([st.nodes, st.rows], bPack);

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, n])), rows = new Map(st.rows.map(r => [`${r.from}→${r.to}`, r]));
  st.nodes.forEach(n => {
    if (n.sender) n.sender = by.get(n.sender);
    if (n.via) n.via = rows.get(n.via.join('→'));
    if (n.out) n.out = n.out.map(x => by.get(x)).filter(Boolean);
    if (!n.sender || !n.via) { delete n.sender; delete n.via; }
  });
}

function bRestore(st, keep) {
  [st.nodes, st.rows] = JSON.parse(keep);
  bWire(st);
}

function bLetGo(n) {
  if (n.sender) n.sender.out = n.sender.out.filter(x => x !== n);
  delete n.sender;
  delete n.via;
}

function bRefer(st, row) {
  const n = bNode(st, row.to), from = bNode(st, row.from);
  if (!n || !from || bFloat(row) || n.via === row) return;
  bLetGo(n);
  n.sender = from;
  n.via = row;
  (from.out ||= []).push(n);
}

function bLink(st, row) {
  const n = bNode(st, row.to);
  if (n && !n.sender) bRefer(st, row);
}

function bDropRow(st, row) {
  const n = bNode(st, row.to);
  if (n && n.via === row) bLetGo(n);
  st.rows = st.rows.filter(x => x !== row);
}

function bDropNode(st, n) {
  (n.out || []).forEach(x => { delete x.sender; delete x.via; });
  bLetGo(n);
  st.nodes = st.nodes.filter(x => x !== n);
  st.rows = st.rows.filter(r => r.from !== n.name && r.to !== n.name);
}

function bTakeDir(st, n) {
  const row = n.via, f = n.sender, into = f.via;
  if (bKind(row) !== 'change' || !into || bKind(into) !== 'angle') return bDir(st, row);
  const k = bSOf(st, f) / bSOf(st, f.sender);
  return bChangeDir({ phi: into.start.phi * k, theta: into.start.theta * k }, row, bSOf(st, f));
}

function bTake(st, n) {
  const f = n.sender;
  if (!n.known || !f || !f.known) return false;
  const t = bTakeDir(st, n);
  if (!t) return false;
  const J = bIndexOf(st, bAdd(bPlace(st, f.idx), bVec(st, bSOf(st, f), t)));
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
  return true;
}

function bSend(st, n) {
  const inbox = [...(n.out || [])];
  for (let i = 0; i < inbox.length; i++) {
    if (i >= st.nodes.length) return false;
    if (bTake(st, inbox[i])) inbox.push(...(inbox[i].out || []));
  }
  return true;
}

function bMove(st, row) {
  bRefer(st, row);
  const n = bNode(st, row.to);
  return !n || !bTake(st, n) || bSend(st, n);
}
