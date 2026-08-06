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

  const panel = docsPanel ?? vscode.window.createWebviewPanel(
    "topologyDocs", "Pair node architecture", vscode.ViewColumn.Active,
    {
      enableScripts: false,
      enableCommandUris: true,            // what makes the links open editor tabs
      localResourceRoots: [vscode.Uri.file(dir)],
    },
  );
  docsPanel = panel;
  panel.onDidDispose(() => { docsPanel = undefined; }, undefined, context.subscriptions);

  panel.title = "Pair node — " + name;
  panel.webview.html = render(panel.webview, dir, root, fs.readFileSync(file, "utf8"));
  panel.reveal(panel.viewColumn);
}

// One panel, reused: opening a nav link replaces the page rather than stacking tabs.
let docsPanel: vscode.WebviewPanel | undefined;

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
    // The stylesheet has to be addressed as a webview resource.
    .replace(/href="pair\.css"/g, `href="${css.toString()}"`)
    // pair.js is for the Live-Preview/browser reading of these same files; here the
    // links are rewritten below, and scripts are off.
    .replace(/<script[^>]*><\/script>\s*/g, "")
    // Nav and index links reopen this panel on another page.
    .replace(/href="([a-z0-9-]+)\.html"/gi,
      (_m: string, p: string) => `href="${commandUri("topology.openDocs", [p])}"`)
    // A source cell becomes a link that opens that file as an editor tab.
    .replace(/<td class="s" data-src="([^"]+)">([^<]*)<\/td>/g, (_m: string, rel: string, text: string) => {
      const uri = vscode.Uri.file(path.join(root, rel));
      return `<td class="s"><a class="srclink" title="${rel}" href="${commandUri("vscode.open", [uri.toString()])}">${text}</a></td>`;
    });
}
