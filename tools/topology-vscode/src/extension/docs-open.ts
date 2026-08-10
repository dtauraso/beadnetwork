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
// findDefinitionLine resolves a NAME to the line that defines it, so the pages can
// say `node.go#handleVectorCycle` and never carry a line number. A line number in a
// doc is wrong the first time anyone inserts a line above it, and nothing tells you —
// the link still opens, just somewhere else. A name is checkable, and
// tools/docs/check-docs-symbols.sh checks it.
//
// Deliberately a regex over the text rather than vscode.executeDocumentSymbolProvider:
// the symbol provider is gopls, which returns nothing until the language server has
// warmed up on that file, so the FIRST click after a window reload — the common one —
// would silently land at the top. This has no warm-up.
//
// Returns a 0-based line, or -1 when the name is not found, in which case the file
// opens at the top exactly as it did before.
export function findDefinitionLine(abs: string, symbol: string): number {
  let text: string;
  try {
    text = fs.readFileSync(abs, "utf8");
  } catch {
    return -1;
  }
  const name = symbol.replace(/[^\w$]/g, "");   // the pages only ever name identifiers
  if (!name) return -1;
  const lines = text.split("\n");
  const patterns = [
    // Go: a plain func or a method with any receiver.
    new RegExp(`^func\\s+(\\([^)]*\\)\\s*)?${name}\\b`),
    // Go: a type, and a const/var/type on its own line.
    new RegExp(`^(type|const|var)\\s+${name}\\b`),
    // Go: a name inside a const/var/type ( ... ) block, indented one tab.
    new RegExp(`^\\t${name}\\s+[\\w*\\[\\].]|^\\t${name}\\s*=`),
    // TS/JS: function, class, and the const-arrow / const-object form.
    new RegExp(`^\\s*(export\\s+)?(async\\s+)?function\\s+${name}\\b`),
    new RegExp(`^\\s*(export\\s+)?(abstract\\s+)?(class|interface|type|enum)\\s+${name}\\b`),
    new RegExp(`^\\s*(export\\s+)?(const|let|var)\\s+${name}\\b`),
  ];
  for (const pattern of patterns) {
    const i = lines.findIndex((l) => pattern.test(l));
    if (i >= 0) return i;
  }
  return -1;
}

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
    const symbol = url.searchParams.get("symbol") ?? "";
    log.appendLine(`request ${url.pathname} file=${rel} symbol=${symbol}`);

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
    const line = symbol ? findDefinitionLine(abs, symbol) : -1;
    log.appendLine(`  opening ${abs}${line >= 0 ? ` at line ${line + 1}` : ""}`);
    const options: vscode.TextDocumentShowOptions = { preview: false };
    if (line >= 0) {
      // Put the definition at the TOP of the viewport rather than centred: what a
      // reader wants from here is the function and its body, and the doc comment
      // above it is what they just read on the page.
      const at = new vscode.Range(line, 0, line, 0);
      options.selection = at;
    }
    vscode.window.showTextDocument(vscode.Uri.file(abs), options).then(
      (ed) => {
        if (line >= 0) {
          ed.revealRange(new vscode.Range(line, 0, line, 0), vscode.TextEditorRevealType.AtTop);
        }
        log.appendLine(`  opened`);
      },
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
