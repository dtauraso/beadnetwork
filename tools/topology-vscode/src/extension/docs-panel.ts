import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";

// Renders docs/pair-node/*.html in a webview this extension owns, so their source
// names can open the file as an editor tab.
//
// WHY NOT JUST A LINK IN LIVE PREVIEW. A vscode://file link cannot work there:
// Live Preview's injected script intercepts every click and posts
// `open-external-link` to its panel, which answers `open-browser` and DROPS the
// target URL (its media/main.js). Nor is that fixable page-side.
//
// VS Code's own mechanism for this is a command: URI, honoured in a webview whose
// owner sets enableCommandUris. Live Preview's webview is not ours and does not
// set it; this panel is, and does. No port, no listener, no address.
//
// The pages on disk stay the single source: this reads the same file Live Preview
// serves and rewrites three things into webview terms — the stylesheet, the nav
// links (to reopen this panel on another page), and the source cells (to
// vscode.open on the file named by data-src).
export function openDocsPanel(context: vscode.ExtensionContext, page?: string): void {
  const folder = vscode.workspace.workspaceFolders?.[0];
  if (!folder || folder.uri.scheme !== "file") {
    vscode.window.showErrorMessage("Open the wirefold folder to read its docs.");
    return;
  }
  const root = folder.uri.fsPath;
  const dir = path.join(root, "docs", "pair-node");
  const name = sanitize(page) ?? "index";
  const file = path.join(dir, name + ".html");
  if (!fs.existsSync(file)) {
    vscode.window.showErrorMessage(`No docs page ${name}.html`);
    return;
  }

  let panel = docsPanel;
  if (!panel) {
    panel = vscode.window.createWebviewPanel(
      "topologyDocs", "Pair node architecture", vscode.ViewColumn.Active,
      {
        // A click posts a message, exactly as the topology editor's own webview
        // does — the one webview→extension path this repo exercises daily.
        // enableCommandUris stays on for the command: links, but nothing depends
        // on them any more.
        enableScripts: true,
        enableCommandUris: true,
        localResourceRoots: [vscode.Uri.file(dir)],
      },
    );
    docsPanel = panel;
    panel.onDidDispose(() => { docsPanel = undefined; }, undefined, context.subscriptions);
    panel.webview.onDidReceiveMessage((msg: { kind?: string; value?: string }) => {
      trace(root, `message ${JSON.stringify(msg)}`);
      if (msg?.kind === "open-source" && msg.value) openSource(msg.value);
      else if (msg?.kind === "open-page" && msg.value) openDocsPanel(context, msg.value);
    }, undefined, context.subscriptions);
  }

  panel.title = "Pair node — " + name;
  panel.webview.html = render(panel.webview, dir, root, fs.readFileSync(file, "utf8"));
  panel.reveal(panel.viewColumn);
  trace(root, `panel opened ${name}`);
}

// trace appends a breadcrumb to .probe/docs.jsonl. The probe dir is where this
// repo's runtime logs already live (project_probe_log_layout.md) and is
// gitignored. It exists so "nothing happens" can be diagnosed from the log
// instead of from another theory: whether the panel opened, whether a click
// arrived, and what came back.
function trace(root: string, line: string): void {
  try {
    const dir = path.join(root, ".probe");
    fs.mkdirSync(dir, { recursive: true });
    fs.appendFileSync(path.join(dir, "docs.jsonl"),
      JSON.stringify({ t: new Date().toISOString(), line }) + "\n");
  } catch { /* diagnostics must never break the thing they diagnose */ }
}

// One panel, reused: opening a nav link replaces the page rather than stacking tabs.
let docsPanel: vscode.WebviewPanel | undefined;

