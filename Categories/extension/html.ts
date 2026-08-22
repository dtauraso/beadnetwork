import * as crypto from "crypto";
import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { SCENES } from "../Scene/scenes-gen";

export function buildWebviewHtml(
  webview: vscode.Webview,
  extensionPath: string,
  scenePath: string,
  anchorPath?: string,
): string {
  const scriptPath = path.join(extensionPath, "out", "webview.js");
  const sceneBase = webview.asWebviewUri(vscode.Uri.file(scenePath)).toString();
  const srcBase = webview.asWebviewUri(vscode.Uri.file(path.join(extensionPath, "src"))).toString();

  const anchor = anchorPath ?? scenePath;
  const anchorBase = webview.asWebviewUri(vscode.Uri.file(anchor)).toString();
  const container = path.dirname(anchor);
  const sceneBases: Record<string, string> = {};
  for (const scene of SCENES) {
    sceneBases[scene.name] = webview
      .asWebviewUri(vscode.Uri.file(path.join(container, scene.dir)))
      .toString();
  }

  const scriptExists = fs.existsSync(scriptPath);
  const scriptUri = webview
    .asWebviewUri(vscode.Uri.file(scriptPath))
    .with({ query: `v=${mtimeMs(scriptPath)}` });
  const nonce = randomNonce();

  const csp = [
    `default-src 'none'`,
    `img-src ${webview.cspSource} data:`,
    `style-src ${webview.cspSource} 'unsafe-inline'`,
    `script-src 'nonce-${nonce}'`,
    `font-src ${webview.cspSource}`,
    `connect-src ${webview.cspSource}`,
  ].join("; ");

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta http-equiv="Content-Security-Policy" content="${csp}" />
  <title>Topology Editor</title>
  <!-- All that is left of the stylesheet: the page itself. Every panel is drawn in the
       canvas, so there is nothing else with a class to style. -->
  <style>
    html, body { margin: 0; padding: 0; height: 100%; background: #fafafa; }
    #app { position: relative; width: 100%; height: 100%; }
  </style>
</head>
<body>
  <!-- Every panel is drawn in the canvas now, by Go's own stacks, so the canvas is all that
       is mounted. The two drag-log slots went with the rest: nothing had rendered into them
       since before this change. -->
  <div id="app"></div>
  <script nonce="${nonce}">
    setTimeout(function () {
      if (window.WIREFOLD_BOOTED) return;
      var app = document.getElementById("app");
      if (!app) return;
      app.innerHTML =
        '<pre style="margin:0;padding:16px;font:12px ui-monospace,monospace;color:#8b0000;white-space:pre-wrap">' +
        'topology editor: the webview bundle did not run.\\n\\n' +
        'script: ${scriptUri.toString()}\\n' +
        'bundle on disk: ${scriptExists ? "present" : "MISSING"}\\n\\n' +
        'If MISSING, out/webview.js was not built (or was mid-write when this page\\n' +
        'loaded) - run "npm run build" and reload the window. If present, the bundle\\n' +
        'threw while evaluating; the webview devtools console has the error.' +
        '</pre>';
    }, 3000);
    window.WIREFOLD_SCENE_BASE = "${sceneBase}";
    window.WIREFOLD_SRC_BASE = "${srcBase}";
    window.WIREFOLD_ANCHOR_BASE = "${anchorBase}";
    window.WIREFOLD_SCENE_BASES = ${JSON.stringify(sceneBases)};
  </script>
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
