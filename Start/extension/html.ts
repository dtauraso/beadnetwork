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
  const scriptPath = path.join(realPath(extensionPath), "out", "webview.js");
  const sceneBase = webview.asWebviewUri(vscode.Uri.file(scenePath)).toString();
  const srcBase = webview.asWebviewUri(vscode.Uri.file(realPath(extensionPath))).toString();

  const anchor = anchorPath ?? scenePath;
  const anchorBase = webview.asWebviewUri(vscode.Uri.file(anchor)).toString();
  const container = path.dirname(anchor);
  const sceneBases: Record<string, string> = {};
  for (const scene of SCENES) {
    const base = webview
      .asWebviewUri(vscode.Uri.file(path.join(container, scene.dir)))
      .toString();

    sceneBases[scene.dir] = base;
    sceneBases[scene.name] = base;
  }

  const scriptExists = fs.existsSync(scriptPath);
  const scriptUri = `${webview.asWebviewUri(vscode.Uri.file(scriptPath)).toString()}?v=${mtimeMs(scriptPath)}`;
  const nonce = randomNonce();

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
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
  <!-- A marker that needs NO script. A blank panel cannot tell you whether the html
       failed to load or the scripts in it failed to run; those have opposite causes
       and every diagnosis so far has had to guess between them. If this text is on
       screen the html arrived and the problem is script execution. The boot trace
       replaces it the moment any script runs. -->
  <div id="html-marker" style="position:fixed;left:0;bottom:0;z-index:2147483646;padding:6px;font:11px ui-monospace,monospace;color:#888;background:#111">html loaded, no script has run yet &nbsp; build ${mtimeMs(scriptPath)} &nbsp; nonce ${nonce.slice(0, 6)} &nbsp; scriptExists ${scriptExists} &nbsp; enableScripts ${String(webview.options.enableScripts)} &nbsp; roots ${(webview.options.localResourceRoots ?? []).length}</div>
  <div id="app"></div>
  <script nonce="${nonce}">
    window.BEADNETWORK_SCENE_BASE = "${sceneBase}";
    window.BEADNETWORK_SRC_BASE = "${srcBase}";
    window.BEADNETWORK_ANCHOR_BASE = "${anchorBase}";
    window.BEADNETWORK_SCENE_BASES = ${JSON.stringify(sceneBases)};

    window.__vscodeApi = window.__vscodeApi || acquireVsCodeApi();

    window.BEADNETWORK_BLOCKS = (function () {
      var handlers = [];
      window.addEventListener("message", function (e) {
        var m = e.data;
        if (!m || m.type !== "block") return;
        for (var i = 0; i < handlers.length; i++) handlers[i](m);
      });
      return {
        on: function (fn) { handlers.push(fn); },
        want: function (pathsDir, cadenceMs) {
          window.__vscodeApi.postMessage({ type: "want-block", pathsDir: pathsDir, cadenceMs: cadenceMs });
        },
        wantFile: function (pathsDir, pathFile, cadenceMs) {
          window.__vscodeApi.postMessage({ type: "want-block", pathsDir: pathsDir, pathFile: pathFile, cadenceMs: cadenceMs });
        },
        wantRows: function (pathsDir, rows, cadenceMs) {
          window.__vscodeApi.postMessage({ type: "want-block", pathsDir: pathsDir, rows: rows, cadenceMs: cadenceMs });
        },
        bytes: function (b64) {
          var bin = atob(b64);
          var out = new Uint8Array(bin.length);
          for (var i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
          return out.buffer;
        },
      };
    })();
  </script>
  <!-- The failure we are chasing shows nothing on screen and nothing in the window's own
       console, because both live in the webview's separate devtools. So the page reports
       itself. Capture phase is what makes this work: a script that 404s fires an error
       event on the ELEMENT (no message, no stack) and never bubbles, while a script that
       throws while evaluating fires on the WINDOW with both. Those two have opposite
       causes and every round of this so far has had to guess between them. -->
  <script nonce="${nonce}">
    window.BEADNETWORK_FAULTS = [];
    window.addEventListener("error", function (e) {
      var onElement = e.target && e.target !== window && e.target.tagName;
      window.BEADNETWORK_FAULTS.push(onElement
        ? "RESOURCE FAILED TO LOAD\\n  " + e.target.tagName + "  " + (e.target.src || e.target.href || "")
        : "THREW WHILE EVALUATING\\n  " + e.message + "\\n  at " + e.filename + ":" + e.lineno + ":" + e.colno +
          "\\n\\n" + ((e.error && e.error.stack) || "(no stack)"));
    }, true);
    window.addEventListener("unhandledrejection", function (e) {
      window.BEADNETWORK_FAULTS.push("REJECTED PROMISE\\n  " + ((e.reason && (e.reason.stack || e.reason.message)) || String(e.reason)));
    });

    setTimeout(function () {
      if (window.BEADNETWORK_BOOTED || window.BEADNETWORK_SETTLED) return;
      var box = document.createElement("pre");
      box.setAttribute("style", "position:fixed;left:0;top:0;z-index:2147483647;margin:0;padding:16px;" +
        "font:12px ui-monospace,monospace;color:#ffd166;background:rgba(0,0,0,.9);white-space:pre-wrap");
      box.textContent =
        'topology editor: the bundle request has not settled after 5s.\\n\\n' +
        'script: ${scriptUri}\\n' +
        'readyState: ' + document.readyState + '\\n' +
        'faults so far: ' + (window.BEADNETWORK_FAULTS.length || 'none') + '\\n\\n' +
        'The request was issued and neither loaded nor errored.\\n\\n' +
        'probing the resource path directly...';
      document.body.appendChild(box);

      var probe = function (label, url) {
        var started = Date.now();
        var stop = new AbortController();
        setTimeout(function () { stop.abort(); }, 4000);
        return fetch(url, { signal: stop.signal })
          .then(function (r) { return label + ': HTTP ' + r.status + ' in ' + (Date.now() - started) + 'ms'; })
          .catch(function (e) { return label + ': ' + (e && e.name === 'AbortError' ? 'STILL PENDING after 4s' : String(e)); });
      };

      Promise.all([
        probe('bundle', '${scriptUri}'),
        probe('scene file', window.BEADNETWORK_SCENE_BASE + '/view/speed.bin'),
        probe('anchor file', window.BEADNETWORK_ANCHOR_BASE + '/view/scene/selected.bin'),
      ]).then(function (lines) {
        box.textContent =
          'topology editor: the first bundle request did not settle. retrying.\\n\\n' +
          'script: ${scriptUri}\\n' +
          'readyState: ' + document.readyState + '\\n\\n' +
          lines.join('\\n');

        try {
          acquireVsCodeApi().postMessage({ type: "resources-dead" });
          box.textContent += '\\n\\ntold the host. Reopening does not help — this needs a quit.';
        } catch (e) {
          box.textContent += '\\n\\ncould not reach the host: ' + String(e);
        }
      });
    }, 5000);
  </script>
  <script nonce="${nonce}">${inlineBundle(scriptPath)}</script>
  <!-- Classic scripts run in order, so by the time this one starts the bundle above has
       either finished evaluating or thrown. Checking here rather than on a timer is what
       makes this honest: the old 3s deadline raced the bundle's own evaluation, and a
       machine busy with a verify run lost that race and cried wolf on a bundle that was
       about to boot. There is no duration to tune, because there is no waiting. -->
  <script nonce="${nonce}">
    window.BEADNETWORK_SETTLED = true;
    var marker = document.getElementById("html-marker");
    if (marker) marker.textContent = "inline scripts DO run. booted=" + !!window.BEADNETWORK_BOOTED;
    if (!window.BEADNETWORK_BOOTED) {
      var box = document.createElement("pre");
      box.setAttribute("style", "position:fixed;left:0;top:0;z-index:2147483647;margin:0;padding:16px;" +
        "font:12px ui-monospace,monospace;color:#ff6b6b;background:rgba(0,0,0,.9);white-space:pre-wrap");
      box.textContent =
        'topology editor: the webview bundle did not run.\\n\\n' +
        'script: ${scriptUri}\\n' +
        'bundle on disk: ${scriptExists ? "present" : "MISSING"}\\n\\n' +
        'The inline scripts ran, so scripts are allowed.\\n\\n' +
        (window.BEADNETWORK_FAULTS.length
          ? window.BEADNETWORK_FAULTS.join('\\n\\n')
          : 'No error event fired at all: the bundle request has not settled.');
      document.body.appendChild(box);
    }
  </script>
</body>
</html>`;
}

function inlineBundle(scriptPath: string): string {
  try {
    return fs.readFileSync(scriptPath, "utf8").replace(/<\/script>/gi, "<\\/script>");
  } catch (e) {
    return `document.title = "topology: bundle unreadable"; throw new Error(${JSON.stringify(String(e))});`;
  }
}

export function realPath(p: string): string {
  try {
    return fs.realpathSync(p);
  } catch {
    return p;
  }
}

function randomNonce(): string {
  return crypto.randomBytes(24).toString("base64url");
}

function mtimeMs(p: string): number {
  try {
    return Math.floor(fs.statSync(p).mtimeMs);
  } catch {
    return 0;
  }
}
