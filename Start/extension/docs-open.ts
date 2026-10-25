import * as fs from "fs";
import * as http from "http";
import * as path from "path";
import * as crypto from "crypto";
import * as vscode from "vscode";
import { resolveRepoRoot } from "./repo-root";

export function findDefinitionLine(abs: string, symbol: string): number {
  let text: string;
  try {
    text = fs.readFileSync(abs, "utf8");
  } catch {
    return -1;
  }
  const name = symbol.replace(/[^\w$]/g, "");   
  if (!name) return -1;
  const lines = text.split("\n");
  const patterns = [

    new RegExp(`^func\\s+(\\([^)]*\\)\\s*)?${name}\\b`),

    new RegExp(`^(type|const|var)\\s+${name}\\b`),

    new RegExp(`^\\t${name}\\s+[\\w*\\[\\].]|^\\t${name}\\s*=`),

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
  const root = resolveRepoRoot(folder.uri.fsPath);
  if (!root) return;
  const docsDir = path.join(root, "docs", "pair-node");
  if (!fs.existsSync(docsDir)) return;      
  const portFile = path.join(docsDir, "port.js");

  const token = crypto.randomBytes(16).toString("hex");
  if (!/^[0-9a-f]{32}$/.test(token)) {
    throw new Error(`docs-open: token is not 32 hex characters, so it cannot be written into port.js unescaped: ${token}`);
  }
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
      `window.BEADNETWORK_DOCS_OPEN = { port: ${addr.port}, token: "${token}" };\n`,
    );
    log.appendLine(`listening on localhost:${addr.port}, wrote ${portFile}`);
  });

  context.subscriptions.push({
    dispose() {
      server.close();
      try { fs.unlinkSync(portFile); } catch { /* eslint-disable-line no-empty */ }
    },
  });
}
