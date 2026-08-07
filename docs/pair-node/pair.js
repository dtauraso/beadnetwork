// Make each source name open that file as a VS Code editor tab.
//
// IN THE LIVE PREVIEW PANE the click cannot be a link. Live Preview intercepts
// every click and, for anything off its own host, tells its panel to offer a
// browser — dropping the target URL. So the click asks the running topology
// extension instead, over the address in port.js (written by that extension
// while it runs, gitignored). Live Preview serves pages with no CSP, so a page
// it serves can make that request.
//
// IN A BROWSER there is no extension listening, but vscode://file works there,
// so the href stays a real link and is used when the request is unavailable or
// fails. root.js supplies the absolute path (tools/docs-root.sh).
//
// With neither available the names stay plain text rather than becoming links
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
  const ask = window.WIREFOLD_DOCS_OPEN;     // {port, token} while the extension runs
  if (!root && !ask) return;

  for (const cell of document.querySelectorAll("[data-src]")) {
    const rel = cell.getAttribute("data-src");
    const a = document.createElement("a");
    a.className = "srclink";
    a.textContent = cell.textContent;
    a.href = root ? "vscode://file" + root + "/" + rel : "#";
    a.title = root ? root + "/" + rel : rel;

    if (ask) {
      a.addEventListener("click", function (ev) {
        ev.preventDefault();
        fetch("http://localhost:" + ask.port + "/open"
          + "?token=" + encodeURIComponent(ask.token)
          + "&file=" + encodeURIComponent(rel))
          .catch(function () {
            // Listener gone (extension reloaded, window closed): fall back to the
            // link, which is what a browser would have used anyway.
            if (root) location.href = a.href;
          });
      });
    }

    cell.textContent = "";
    cell.appendChild(a);
  }
})();
