const bNode = (st, name) => st.nodes.find(x => x.name === name);

function bPack(k, v) {
  if (k === 'from' || k === 'to' || k === 'sender') return v.name;
  if (k === 'via') return [v.from.name, v.to.name];
  if (k === 'out') return undefined;
  return v;
}

const bKeep = st => JSON.stringify([st.nodes, st.rows], bPack);

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, n])), rows = new Map(st.rows.map(r => [`${r.from}→${r.to}`, r]));
  st.rows.forEach(r => { r.from = by.get(r.from); r.to = by.get(r.to); });
  st.rows = st.rows.filter(r => r.from && r.to);
  st.nodes.forEach(n => {
    delete n.ref; delete n.out;
    n.sender = n.sender && by.get(n.sender);
    n.via = n.via && rows.get(n.via.join('→'));
    if (!n.sender || !n.via || n.via.to !== n) { delete n.sender; delete n.via; }
  });
  st.nodes.forEach(n => { if (n.sender) (n.sender.out ||= []).push(n); });
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
  const n = row.to;
  if (bFloat(row) || n.via === row) return;
  bLetGo(n);
  n.sender = row.from;
  n.via = row;
  (row.from.out ||= []).push(n);
}

function bLink(st, row) {
  if (!row.to.sender) bRefer(st, row);
}

function bDropRow(st, row) {
  if (row.to.via === row) bLetGo(row.to);
  st.rows = st.rows.filter(x => x !== row);
}

function bDropNode(st, n) {
  (n.out || []).forEach(x => { delete x.sender; delete x.via; });
  bLetGo(n);
  st.nodes = st.nodes.filter(x => x !== n);
  st.rows = st.rows.filter(r => r.from !== n && r.to !== n);
}

function bTake(st, n) {
  const f = n.sender;
  if (!n.known || !f || !f.known) return false;
  const t = bDir(st, n.via);
  if (!t) return false;
  const J = bIndexOf(st, bAdd(bPlace(st, f.idx), bVec(st, bS(st, f), t)));
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
  return !bTake(st, row.to) || bSend(st, row.to);
}
