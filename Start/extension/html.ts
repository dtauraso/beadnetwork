import * as crypto from "crypto";
import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { SCENES } from "../../Categories/Scene/scenes-gen";

export function buildWebviewHtml(
  webview: vscode.Webview,
  extensionPath: string,
  scenePath: string,
  anchorPath?: string,
): string {
  const scriptPath = path.join(extensionPath, "out", "webview.js");
  const sceneBase = webview.asWebviewUri(vscode.Uri.file(scenePath)).toString();
  const srcBase = webview.asWebviewUri(vscode.Uri.file(realPath(extensionPath))).toString();

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
    window.BEADNETWORK_SCENE_BASE = "${sceneBase}";
    window.BEADNETWORK_SRC_BASE = "${srcBase}";
    window.BEADNETWORK_ANCHOR_BASE = "${anchorBase}";
    window.BEADNETWORK_SCENE_BASES = ${JSON.stringify(sceneBases)};
  </script>
  <script nonce="${nonce}" src="${scriptUri.toString()}"></script>
  <!-- Classic scripts run in order, so by the time this one starts the bundle above has
       either finished evaluating or thrown. Checking here rather than on a timer is what
       makes this honest: the old 3s deadline raced the bundle's own evaluation, and a
       machine busy with a verify run lost that race and cried wolf on a bundle that was
       about to boot. There is no duration to tune, because there is no waiting. -->
  <script nonce="${nonce}">
    if (!window.BEADNETWORK_BOOTED) {
      var app = document.getElementById("app");
      if (app) {
        app.innerHTML =
          '<pre style="margin:0;padding:16px;font:12px ui-monospace,monospace;color:#8b0000;white-space:pre-wrap">' +
          'topology editor: the webview bundle did not run.\\n\\n' +
          'script: ${scriptUri.toString()}\\n' +
          'bundle on disk: ${scriptExists ? "present" : "MISSING"}\\n\\n' +
          'If MISSING, out/webview.js was not built - run "npm run build" and reload\\n' +
          'the window. If present, the bundle threw while evaluating; the webview\\n' +
          'devtools console has the error.' +
          '</pre>';
      }
    }
  </script>
</body>
</html>`;
}

export function realPath(p: string): string {
  try {
    return fs.realpathSync(p);
  } catch {
    return p;
  }
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
