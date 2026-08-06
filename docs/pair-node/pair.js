// Turn each source cell into a link that opens that file as a VS Code tab.
//
// VS Code answers on vscode://file/<ABSOLUTE path>, so the page needs this
// clone's own filesystem path. Two ways to get it, in order:
//
//   1. window.WIREFOLD_ROOT, from root.js — generated per clone by
//      tools/docs-root.sh, gitignored, so no machine's path is committed.
//   2. the page's own location, when read as a file:// URL: the pages live at
//      <root>/docs/pair-node/<name>.html, so stripping that suffix gives the
//      root. This does NOT work under Live Preview, which serves the workspace
//      over http — the path there is already workspace-relative, with no
//      filesystem root left in it. That is what (1) is for.
//
// If neither is available the names stay as plain text rather than becoming
// links that go nowhere.
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
