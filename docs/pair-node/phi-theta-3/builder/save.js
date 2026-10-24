const B_PAST = [];
let bNow = null, bUndo = null;

function bKeep(builder) {
  const now = JSON.stringify(builder);
  if (bNow !== null && now !== bNow) { B_PAST.push(bNow); if (B_PAST.length > 100) B_PAST.shift(); }
  bNow = now;
  if (bUndo) { bUndo.disabled = !B_PAST.length; bUndo.textContent = `undo (${B_PAST.length})`; }
  bSend();
}

const B_FILE = 'phi-theta-3/builder/sequences.js';
let bSent = null, bSending = Promise.resolve();

function bSend() {
  const at = window.BEADNETWORK_DOCS_OPEN;
  if (!at || bNow === bSent) return;
  const text = `window.B_SAVED = ${JSON.stringify(JSON.parse(bNow), null, 2)};\n`;
  bSent = bNow;
  bSending = bSending.then(() => fetch(`http://localhost:${at.port}/save?token=${encodeURIComponent(at.token)}&file=${encodeURIComponent(B_FILE)}`, { method: 'POST', body: text }))
    .then(r => { if (!r.ok) throw new Error(`status ${r.status}`); })
    .catch(e => { bSent = null; console.warn(`builder: ${B_FILE} not saved: ${e.message}`); });
}
