(function () {

  let ask = window.BEADNETWORK_DOCS_OPEN;      
  if (!ask) return;

  function reread() {
    return fetch("port.js?t=" + Date.now())
      .then(function (r) { return r.ok ? r.text() : ""; })
      .then(function (text) {
        const m = text.match(/=\s*(\{[^;]*\})\s*;/);
        if (!m) return false;
        const next = JSON.parse(m[1]);
        const changed = !ask || next.port !== ask.port || next.token !== ask.token;
        ask = next;
        return changed;
      })
      .catch(function () { return false; });   
  }

  function call(rel, symbol) {
    return fetch("http://localhost:" + ask.port + "/open"
      + "?token=" + encodeURIComponent(ask.token)
      + "&file=" + encodeURIComponent(rel)
      + (symbol ? "&symbol=" + encodeURIComponent(symbol) : ""));
  }

  function send(rel, symbol) {
    call(rel, symbol).catch(function () {
      reread().then(function (changed) { if (changed) call(rel, symbol).catch(function () { }); });
    });
  }

  for (const cell of document.querySelectorAll("[data-src]")) {
    const hash = cell.getAttribute("data-src").split("#");
    const rel = hash[0];
    const symbol = hash[1] || "";
    const a = document.createElement("a");
    a.className = "srclink";
    a.href = "#";
    a.title = symbol ? rel + " — " + symbol : rel;
    a.addEventListener("click", function (ev) {
      ev.preventDefault();
      send(rel, symbol);
    });

    while (cell.firstChild) a.appendChild(cell.firstChild);
    cell.appendChild(a);
  }
})();
