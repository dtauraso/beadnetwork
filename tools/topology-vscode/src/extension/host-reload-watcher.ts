import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { TrailingDebouncer } from "../hotRestart";
import { hashBundle, isHostReloadEnabled, shouldReloadHost } from "../hostReload";

export function armHostReloadWatcher(context: vscode.ExtensionContext): void {
  const hostBundlePath = path.join(context.extensionPath, "out", "extension.js");
  const hostChannel = vscode.window.createOutputChannel("topology host-reload");
  context.subscriptions.push(hostChannel);

  let loadedHash: string | undefined;
  try {
    loadedHash = hashBundle(fs.readFileSync(hostBundlePath));
  } catch { /* eslint-disable-line no-empty */ }
  const hostWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "extension.js",
    ),
  );
  context.subscriptions.push(hostWatcher);
  const debouncer = new TrailingDebouncer(250);
  context.subscriptions.push({ dispose: () => debouncer.dispose() });

  let reloading = false;
  const maybeReload = () => {
    debouncer.schedule(() => {
      if (reloading) return;

      if (!isHostReloadEnabled()) return;
      let newHash: string;
      try {
        newHash = hashBundle(fs.readFileSync(hostBundlePath));
      } catch {
        return;
      }
      if (!shouldReloadHost(loadedHash, newHash)) return;

      reloading = true;
      hostChannel.appendLine("[topology] extension host bundle changed — reloading window");
      vscode.window.setStatusBarMessage(
        "Topology: extension updated, reloading window…",
        3000,
      );

      void vscode.commands.executeCommand("workbench.action.reloadWindow");
    });
  };
  hostWatcher.onDidChange(maybeReload);
  hostWatcher.onDidCreate(maybeReload);
}
