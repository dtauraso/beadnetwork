import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { buildWebviewHtml } from "./html";
import { webviewOptions } from "./webview-options";

export function armBundleWatcher(panel: vscode.WebviewPanel, context: vscode.ExtensionContext, scenePath: string, anchorPath: string): vscode.FileSystemWatcher {
  const bundleWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "webview.js",
    ),
  );
  console.log("[topology] bundleWatcher armed for", path.join(context.extensionPath, "out", "webview.js"));
  let pending: NodeJS.Timeout | undefined;
  const reload = (kind: string) => () => {
    console.log("[topology] bundleWatcher fired:", kind);
    if (pending) clearTimeout(pending);
    pending = setTimeout(() => {
      if (!vscode.workspace.getConfiguration("beadnetwork").get<boolean>("reloadOnBundleBuild", false)) {
        console.log("[topology] bundle changed; not re-rendering (beadnetwork.reloadOnBundleBuild is off)");
        return;
      }
      console.log("[topology] hot-reload: re-rendering webview.html");
      panel.webview.options = webviewOptions(context.extensionPath, anchorPath);
      const html = buildWebviewHtml(panel.webview, context.extensionPath, scenePath, anchorPath);
      panel.webview.html = html;
      try {
        const dir = path.join(path.dirname(anchorPath), ".probe");
        fs.mkdirSync(dir, { recursive: true });
        fs.writeFileSync(path.join(dir, "webview.html"), html, "utf8");
      } catch (err) {
        console.log("[topology] could not record webview.html:", err);
      }
    }, 150);
  };
  bundleWatcher.onDidChange(reload("change"));
  bundleWatcher.onDidCreate(reload("create"));
  return bundleWatcher;
}
