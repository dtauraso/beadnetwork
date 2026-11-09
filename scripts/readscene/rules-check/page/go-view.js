const goData = window.GO_SOURCE || { root: "", functions: [] };
const goSource = {};
for (const f of goData.functions) goSource[f.name] = f;
const portFile = "port.js";
let goAsk = window.BEADNETWORK_DOCS_OPEN;

function rereadPort() {
  return fetch(portFile + "?t=" + Date.now())
    .then((r) => (r.ok ? r.text() : ""))
    .then((text) => {
      const m = text.match(/=\s*(\{[^;]*\})\s*;/);
      if (!m) return false;
      const next = JSON.parse(m[1]);
      const changed = !goAsk || next.port !== goAsk.port || next.token !== goAsk.token;
      goAsk = next;
      return changed;
    })
    .catch(() => false);
}

function askOpen(file, symbol) {
  return fetch("http://localhost:" + goAsk.port + "/open"
    + "?token=" + encodeURIComponent(goAsk.token)
    + "&file=" + encodeURIComponent(file)
    + "&symbol=" + encodeURIComponent(symbol));
}

function openInEditor(file, symbol) {
  const retry = () => rereadPort().then((changed) => { if (changed) askOpen(file, symbol).catch(() => {}); });
  if (!goAsk) { retry(); return; }
  askOpen(file, symbol).catch(retry);
}

function goView() {
  for (const a of document.querySelectorAll("a.gofn")) {
    const f = goSource[a.dataset.go];
    if (!f) {
      a.replaceWith(el("code", { class: "bad" }, a.dataset.go + " (not in data/go-source.js)"));
      continue;
    }
    const symbol = f.name.split(".").pop();
    a.replaceChildren(el("code", {}, f.name), " in ", el("code", {}, f.file.split("/").pop()));
    a.href = "#";
    a.title = "Open " + f.file + " at " + f.name + " as a tab in VS Code";
    a.addEventListener("click", (ev) => {
      ev.preventDefault();
      openInEditor(f.file, symbol);
    });
  }
}
