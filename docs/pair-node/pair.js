// Make each source name open that file as a VS Code editor tab.
//
// The click is a REQUEST, not a link. These pages are read in the Live Preview
// pane, and a link cannot get out of there: Live Preview intercepts every click
// and, for anything off its own host, tells its panel to offer a browser —
// dropping the target URL (its media/main.js). command: URIs need
// enableCommandUris on the owning webview, which is Live Preview's, not ours.
// A fetch does get out: Live Preview serves pages with no Content-Security-
// Policy, so a page it serves can call the topology extension's listener.
//
// port.js carries where to call and this session's token. The extension writes
// it while running and removes it on shutdown, and it is gitignored — so with
// the extension not running there is no listener and the names stay plain text,
// rather than becoming links that go nowhere.
(function () {
  // Where to call. Starts as whatever port.js said when the page loaded, and is
  // re-read from disk when a call fails — see send().
  let ask = window.WIREFOLD_DOCS_OPEN;      // {port, token}
  if (!ask) return;

  // Reloading the VS Code window restarts the extension, which listens on a NEW
  // port with a NEW token and rewrites port.js. A page loaded before that reload
  // still holds the old pair and every click fails silently against a port that
  // is now closed. So a failed call re-reads port.js — cache-busted, since the
  // stale copy is exactly what is wrong — and tries once more. The page heals
  // itself instead of needing to be refreshed by hand.
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
      .catch(function () { return false; });   // extension not running at all
  }

  // A data-src may name a definition inside the file: "nodes/PairNode/node.go#clear".
  // The NAME travels, never a line number — the extension resolves it when the click
  // arrives (docs-open.ts's findDefinitionLine), so nothing here goes stale when the
  // file is edited above the definition.
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
    // MOVE the children in rather than copying textContent out. A data-src used to be
    // put only on a bare name, so flattening to text cost nothing; put one on a whole
    // line — which is what makes a RULE clickable rather than just the name of the
    // function holding it — and flattening throws away the italics, the colours and the
    // <var>s that make the line readable. Moving the nodes keeps all of it inside the
    // link, and is identical to the old behaviour wherever the content was plain text.
    while (cell.firstChild) a.appendChild(cell.firstChild);
    cell.appendChild(a);
  }
})();
