import * as fs from "fs";
import * as http from "http";
import * as path from "path";
import * as crypto from "crypto";
import * as vscode from "vscode";

// Lets the docs pages open a source file as an editor tab FROM THE LIVE PREVIEW
// PANE, which is where they are actually read.
//
// A link cannot do it. Live Preview injects a click handler into every page it
// serves; for anything off its own host it posts `open-external-link` to its
// panel, which answers `open-browser` and DROPS the target URL (its media/
// main.js). command: URIs need enableCommandUris on the owning webview, which is
// Live Preview's, not ours. So the click has to leave the page as a request,
// which is what fetch is for: Live Preview serves files with no
// Content-Security-Policy (the only CSP in its bundle belongs to its own panel),
// so a page it serves can call out.
//
// Bound to localhost on an ephemeral port. The page cannot guess the port, so
// address and token are written to docs/pair-node/port.js, which the page loads
// and .gitignore excludes.
//
// The token is not decoration: an endpoint that opens editor tabs is reachable by
// any page in any browser on this machine, and CORS stops a reply being READ, not
// a request being SENT. Requests are also refused unless the path resolves inside
// this workspace.
export function serveDocsOpen(context: vscode.ExtensionContext): void {
  const folder = vscode.workspace.workspaceFolders?.[0];
  if (!folder || folder.uri.scheme !== "file") return;
  const root = folder.uri.fsPath;
  const docsDir = path.join(root, "docs", "pair-node");
  if (!fs.existsSync(docsDir)) return;      // not this workspace's concern
  const portFile = path.join(docsDir, "port.js");

  const token = crypto.randomBytes(16).toString("hex");
  const log = vscode.window.createOutputChannel("topology docs");
  context.subscriptions.push(log);

  const server = http.createServer((req, res) => {
    const url = new URL(req.url ?? "/", "http://localhost");
    res.setHeader("Access-Control-Allow-Origin", "*");
    const rel = url.searchParams.get("file") ?? "";
    log.appendLine(`request ${url.pathname} file=${rel}`);

    if (url.pathname !== "/open" || url.searchParams.get("token") !== token) {
      res.writeHead(404).end();
      return;
    }
    const abs = path.resolve(root, rel);
    if (abs !== root && !abs.startsWith(root + path.sep)) {
      log.appendLine(`  refused: outside workspace`);
      res.writeHead(403).end();
      return;
    }
    if (!fs.existsSync(abs)) {
      log.appendLine(`  refused: no such file`);
      res.writeHead(404).end();
      return;
    }
    res.writeHead(200).end("ok");
    log.appendLine(`  opening ${abs}`);
    vscode.window.showTextDocument(vscode.Uri.file(abs), { preview: false }).then(
      () => log.appendLine(`  opened`),
      (err) => {
        log.appendLine(`  FAILED: ${String(err)}`);
        vscode.window.showErrorMessage(`Could not open ${rel}: ${String(err)}`);
      },
    );
  });

  server.on("error", (err) => log.appendLine(`listener error: ${String(err)}`));
  server.listen(0, "localhost", () => {
    const addr = server.address();
    if (!addr || typeof addr === "string") return;
    fs.writeFileSync(
      portFile,
      "// Written by the topology extension while it runs — gitignored.\n" +
      "// Where the docs pages ask for a file to be opened as an editor tab.\n" +
      `window.WIREFOLD_DOCS_OPEN = ${JSON.stringify({ port: addr.port, token })};\n`,
    );
    log.appendLine(`listening on localhost:${addr.port}, wrote ${portFile}`);
  });

  context.subscriptions.push({
    dispose() {
      server.close();
      try { fs.unlinkSync(portFile); } catch { /* already gone */ }
    },
  });
}
