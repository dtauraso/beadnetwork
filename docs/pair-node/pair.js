// Make each source name a link, for the readings of these pages that allow one.
//
// This file is for reading the pages OUTSIDE VS Code — from a browser, where
// vscode://file/<absolute path> launches the editor. tools/docs-root.sh writes
// root.js with this clone's path (gitignored, so no machine's path is committed);
// under file:// the path is derived from the page's own location instead.
//
// It does nothing useful in the Live Preview pane: that webview refuses the
// vscode: scheme, and Live Preview's injected click handler drops the target URL
// before it gets anywhere. Inside VS Code, read these pages with
// "Topology: Open Pair Node Docs" — that panel rewrites the same source cells
// into command: URIs, which is VS Code's own mechanism for opening an editor tab
// (see tools/topology-vscode/src/extension/docs-panel.ts).
//
// With no root available the names stay plain text rather than becoming links
// that go nowhere.
(function () {
  const MARKER = "/docs/pair-node/";

  function fromLocation() {
    if (location.protocol !== "file:") return "";
    const here = decodeURIComponent(location.pathname);
    const at = here.lastIndexOf(MARKER);
    return at > 0 ? here.slice(0, at) : "";   // "" is not a root
  }

  const root = (window.WIREFOLD_ROOT || fromLocation()).replace(/\/+$/, "");
  if (!root) return;

  for (const cell of document.querySelectorAll("[data-src]")) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.className = "srclink";
    a.textContent = cell.textContent;
    a.href = "vscode://file" + root + "/" + rel;
    a.title = root + "/" + rel;
    cell.textContent = "";
    cell.appendChild(a);
  }
})();
