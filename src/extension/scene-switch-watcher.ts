import * as path from "path";
import * as vscode from "vscode";
import { buildWebviewHtml } from "./html";
import { resolveScenePath } from "./runner/counts";
import { SCENES } from "../Scene/scenes-gen";

export function sceneRoots(anchorPath: string): string[] {
  const container = path.dirname(anchorPath);
  return SCENES.map((s) => path.join(container, s.dir));
}

export function armSceneSwitchWatcher(
  panel: vscode.WebviewPanel,
  context: vscode.ExtensionContext,
  anchorPath: string,
  current: string,
): vscode.FileSystemWatcher {
  let scenePath = current;
  const watcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(anchorPath, "view", "scene")),
      "selected.bin",
    ),
  );
  const rebuild = () => {
    const next = resolveScenePath(anchorPath);
    if (next === scenePath) return;
    scenePath = next;
    console.log("[topology] scene switched, re-rendering webview for", next);
    panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath, next);
  };
  watcher.onDidChange(rebuild);
  watcher.onDidCreate(rebuild);
  return watcher;
}
