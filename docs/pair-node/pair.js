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
  const ask = window.WIREFOLD_DOCS_OPEN;     // {port, token}, from port.js
  if (!ask) return;

  for (const cell of document.querySelectorAll("[data-src]")) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.className = "srclink";
    a.textContent = cell.textContent;
    a.href = "#";
    a.title = rel;
    a.addEventListener("click", function (ev) {
      ev.preventDefault();
      fetch("http://localhost:" + ask.port + "/open"
        + "?token=" + encodeURIComponent(ask.token)
        + "&file=" + encodeURIComponent(rel));
    });
    cell.textContent = "";
    cell.appendChild(a);
  }
})();
