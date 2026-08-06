// Turn each source cell into a link that opens that file as a VS Code tab.
//
// VS Code answers on vscode://file/<ABSOLUTE path>, so the checkout's own
// filesystem path has to come from somewhere. Nothing machine-specific is
// committed, so it is resolved at read time:
//
//   1. a path the reader has given us (localStorage), or
//   2. the page's own location, but ONLY under file:// — the pages live at
//      <root>/docs/pair-node/<name>.html, so stripping that suffix gives the
//      root of whatever clone this is.
//
// Case 2 cannot work under VS Code's Live Preview, which serves the workspace
// over http://127.0.0.1:PORT: the path is then already workspace-relative, so
// there is no filesystem root in it to recover.
//
// The missing root is asked for with an in-page bar rather than prompt():
// VS Code's webview blocks window.prompt outright, so the prompt version of
// this returned null every time and the links stayed dead with no way to fix
// them from inside the pane.
(function () {
  const KEY = "wirefold-root";
  const MARKER = "/docs/pair-node/";

  function stored() {
    try { return localStorage.getItem(KEY) || ""; } catch (e) { return ""; }
  }

  function fromLocation() {
    if (location.protocol !== "file:") return "";
    const here = decodeURIComponent(location.pathname);
    const at = here.lastIndexOf(MARKER);
    return at > 0 ? here.slice(0, at) : "";   // "" is not a root
  }

  let root = stored() || fromLocation();
  const links = [];

  // The link always looks the same — a plain dotted underline. Whether the
  // checkout path is known is said once, in the bar below, not marked on
  // eighteen separate cells.
  function wire(a, rel) {
    a.href = root ? "vscode://file" + root + "/" + rel : "#";
    a.title = root ? root + "/" + rel : rel;
  }

  for (const cell of document.querySelectorAll("[data-src]")) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.className = "srclink";
    a.textContent = cell.textContent;
    wire(a, rel);
    links.push({ a: a, rel: rel });
    cell.textContent = "";
    cell.appendChild(a);
  }
  if (!links.length) return;

  // The bar: shown until a root is known, and re-openable afterwards.
  const bar = document.createElement("div");
  bar.className = "rootbar";

  const msg = document.createElement("span");
  const input = document.createElement("input");
  input.type = "text";
  input.placeholder = "/absolute/path/to/wirefold";
  input.value = root;
  input.spellcheck = false;
  const save = document.createElement("button");
  save.textContent = "Save";

  const note = document.createElement("span");
  note.className = "rootnote";

  function render() {
    msg.textContent = root
      ? "Source links open VS Code tabs under:"
      : "Source links need this checkout's absolute path:";
    note.textContent = location.protocol === "file:"
      ? ""
      : "If a click does nothing in the Live Preview pane, VS Code is blocking the vscode: scheme there — open this page in a browser (the pane's Open in Browser button) and the links work.";
  }

  save.addEventListener("click", function () {
    const v = input.value.trim().replace(/\/+$/, "");
    if (!v) return;
    root = v;
    try { localStorage.setItem(KEY, root); } catch (e) { /* this session only */ }
    links.forEach(function (l) { wire(l.a, l.rel); });
    render();
  });
  input.addEventListener("keydown", function (ev) { if (ev.key === "Enter") save.click(); });

  render();
  bar.append(msg, input, save, note);
  document.querySelector("main").prepend(bar);
})();