// openSource opens one of this repo's files as an editor tab, named by its path
// relative to the workspace root. Called by the source links in the docs panel.
// Every failure says so out loud: a link that quietly does nothing is the exact
// thing this whole mechanism exists to stop being.
export function openSource(rel: string): void {
  const folder = vscode.workspace.workspaceFolders?.[0];
  if (folder?.uri.scheme === "file") trace(folder.uri.fsPath, `openSource ${rel}`);
  if (!folder || folder.uri.scheme !== "file") {
    vscode.window.showErrorMessage("No folder open, so there is nothing to open a source file from.");
    return;
  }
  const root = folder.uri.fsPath;
  const abs = path.resolve(root, rel);
  if (abs !== root && !abs.startsWith(root + path.sep)) {
    vscode.window.showErrorMessage(`Refusing to open ${rel}: outside this workspace.`);
    return;
  }
  if (!fs.existsSync(abs)) {
    vscode.window.showErrorMessage(`No such file: ${rel}`);
    return;
  }
  vscode.window.showTextDocument(vscode.Uri.file(abs), { preview: false }).then(
    () => { },
    (err) => vscode.window.showErrorMessage(`Could not open ${rel}: ${String(err)}`),
  );
}

// Page names come from links inside the pages, but they arrive as command
// arguments, so they are treated as untrusted: one path segment, no traversal.
function sanitize(page?: string): string | undefined {
  if (!page) return undefined;
  return /^[a-z0-9-]+$/i.test(page) ? page : undefined;
}

function commandUri(command: string, args: unknown): string {
  return `command:${command}?${encodeURIComponent(JSON.stringify(args))}`;
}

function render(webview: vscode.Webview, dir: string, root: string, html: string): string {
  const css = webview.asWebviewUri(vscode.Uri.file(path.join(dir, "pair.css")));

  return html
    // A marker, so which surface you are reading is never in doubt: the source
    // links are clickable HERE and inert in Live Preview, and those two look
    // otherwise identical. It doubles as a status line — the click handler
    // writes into it, so a click that reaches the page is visible even if
    // nothing downstream happens.
    .replace(/<body>/, `<body>\n  <div id="panel-marker" style="padding:8px 18px 0;font-size:11px;color:#5fd68a">`
      + `topology docs panel — source names open as editor tabs</div>`)
    // The click handler: every link posts to the extension instead of navigating.
    // This is the topology editor's own webview→extension path, the one this repo
    // relies on daily, rather than a scheme some surface may refuse.
    .replace(/<\/body>/, `  <script>
    (function () {
      const vscodeApi = acquireVsCodeApi();
      const marker = document.getElementById("panel-marker");
      document.addEventListener("click", function (ev) {
        const a = ev.target && ev.target.closest ? ev.target.closest("a[data-open], a[data-page]") : null;
        if (!a) return;
        ev.preventDefault();
        const kind = a.dataset.open ? "open-source" : "open-page";
        const value = a.dataset.open || a.dataset.page;
        marker.textContent = "topology docs panel — asked to open " + value;
        vscodeApi.postMessage({ kind: kind, value: value });
      });
    })();
  </script>
</body>`)
    // The stylesheet has to be addressed as a webview resource.
    .replace(/href="pair\.css"/g, `href="${css.toString()}"`)
    // pair.js is for the Live-Preview/browser reading of these same files; here the
    // links are rewritten below, and scripts are off.
    .replace(/<script[^>]*><\/script>\s*/g, "")
    // Nav and index links reopen this panel on another page.
    .replace(/href="([a-z0-9-]+)\.html"/gi,
      (_m: string, p: string) => `data-page="${p}" href="${commandUri("topology.openDocs", [p])}"`)
    // A source cell becomes a link that opens that file as an editor tab.
    //
    // Deliberately NOT the built-in vscode.open: that takes a Uri, and a command
    // URI carries plain JSON, so a Uri does not survive the trip — the command
    // gets a string, and fails silently, which is a dead link that looks live.
    // topology.openSource takes the repo-relative path as a string and builds the
    // Uri on the extension side, where it is a Uri already.
    .replace(/<td class="s" data-src="([^"]+)">([^<]*)<\/td>/g, (_m: string, rel: string, text: string) =>
      `<td class="s"><a class="srclink" title="${rel}" data-open="${rel}"`
      + ` href="${commandUri("topology.openSource", [rel])}">${text}</a></td>`);
}
