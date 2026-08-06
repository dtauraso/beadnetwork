// Turn each source cell into a link that opens that file as a VS Code tab.
//
// The repo root is derived from THIS page's own location rather than written
// down: the pages live at <root>/docs/pair-node/<name>.html, so stripping that
// suffix gives the root in whatever clone the page is being read from. Nothing
// machine-specific is committed.
//
// vscode://file/<absolute path> is what VS Code answers on. It needs a real
// absolute path, which is why this cannot be a plain relative href.
(function () {
  const here = decodeURIComponent(location.pathname);
  const marker = "/docs/pair-node/";
  const at = here.lastIndexOf(marker);
  if (at < 0) return;
  const root = here.slice(0, at);

  for (const cell of document.querySelectorAll("[data-src]")) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.href = "vscode://file" + root + "/" + rel;
    a.textContent = cell.textContent;
    a.title = rel;
    a.className = "srclink";
    cell.textContent = "";
    cell.appendChild(a);
  }
})();
