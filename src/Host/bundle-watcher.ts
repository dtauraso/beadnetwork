import * as path from "path";
import * as vscode from "vscode";
import { buildWebviewHtml } from "./html";

export function armBundleWatcher(panel: vscode.WebviewPanel, context: vscode.ExtensionContext): vscode.FileSystemWatcher {
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
      console.log("[topology] hot-reload: re-rendering webview.html");
      panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath);
    }, 150);
  };
  bundleWatcher.onDidChange(reload("change"));
  bundleWatcher.onDidCreate(reload("create"));
  return bundleWatcher;
}
