// Turn each source cell into a link that opens that file as a VS Code tab.
//
// VS Code answers on vscode://file/<ABSOLUTE path>, so the checkout's own
// filesystem path has to come from somewhere. Nothing machine-specific is
// committed, so it is resolved at read time, in this order:
//
//   1. a path the reader has already given us (localStorage), or
//   2. the page's own location, when it is being read as a file:// URL —
//      the pages live at <root>/docs/pair-node/<name>.html, so stripping
//      that suffix gives the root of whatever clone this is.
//
// Case 2 does NOT work under VS Code's Live Preview, which serves the
// workspace over http://127.0.0.1:PORT: the path is then already
// workspace-relative (/docs/pair-node/formulas.html), so stripping the
// suffix leaves an empty root and every link points at a file that does not
// exist. That is what a dead link looked like here, and it is why the first
// click asks rather than assuming.
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
    // A root of "" is not a root — under a server the suffix is the whole path.
    return at > 0 ? here.slice(0, at) : "";
  }

  let root = stored() || fromLocation();

  const cells = [...document.querySelectorAll("[data-src]")];
  const links = [];

  function wire(a, rel) {
    if (root) {
      a.href = "vscode://file" + root + "/" + rel;
      a.title = root + "/" + rel;
      a.classList.remove("needsroot");
    } else {
      a.href = "#";
      a.title = rel + " — click to point this page at your checkout";
      a.classList.add("needsroot");
    }
  }

  for (const cell of cells) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.className = "srclink";
    a.textContent = cell.textContent;
    wire(a, rel);
    a.addEventListener("click", function (ev) {
      if (root) return;                       // a real vscode:// link; let it go
      ev.preventDefault();
      const guess = prompt(
        "Absolute path of this wirefold checkout, so the source links can open VS Code tabs:",
        "/Users/you/Documents/github/wirefold");
      if (!guess) return;
      root = guess.replace(/\/+$/, "");
      try { localStorage.setItem(KEY, root); } catch (e) { /* private mode: this session only */ }
      links.forEach(function (l) { wire(l.a, l.rel); });
      location.href = a.href;                 // open the one that was clicked
    });
    links.push({ a: a, rel: rel });
    cell.textContent = "";
    cell.appendChild(a);
  }
})();
