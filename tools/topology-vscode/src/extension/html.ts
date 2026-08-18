import * as crypto from "crypto";
import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";

export function buildWebviewHtml(
  webview: vscode.Webview,
  extensionPath: string,
): string {
  const scriptPath = path.join(extensionPath, "out", "webview.js");
  const stylePath = path.join(extensionPath, "out", "webview.css");

  const scriptUri = webview
    .asWebviewUri(vscode.Uri.file(scriptPath))
    .with({ query: `v=${mtimeMs(scriptPath)}` });
  const styleUri = webview
    .asWebviewUri(vscode.Uri.file(stylePath))
    .with({ query: `v=${mtimeMs(stylePath)}` });
  const nonce = randomNonce();

  const csp = [
    `default-src 'none'`,
    `img-src ${webview.cspSource} data:`,
    `style-src ${webview.cspSource} 'unsafe-inline'`,
    `script-src 'nonce-${nonce}'`,
    `font-src ${webview.cspSource}`,
  ].join("; ");

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta http-equiv="Content-Security-Policy" content="${csp}" />
  <title>Topology Editor</title>
  <link rel="stylesheet" href="${styleUri.toString()}" />
</head>
<body>
  <!-- What is left of the top-left column. The speed and tilt panels are drawn in the canvas
       now, by Go's own stack, and their two toolbar boxes went with them; the rules panel is
       the last one still in the DOM and starts below them, at the offset the canvas stack
       reports. -->
  <div class="top-stack">
    <div id="node-rules-mount"></div>
  </div>
  <div class="drag-log-row">
    <div id="abc-drag-mount"></div>
    <div id="delta-forward-mount"></div>
  </div>
  <div id="app"></div>
  <script nonce="${nonce}" src="${scriptUri.toString()}"></script>
</body>
</html>`;
}

function randomNonce(): string {
  return crypto.randomBytes(24).toString("base64");
}

function mtimeMs(p: string): number {
  try {
    return Math.floor(fs.statSync(p).mtimeMs);
  } catch {
    return 0;
  }
}
